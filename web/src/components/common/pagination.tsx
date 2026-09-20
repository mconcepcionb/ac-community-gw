import { Button } from "@/components/ui/button";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";

export interface PaginationRange {
  from: number;
  to: number;
  canPrevious: boolean;
  canNext: boolean;
}

interface RangeInput {
  limit: number;
  offset: number;
  count: number;
  total?: number;
}

/** usePagination derives the range label and navigation flags for a page. */
export function usePagination({ limit, offset, count, total }: RangeInput): PaginationRange {
  const canPrevious = offset > 0;
  const canNext = total !== undefined ? offset + count < total : count >= limit;
  return {
    from: count === 0 ? 0 : offset + 1,
    to: offset + count,
    canPrevious,
    canNext,
  };
}

interface PaginationProps {
  limit: number;
  offset: number;
  /** Number of items on the current page. */
  count: number;
  /** Total number of items, when the API reports it. */
  total?: number;
  onOffsetChange: (offset: number) => void;
  onLimitChange?: (limit: number) => void;
  pageSizeOptions?: number[];
  itemLabel?: string;
}

/** Pagination is the shared offset pagination strip for console lists. */
export function Pagination({
  limit,
  offset,
  count,
  total,
  onOffsetChange,
  onLimitChange,
  pageSizeOptions = [25, 50, 100],
  itemLabel = "items",
}: PaginationProps) {
  const { from, to, canPrevious, canNext } = usePagination({ limit, offset, count, total });

  return (
    <div className="mt-4 flex flex-wrap items-center justify-between gap-3 text-sm text-muted-foreground">
      <span>
        {count === 0
          ? `No ${itemLabel}`
          : total !== undefined
            ? `Showing ${from}–${to} of ${total}`
            : `Showing ${from}–${to}`}
      </span>
      <div className="flex items-center gap-2">
        {onLimitChange ? (
          <Select value={String(limit)} onValueChange={(value) => onLimitChange(Number(value))}>
            <SelectTrigger size="sm" aria-label="Page size">
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              {pageSizeOptions.map((option) => (
                <SelectItem key={option} value={String(option)}>
                  {option} per page
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
        ) : null}
        <Button
          variant="outline"
          size="sm"
          disabled={!canPrevious}
          onClick={() => onOffsetChange(Math.max(0, offset - limit))}
        >
          Previous
        </Button>
        <Button
          variant="outline"
          size="sm"
          disabled={!canNext}
          onClick={() => onOffsetChange(offset + limit)}
        >
          Next
        </Button>
      </div>
    </div>
  );
}
