import { describe, expect, it } from "vitest";

import { cn } from "./utils";

describe("cn", () => {
  it("merges conflicting tailwind classes, keeping the last", () => {
    expect(cn("p-2", "p-4")).toBe("p-4");
  });

  it("ignores falsy values and joins the rest", () => {
    expect(cn("flex", false && "hidden", undefined, "items-center")).toBe("flex items-center");
  });
});
