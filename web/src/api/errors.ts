import type { ErrorResponse } from "./generated/types.gen";

/**
 * ApiError is the normalised error thrown by the generated client for any
 * non-2xx response. It mirrors the gateway error envelope.
 */
export class ApiError extends Error {
  readonly status: number;
  readonly code: string;
  readonly details: unknown;
  readonly requestId?: string;

  constructor(params: {
    status: number;
    code: string;
    message: string;
    details?: unknown;
    requestId?: string;
  }) {
    super(params.message);
    this.name = "ApiError";
    this.status = params.status;
    this.code = params.code;
    this.details = params.details;
    this.requestId = params.requestId;
  }
}

export function isApiError(error: unknown): error is ApiError {
  return error instanceof ApiError;
}

export function isUnauthorized(error: unknown): boolean {
  return isApiError(error) && error.status === 401;
}

function isErrorEnvelope(value: unknown): value is ErrorResponse {
  return typeof value === "object" && value !== null && "error" in value;
}

/**
 * toApiError normalises whatever the client produced (parsed JSON envelope,
 * string body, network failure or an already-normalised ApiError) into an
 * ApiError.
 */
export function toApiError(error: unknown, status?: number): ApiError {
  if (error instanceof ApiError) {
    return error;
  }

  if (isErrorEnvelope(error)) {
    const body = error.error;
    return new ApiError({
      status: status ?? 0,
      code: body?.code ?? "unknown_error",
      message: body?.message ?? "Unexpected error",
      details: body?.details,
      requestId: error.request_id,
    });
  }

  if (error && typeof error === "object") {
    const record = error as Record<string, unknown>;
    const message =
      typeof record.message === "string"
        ? record.message
        : typeof record.error === "string"
          ? record.error
          : safeStringify(record);
    return new ApiError({
      status: status ?? 0,
      code: typeof record.code === "string" ? record.code : "unknown_error",
      message,
    });
  }

  return new ApiError({
    status: status ?? 0,
    code: "unknown_error",
    message: typeof error === "string" ? error : "Unexpected error",
  });
}

/** safeStringify renders an object readably without throwing or overflowing. */
function safeStringify(value: unknown): string {
  try {
    const json = JSON.stringify(value);
    if (!json) {
      return "Unexpected error";
    }
    return json.length > 300 ? `${json.slice(0, 300)}…` : json;
  } catch {
    return "Unexpected error";
  }
}
