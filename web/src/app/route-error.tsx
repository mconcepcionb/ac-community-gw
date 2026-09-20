import { isApiError } from "@/api/errors";

interface RouteErrorProps {
  error: unknown;
  reset: () => void;
}

/**
 * RouteError is the router-level error boundary. It surfaces the typed
 * ApiError code when available instead of a blank page.
 */
export function RouteError({ error, reset }: RouteErrorProps) {
  const apiError = isApiError(error) ? error : undefined;
  const message = apiError
    ? `${apiError.message} (${apiError.code})`
    : error instanceof Error
      ? error.message
      : String(error);

  return (
    <div role="alert" className="mx-auto max-w-3xl p-8">
      <h1 className="text-xl font-semibold">Something went wrong</h1>
      <p className="mt-2 text-sm text-neutral-400">{message}</p>
      <button
        type="button"
        className="mt-4 rounded border border-neutral-700 px-3 py-1.5 text-sm hover:bg-neutral-800"
        onClick={reset}
      >
        Retry
      </button>
    </div>
  );
}
