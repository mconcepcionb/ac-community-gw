import { isApiError } from "@/api/errors";
import { Button } from "@/components/ui/button";

interface ErrorStateProps {
  error: unknown;
  onRetry?: () => void;
}

export function ErrorState({ error, onRetry }: ErrorStateProps) {
  const message = isApiError(error)
    ? `${error.message} (${error.code})`
    : error instanceof Error
      ? error.message
      : String(error);

  return (
    <div role="alert" className="rounded-md border border-destructive/40 p-6 text-center">
      <p className="text-sm font-medium text-destructive">Something went wrong</p>
      <p className="mt-1 text-sm text-muted-foreground">{message}</p>
      {onRetry ? (
        <Button variant="outline" size="sm" className="mt-4" onClick={onRetry}>
          Retry
        </Button>
      ) : null}
    </div>
  );
}
