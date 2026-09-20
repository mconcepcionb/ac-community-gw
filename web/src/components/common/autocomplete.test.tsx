import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { useState } from "react";
import { describe, expect, it, vi } from "vitest";

import { Autocomplete, type AutocompleteOption } from "./autocomplete";

const options: AutocompleteOption[] = [
  { value: "Thrall", label: "Thrall", description: "80 Shaman" },
  { value: "Jaina", label: "Jaina", description: "80 Mage" },
];

function Harness({
  loadOptions,
  onSelect,
}: {
  loadOptions: (query: string) => Promise<AutocompleteOption[]>;
  onSelect?: (option: AutocompleteOption) => void;
}) {
  const [value, setValue] = useState("");
  return (
    <Autocomplete
      value={value}
      onValueChange={setValue}
      loadOptions={loadOptions}
      onSelect={onSelect}
      ariaLabel="Character"
      placeholder="Character"
    />
  );
}

describe("Autocomplete", () => {
  it("loads and renders options as the user types", async () => {
    const loadOptions = vi.fn(async () => options);
    render(<Harness loadOptions={loadOptions} />);

    const user = userEvent.setup();
    await user.type(screen.getByLabelText("Character"), "Thr");

    expect(await screen.findByRole("option", { name: /Thrall/ })).toBeInTheDocument();
    expect(loadOptions).toHaveBeenCalled();
  });

  it("selects an option on click and notifies onSelect", async () => {
    const onSelect = vi.fn();
    render(<Harness loadOptions={async () => options} onSelect={onSelect} />);

    const user = userEvent.setup();
    await user.type(screen.getByLabelText("Character"), "Ja");
    await user.click(await screen.findByRole("option", { name: /Jaina/ }));

    expect(onSelect).toHaveBeenCalledWith(expect.objectContaining({ value: "Jaina" }));
    expect(screen.queryByRole("listbox")).not.toBeInTheDocument();
  });

  it("selects the highlighted option with the keyboard", async () => {
    const onSelect = vi.fn();
    render(<Harness loadOptions={async () => options} onSelect={onSelect} />);

    const user = userEvent.setup();
    const input = screen.getByLabelText("Character");
    await user.type(input, "T");
    await screen.findByRole("option", { name: /Thrall/ });
    await user.keyboard("{ArrowDown}{ArrowUp}{Enter}");

    expect(onSelect).toHaveBeenCalledWith(expect.objectContaining({ value: "Thrall" }));
  });

  it("shows the empty state when nothing matches", async () => {
    render(<Harness loadOptions={async () => []} />);

    const user = userEvent.setup();
    await user.type(screen.getByLabelText("Character"), "zzz");

    await waitFor(() => expect(screen.getByText("No matches")).toBeInTheDocument());
  });
});
