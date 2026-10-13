export function LoadingState({ count = 4 }: { count?: number }) {
  return (
    <div className="destination-grid" aria-label="Loading places" aria-busy="true">
      {Array.from({ length: count }, (_, index) => (
        <div className="place-card" key={index}>
          <div className="card-image skeleton" />
          <div className="card-body">
            <div className="skeleton-line" />
            <div className="skeleton-line short" />
          </div>
        </div>
      ))}
    </div>
  );
}
