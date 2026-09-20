import { describe, expect, it } from "vitest";

import { buildLoginUrl, isSafeReturnTo } from "./actions";

describe("auth actions", () => {
  it("builds a login URL with an encoded safe return_to", () => {
    expect(buildLoginUrl("/dashboard?tab=1")).toBe(
      "/api/v1/auth/discord/login?return_to=%2Fdashboard%3Ftab%3D1",
    );
  });

  it("rejects unsafe return_to values", () => {
    expect(isSafeReturnTo("https://evil.test")).toBe(false);
    expect(isSafeReturnTo("//evil.test")).toBe(false);
    expect(isSafeReturnTo("/\\evil.test")).toBe(false);
    expect(isSafeReturnTo("/%5Cevil.test")).toBe(false);
    expect(isSafeReturnTo("/ok")).toBe(true);

    expect(buildLoginUrl("https://evil.test")).not.toContain("evil.test");
  });
});
