"use client";

import { type KeyboardEvent, useEffect, useId, useRef, useState } from "react";

import { Input } from "@/components/ui/input";
import { cn } from "@/lib/utils";

export interface AutocompleteOption {
  value: string;
  label: string;
  description?: string;
}

interface AutocompleteProps {
  value: string;
  onValueChange: (value: string) => void;
  /** loadOptions resolves the suggestions for the current query. Memoize it. */
  loadOptions: (query: string) => Promise<AutocompleteOption[]>;
  onSelect?: (option: AutocompleteOption) => void;
  placeholder?: string;
  ariaLabel?: string;
  id?: string;
  disabled?: boolean;
  minChars?: number;
  debounceMs?: number;
  emptyMessage?: string;
  className?: string;
}

/**
 * Autocomplete is a combobox that resolves free-form input through an async
 * option loader. Typing is always allowed; selecting an option calls onSelect.
 */
export function Autocomplete({
  value,
  onValueChange,
  loadOptions,
  onSelect,
  placeholder,
  ariaLabel,
  id,
  disabled,
  minChars = 1,
  debounceMs = 250,
  emptyMessage = "No matches",
  className,
}: AutocompleteProps) {
  const [options, setOptions] = useState<AutocompleteOption[]>([]);
  const [open, setOpen] = useState(false);
  const [loading, setLoading] = useState(false);
  const [highlighted, setHighlighted] = useState(0);

  const listId = useId();
  const selectedRef = useRef<string | null>(null);
  const loadRef = useRef(loadOptions);
  loadRef.current = loadOptions;

  useEffect(() => {
    if (selectedRef.current !== null && selectedRef.current === value) {
      selectedRef.current = null;
      return;
    }
    const query = value.trim();
    if (query.length < minChars) {
      setOptions([]);
      setOpen(false);
      setLoading(false);
      return;
    }

    let cancelled = false;
    setLoading(true);
    const handle = window.setTimeout(async () => {
      try {
        const result = await loadRef.current(query);
        if (!cancelled) {
          setOptions(result);
          setHighlighted(0);
          setOpen(true);
        }
      } catch {
        if (!cancelled) {
          setOptions([]);
        }
      } finally {
        if (!cancelled) {
          setLoading(false);
        }
      }
    }, debounceMs);

    return () => {
      cancelled = true;
      window.clearTimeout(handle);
    };
  }, [value, minChars, debounceMs]);

  const select = (option: AutocompleteOption) => {
    selectedRef.current = option.value;
    onValueChange(option.value);
    onSelect?.(option);
    setOpen(false);
  };

  const onKeyDown = (event: KeyboardEvent<HTMLInputElement>) => {
    if (event.key === "ArrowDown") {
      event.preventDefault();
      setOpen(true);
      setHighlighted((current) => Math.min(current + 1, options.length - 1));
      return;
    }
    if (event.key === "ArrowUp") {
      event.preventDefault();
      setHighlighted((current) => Math.max(current - 1, 0));
      return;
    }
    if (event.key === "Enter") {
      if (open && options[highlighted]) {
        event.preventDefault();
        select(options[highlighted]);
      }
      return;
    }
    if (event.key === "Escape") {
      setOpen(false);
    }
  };

  const showDropdown = open && value.trim().length >= minChars;

  return (
    <div className={cn("relative", className)}>
      <Input
        id={id}
        role="combobox"
        aria-label={ariaLabel}
        aria-expanded={showDropdown}
        aria-controls={listId}
        aria-autocomplete="list"
        aria-activedescendant={
          showDropdown && options[highlighted] ? `${listId}-${highlighted}` : undefined
        }
        autoComplete="off"
        disabled={disabled}
        placeholder={placeholder}
        value={value}
        onChange={(event) => {
          onValueChange(event.target.value);
          setOpen(true);
        }}
        onKeyDown={onKeyDown}
        onBlur={() => setOpen(false)}
      />
      {showDropdown ? (
        <div
          id={listId}
          role="listbox"
          aria-label={ariaLabel ? `${ariaLabel} suggestions` : "Suggestions"}
          className="absolute z-50 mt-1 max-h-60 w-full overflow-auto rounded-md border border-border bg-popover p-1 text-popover-foreground shadow-md"
        >
          {loading ? (
            <div className="px-2 py-1.5 text-sm text-muted-foreground">Loading…</div>
          ) : options.length === 0 ? (
            <div className="px-2 py-1.5 text-sm text-muted-foreground">{emptyMessage}</div>
          ) : (
            options.map((option, index) => (
              <div
                key={option.value}
                id={`${listId}-${index}`}
                role="option"
                tabIndex={-1}
                aria-selected={index === highlighted}
                className={cn(
                  "cursor-default rounded-sm px-2 py-1.5 text-sm",
                  index === highlighted ? "bg-accent text-accent-foreground" : undefined,
                )}
                onMouseDown={(event) => {
                  event.preventDefault();
                  select(option);
                }}
                onMouseEnter={() => setHighlighted(index)}
              >
                <span className="font-medium">{option.label}</span>
                {option.description ? (
                  <span className="ml-2 text-xs text-muted-foreground">{option.description}</span>
                ) : null}
              </div>
            ))
          )}
        </div>
      ) : null}
    </div>
  );
}
