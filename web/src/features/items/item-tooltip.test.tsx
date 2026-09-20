import { render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";

import { WoWItemTooltip } from "./item-tooltip";

describe("WoWItemTooltip", () => {
  it("renders the item name, stats, damage and description", () => {
    render(
      <WoWItemTooltip
        item={{
          name: "Thunderfury, Blessed Blade of the Windseeker",
          quality_color: "#ff8000",
          item_level: 80,
          required_level: 60,
          inventory_type_name: "One-Hand",
          stats: [{ type: 4, name: "Strength", value: 20 }],
          damage: [{ type: 0, type_name: "Physical", min: 44, max: 115, dps: 50.3 }],
          spells: [{ id: 21992, trigger_name: "Chance on hit" }],
          resistances: { nature: 8 },
          description: "A legendary blade.",
        }}
      />,
    );

    expect(screen.getByText("Thunderfury, Blessed Blade of the Windseeker")).toBeInTheDocument();
    expect(screen.getByText("+20 Strength")).toBeInTheDocument();
    expect(screen.getByText(/44–115 Physical/)).toBeInTheDocument();
    expect(screen.getByText("+8 Nature Resistance")).toBeInTheDocument();
    expect(screen.getByText("A legendary blade.")).toBeInTheDocument();
  });
});
