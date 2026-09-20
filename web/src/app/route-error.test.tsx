import { render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";

import { ApiError } from "@/api/errors";
import { RouteError } from "./route-error";

describe("RouteError", () => {
  it("surfaces the ApiError message and code", () => {
    render(
      <RouteError
        error={new ApiError({ status: 500, code: "internal_error", message: "boom" })}
        reset={vi.fn()}
      />,
    );

    expect(screen.getByText(/boom/)).toBeInTheDocument();
    expect(screen.getByText(/internal_error/)).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Retry" })).toBeInTheDocument();
  });

  it("falls back to the plain error message", () => {
    render(<RouteError error={new Error("plain failure")} reset={vi.fn()} />);
    expect(screen.getByText(/plain failure/)).toBeInTheDocument();
  });
});
