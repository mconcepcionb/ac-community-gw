import { zodResolver } from "@hookform/resolvers/zod";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { useForm } from "react-hook-form";
import { describe, expect, it } from "vitest";
import { z } from "zod";

import { Form } from "@/components/ui/form";
import { TextField } from "./form-controls";

const schema = z.object({ username: z.string().min(1, "Required") });

function TestForm() {
  const form = useForm({
    resolver: zodResolver(schema),
    defaultValues: { username: "" },
  });

  return (
    <Form {...form}>
      <form onSubmit={form.handleSubmit(() => {})}>
        <TextField control={form.control} name="username" label="Username" />
        <button type="submit">Submit</button>
      </form>
    </Form>
  );
}

describe("TextField", () => {
  it("renders the label and surfaces validation errors", async () => {
    render(<TestForm />);

    expect(screen.getByLabelText("Username")).toBeInTheDocument();

    await userEvent.click(screen.getByRole("button", { name: "Submit" }));
    expect(await screen.findByText("Required")).toBeInTheDocument();
  });
});
