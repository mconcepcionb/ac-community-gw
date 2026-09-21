import { renderHook, waitFor } from "@testing-library/react";
import { describe, expect, it } from "vitest";

import { useDebouncedValue } from "./use-debounced-value";

describe("useDebouncedValue", () => {
  it("returns the initial value and only updates after the delay", async () => {
    const { result, rerender } = renderHook(
      ({ value }: { value: string }) => useDebouncedValue(value, 10),
      {
        initialProps: { value: "a" },
      },
    );
    expect(result.current).toBe("a");

    rerender({ value: "b" });
    expect(result.current).toBe("a");

    await waitFor(() => expect(result.current).toBe("b"));
  });
});
