import { render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";

import { StatusBadge } from "./status-badge";

describe("StatusBadge", () => {
  it("renders its text", () => {
    render(<StatusBadge tone="positive">Online</StatusBadge>);
    expect(screen.getByText("Online")).toBeInTheDocument();
  });

  it("applies the tone class", () => {
    render(<StatusBadge tone="negative">Banned</StatusBadge>);
    expect(screen.getByText("Banned").className).toContain("text-destructive");
  });
});
