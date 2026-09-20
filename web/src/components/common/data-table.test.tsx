import type { ColumnDef } from "@tanstack/react-table";
import { render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";

import { DataTable } from "./data-table";

interface Row {
  name: string;
}

const columns: ColumnDef<Row, unknown>[] = [{ accessorKey: "name", header: "Name" }];

describe("DataTable", () => {
  it("renders rows", () => {
    render(<DataTable columns={columns} data={[{ name: "Thrall" }]} />);
    expect(screen.getByText("Thrall")).toBeInTheDocument();
  });

  it("renders the empty state", () => {
    render(<DataTable columns={columns} data={[]} emptyMessage="No rows" />);
    expect(screen.getByText("No rows")).toBeInTheDocument();
  });

  it("renders the loading state", () => {
    render(<DataTable columns={columns} data={[]} loading />);
    expect(screen.getAllByText("Loading…").length).toBeGreaterThan(0);
  });

  it("renders the error state", () => {
    render(<DataTable columns={columns} data={[]} error={new Error("boom")} />);
    expect(screen.getByRole("alert")).toBeInTheDocument();
    expect(screen.getByText(/boom/)).toBeInTheDocument();
  });
});
