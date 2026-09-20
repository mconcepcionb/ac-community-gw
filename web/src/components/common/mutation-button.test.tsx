import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";

import { MutationButton } from "./mutation-button";

describe("MutationButton", () => {
  it("calls the action on click", async () => {
    const onAction = vi.fn().mockResolvedValue(undefined);
    render(<MutationButton onAction={onAction}>Save</MutationButton>);

    await userEvent.click(screen.getByRole("button", { name: "Save" }));
    expect(onAction).toHaveBeenCalledOnce();
  });

  it("shows the pending state and disables the button", () => {
    render(
      <MutationButton onAction={vi.fn()} pending>
        Save
      </MutationButton>,
    );

    expect(screen.getByRole("button")).toBeDisabled();
    expect(screen.getByText("Working…")).toBeInTheDocument();
  });
});
