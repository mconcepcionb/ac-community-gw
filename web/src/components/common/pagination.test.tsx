import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";

import { Pagination } from "./pagination";

describe("Pagination", () => {
  it("shows the range and disables Previous on the first page", () => {
    render(
      <Pagination limit={50} offset={0} count={10} onOffsetChange={() => {}} itemLabel="items" />,
    );
    expect(screen.getByText("Showing 1–10")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Previous" })).toBeDisabled();
  });

  it("disables Next on a partial last page", () => {
    render(<Pagination limit={50} offset={0} count={10} onOffsetChange={() => {}} />);
    expect(screen.getByRole("button", { name: "Next" })).toBeDisabled();
  });

  it("enables Next on a full page and reports the offset", async () => {
    const onOffsetChange = vi.fn();
    render(<Pagination limit={50} offset={0} count={50} onOffsetChange={onOffsetChange} />);
    const user = userEvent.setup();
    await user.click(screen.getByRole("button", { name: "Next" }));
    expect(onOffsetChange).toHaveBeenCalledWith(50);
  });

  it("shows the total when known", () => {
    render(<Pagination limit={50} offset={0} count={50} total={120} onOffsetChange={() => {}} />);
    expect(screen.getByText("Showing 1–50 of 120")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Next" })).toBeEnabled();
  });

  it("shows an empty label when there are no items", () => {
    render(
      <Pagination limit={50} offset={0} count={0} onOffsetChange={() => {}} itemLabel="users" />,
    );
    expect(screen.getByText("No users")).toBeInTheDocument();
  });
});
