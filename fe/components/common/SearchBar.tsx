import { Search } from "lucide-react";

type SearchBarProps = {
  placeholder?: string;
  value?: string;
  onChange?: (value: string) => void;
};

export function SearchBar({
  placeholder = "Search destinations, food, stays...",
  value = "",
  onChange,
}: SearchBarProps) {
  return (
    <label className="search-box">
      <Search size={19} />
      <input
        value={value}
        onChange={(event) => onChange?.(event.target.value)}
        placeholder={placeholder}
        aria-label={placeholder}
      />
      {!value && <kbd>⌘ K</kbd>}
    </label>
  );
}
