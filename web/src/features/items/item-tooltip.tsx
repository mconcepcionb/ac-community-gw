import type { ReactNode } from "react";

import type { AzerothItem } from "@/api";
import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from "@/components/ui/tooltip";

const resistanceLabels: Record<string, string> = {
  holy: "Holy",
  fire: "Fire",
  nature: "Nature",
  frost: "Frost",
  shadow: "Shadow",
  arcane: "Arcane",
};

/** WoWItemTooltip renders an in-game-style tooltip from the item data. */
export function WoWItemTooltip({ item }: { item: AzerothItem }) {
  const stats = item.stats ?? [];
  const damage = item.damage ?? [];
  const spells = item.spells ?? [];
  const resistances = Object.entries(item.resistances ?? {}).filter(([, value]) => Boolean(value));

  return (
    <div className="w-72 space-y-1 text-left">
      <p className="text-sm font-semibold" style={{ color: item.quality_color }}>
        {item.name}
      </p>
      {item.bonding_name ? <p className="text-xs">{item.bonding_name}</p> : null}
      {item.inventory_type_name ? <p className="text-xs">{item.inventory_type_name}</p> : null}
      {item.item_level ? (
        <p className="text-xs">
          Item Level <span className="font-medium">{item.item_level}</span>
        </p>
      ) : null}
      {item.required_level ? <p className="text-xs">Requires Level {item.required_level}</p> : null}
      {damage.map((entry) => (
        <p key={`${entry.type}-${entry.min}`} className="text-xs">
          {entry.min}–{entry.max} {entry.type_name} ({(entry.dps ?? 0).toFixed(1)} dps)
        </p>
      ))}
      {item.armor ? <p className="text-xs">{item.armor} Armor</p> : null}
      {stats.map((stat) => (
        <p key={`${stat.type}-${stat.name}`} className="text-xs">
          +{stat.value} {stat.name}
        </p>
      ))}
      {resistances.map(([key, value]) => (
        <p key={key} className="text-xs">
          +{value} {resistanceLabels[key] ?? key} Resistance
        </p>
      ))}
      {spells.map((spell) => (
        <p key={spell.id} className="text-xs">
          {spell.trigger_name}: spell {spell.id}
        </p>
      ))}
      {item.description ? <p className="text-xs italic">{item.description}</p> : null}
      {item.sell_price ? <p className="text-xs">Sell: {item.sell_price}</p> : null}
    </div>
  );
}

/** ItemNameWithTooltip shows an item name with a hover tooltip. */
export function ItemNameWithTooltip({
  item,
  children,
}: {
  item: AzerothItem;
  children: ReactNode;
}) {
  return (
    <TooltipProvider>
      <Tooltip>
        <TooltipTrigger asChild>{children}</TooltipTrigger>
        <TooltipContent
          side="right"
          className="border border-border bg-popover p-3 text-popover-foreground"
        >
          <WoWItemTooltip item={item} />
        </TooltipContent>
      </Tooltip>
    </TooltipProvider>
  );
}
