package router

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"halaltrip/be/internal/catalog"
	"halaltrip/be/internal/config"
	"halaltrip/be/internal/directory"
	"halaltrip/be/internal/health"
	"halaltrip/be/internal/middleware"
	"halaltrip/be/internal/recommendation"
	"halaltrip/be/internal/travelprovider"
	"halaltrip/be/internal/trips"
	"halaltrip/be/internal/users"
)

func New(cfg config.Config, pool *pgxpool.Pool, logger *slog.Logger) *gin.Engine {
	if cfg.AppEnv == "production" {
		gin.SetMode(gin.ReleaseMode)
	}
	r := gin.New()
	r.Use(middleware.CORS(cfg.CORSAllowedOrigins), middleware.RequestLogger(logger), middleware.Recovery(logger))
	check := health.NewHandler(pool)
	userStore := users.NewPostgresStore(pool)
	userHandler := users.NewHandler(userStore, []byte(cfg.JWTSecret), cfg.JWTIssuer, cfg.JWTAudience, int64(cfg.JWTAccessTokenTTL.Seconds()))
	provider := recommendation.NewHandler(travelprovider.NewClient(cfg.TravelRecommendationAPIURL, cfg.HTTPClientTimeout), userStore)
	directoryStore := directory.NewPostgresStore(pool)
	directoryHandler := directory.NewHandler(directoryStore)
	catalogHandler := catalog.NewHandler(pool)
	tripHandler := trips.NewHandler(trips.NewPostgresStore(pool))
	jwtRequired := middleware.JWT([]byte(cfg.JWTSecret), cfg.JWTIssuer, cfg.JWTAudience, false)
	jwtOptional := middleware.JWT([]byte(cfg.JWTSecret), cfg.JWTIssuer, cfg.JWTAudience, true)
	r.GET("/health", check.Live)
	r.GET("/health/ready", check.Ready)
	v1 := r.Group("/api/v1")
	v1.GET("/health/recommendation-provider", provider.ProviderHealth)
	v1.POST("/auth/register", userHandler.Register)
	v1.POST("/auth/login", userHandler.Login)
	protected := v1.Group("/users/me", jwtRequired)
	protected.GET("", userHandler.Me)
	protected.PATCH("", userHandler.PatchMe)
	protected.GET("/preferences", userHandler.GetPreferences)
	protected.PATCH("/preferences", userHandler.SavePreferences)
	tripRoutes := v1.Group("/trips", jwtRequired)
	tripRoutes.POST("", tripHandler.CreateTrip)
	tripRoutes.GET("", tripHandler.ListTrips)
	tripRoutes.GET("/:tripId", tripHandler.GetTrip)
	tripRoutes.PATCH("/:tripId", tripHandler.PatchTrip)
	tripRoutes.DELETE("/:tripId", tripHandler.DeleteTrip)
	tripRoutes.POST("/:tripId/itinerary-items", tripHandler.CreateItem)
	tripRoutes.GET("/:tripId/itinerary-items", tripHandler.ListItems)
	tripRoutes.PUT("/:tripId/itinerary-items/order", tripHandler.ReorderItems)
	tripRoutes.PATCH("/:tripId/itinerary-items/:itemId", tripHandler.PatchItem)
	tripRoutes.DELETE("/:tripId/itinerary-items/:itemId", tripHandler.DeleteItem)
	saved := v1.Group("", jwtRequired)
	saved.GET("/saved-places", directoryHandler.ListSaved)
	saved.POST("/saved-places", directoryHandler.Save)
	saved.DELETE("/saved-places/:place_type/:place_id", directoryHandler.Unsave)
	admin := v1.Group("", jwtRequired, middleware.Admin(pool))
	for _, resource := range []struct {
		path string
		kind directory.Kind
	}{{"restaurants", directory.Restaurant}, {"mosques", directory.Mosque}, {"hotels", directory.Hotel}} {
		admin.POST("/"+resource.path, directoryHandler.Create(resource.kind))
		admin.PUT("/"+resource.path+"/:id", directoryHandler.Update(resource.kind))
		admin.DELETE("/"+resource.path+"/:id", directoryHandler.Delete(resource.kind))
	}
	for _, resource := range []struct {
		path string
		kind directory.Kind
	}{{"restaurants", directory.Restaurant}, {"mosques", directory.Mosque}, {"hotels", directory.Hotel}} {
		v1.GET("/"+resource.path, directoryHandler.List(resource.kind))
		v1.GET("/"+resource.path+"/:id", directoryHandler.Get(resource.kind))
	}
	recommendations := v1.Group("", jwtOptional)
	recommendations.POST("/recommendations/cities", provider.Cities)
	recommendations.POST("/recommendations/destinations", provider.DestinationsRecommendation)
	v1.GET("/destinations", provider.ListDestinations)
	v1.GET("/destinations/:id", provider.GetDestination)
	v1.GET("/catalog/destinations", catalogHandler.ListDestinations)
	v1.GET("/catalog/destinations/:id", catalogHandler.GetDestination)
	v1.POST("/destinations", provider.CreateDestination)
	v1.PUT("/destinations/:id", provider.UpdateDestination)
	v1.DELETE("/destinations/:id", provider.DeleteDestination)
	r.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusNotFound, gin.H{"error": gin.H{
			"code": "not_found", "message": "Endpoint tidak ditemukan.",
		}})
	})
	return r
}
