import { describe, expect, it } from "vitest";

import { ApiError, isApiError, isUnauthorized, toApiError } from "./errors";

describe("toApiError", () => {
  it("returns an existing ApiError unchanged", () => {
    const original = new ApiError({ status: 404, code: "not_found", message: "nope" });
    expect(toApiError(original)).toBe(original);
  });

  it("normalises a gateway error envelope", () => {
    const error = toApiError(
      { error: { code: "invalid_report", message: "bad" }, request_id: "req-1" },
      422,
    );
    expect(error).toBeInstanceOf(ApiError);
    expect(error.status).toBe(422);
    expect(error.code).toBe("invalid_report");
    expect(error.message).toBe("bad");
    expect(error.requestId).toBe("req-1");
  });

  it("falls back for a partial envelope", () => {
    const error = toApiError({ error: {} });
    expect(error.code).toBe("unknown_error");
    expect(error.message).toBe("Unexpected error");
  });

  it("uses an object message and code when present", () => {
    const error = toApiError({ message: "boom", code: "x" }, 500);
    expect(error.message).toBe("boom");
    expect(error.code).toBe("x");
    expect(error.status).toBe(500);
  });

  it("stringifies an object without a message", () => {
    expect(toApiError({ a: 1 }).message).toBe('{"a":1}');
  });

  it("handles primitives, undefined and circular objects", () => {
    expect(toApiError("network down").message).toBe("network down");
    expect(toApiError(undefined).message).toBe("Unexpected error");
    const circular: Record<string, unknown> = {};
    circular.self = circular;
    expect(toApiError(circular).message).toBe("Unexpected error");
  });

  it("truncates a long serialized value", () => {
    const message = toApiError({ data: "x".repeat(400) }).message;
    expect(message.length).toBe(301);
    expect(message.startsWith('{"data":"')).toBe(true);
    expect(message.endsWith("…")).toBe(true);
  });
});

describe("api error guards", () => {
  it("identifies ApiError and unauthorized responses", () => {
    const unauthorized = new ApiError({ status: 401, code: "unauthorized", message: "no" });
    expect(isApiError(unauthorized)).toBe(true);
    expect(isApiError(new Error("x"))).toBe(false);
    expect(isUnauthorized(unauthorized)).toBe(true);
    expect(isUnauthorized(new ApiError({ status: 500, code: "x", message: "x" }))).toBe(false);
  });
});
