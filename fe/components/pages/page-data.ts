import type { Place } from "@/src/types/place";
import { getDestinations, getHotels, getMosques, getRestaurants } from "@/src/services";

export const placeServices: Record<string, () => Promise<Place[]>> = {
  destinations: getDestinations,
  restaurants: getRestaurants,
  hotels: getHotels,
  mosques: getMosques,
};

export const LISTING_TITLES: Record<string, string> = {
  destinations: "Find your next somewhere",
  restaurants: "Good food, wherever you are",
  hotels: "Stay somewhere considered",
  mosques: "A place to pause",
};

export const LISTING_DESCRIPTIONS: Record<string, string> = {
  destinations: "Explore Muslim-friendly places with the details that make travel feel easier.",
  restaurants: "Discover trusted halal meals and local favorites in every city.",
  hotels: "Find a comfortable stay with thoughtful prayer-friendly amenities.",
  mosques: "Locate welcoming prayer spaces wherever your journey takes you.",
};

export const PLANNER_STEPS = ["Destination", "Dates", "Travelers", "Priorities", "Budget", "Getting around"] as const;
export const PLANNER_QUESTIONS = [
  { title: "Where are you going?", description: "Add a destination to your trip draft." },
  { title: "When are you traveling?", description: "Choose your travel dates." },
  { title: "Who’s coming along?", description: "Choose the travel group." },
  { title: "What matters on this trip?", description: "Choose the details to keep in mind." },
  { title: "What feels right for your budget?", description: "Choose a budget range." },
  { title: "How would you like to get around?", description: "Select transport preferences." },
];

export const TRAVEL_STYLES = ["Solo", "Couple", "Family", "Friends"] as const;
export const PREFERENCE_OPTIONS = ["Halal food", "Mosques", "Shopping", "Nature", "History", "Nightlife"] as const;
export const BUDGET_OPTIONS = ["Budget", "Moderate", "Premium"] as const;
export const TRANSPORT_MODES = ["Public Transport", "Taxi", "Rental Car"] as const;
export const SAVED_TABS = ["All", "Destinations", "Restaurants", "Hotels", "Mosques"] as const;
export const PROFILE_PREFERENCES = [
  { title: "Halal food", description: "Show me trusted places to eat" },
  { title: "Prayer facilities", description: "Help me find room to pause" },
  { title: "Muslim-friendly stays", description: "Thoughtful amenities matter" },
];
