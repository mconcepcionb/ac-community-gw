import { describe, expect, it } from "vitest";

import { ApiError } from "@/api/errors";
import { createQueryClient } from "./query-client";

describe("query client defaults", () => {
  it("does not retry client errors but retries transient failures twice", () => {
    const retry = createQueryClient().getDefaultOptions().queries?.retry;
    if (typeof retry !== "function") {
      throw new Error("expected a retry function");
    }

    const unauthorized = new ApiError({ status: 401, code: "unauthorized", message: "nope" });
    expect(retry(0, unauthorized)).toBe(false);
    expect(retry(0, new Error("network"))).toBe(true);
    expect(retry(1, new Error("network"))).toBe(true);
    expect(retry(2, new Error("network"))).toBe(false);
  });

  it("disables mutation retries", () => {
    expect(createQueryClient().getDefaultOptions().mutations?.retry).toBe(false);
  });
});
