# Item tooltips and icons

## Goal

Render items with authentic, in-game-style tooltips (and icons where a reliable
source exists), instead of a plain field grid.

## Context

- `itemview.View` already returns quality name/colour, stats, damage/DPS, spells,
  resistances, bonding, armor, prices and `display_id`
  (`internal/core/itemview/itemview.go:46-76`) — but no tooltip and no icon.
- The backend serves no images/media; the only icon-adjacent value is
  `display_id` (`internal/adapters/azerothmysql/items.go:24`). Mapping
  `display_id → icon` lives in client-side DBC data (`ItemDisplayInfo.dbc`), not
  the world DB.
- Wowhead's tooltip widget (`https://wow.zamimg.com/widgets/power.js`) renders
  from Wowhead's data and requires external script/network plus CSP relaxation
  (`web/Caddyfile:46` currently allows only `self`, `data:` and the Discord CDN).
  Private-server/custom items will not exist there.
- The detail page is a field grid and does not colour the item name by quality
  (`web/src/features/items/item-detail-page.tsx:47-65`).

## Investigation (do before deciding)

1. Determine an icon source: a bundled `display_id → icon` map exported from the
   client DBCs, a prebuilt community dataset, or a per-realm mapping table.
2. Decide whether icons are self-hosted (requires a media route + CSP `img-src`)
   or omitted (fallback: no icon, tooltip still renders).
3. Confirm WoW client version/expansion (icons differ across clients — the item
   model has an `expansion` field on accounts but item data is version-specific).

## Requirements

- A `WoWItemTooltip` component that renders from the data we already return:
  quality-coloured name, item level/required level, binding, inventory type,
  armor, stats, damage/DPS, resistances, spells, item set, description, prices.
- Use it on the item detail header and as hover over item names in the list and
  in any item id input.
- Optional: item icon next to the name when a source is resolved; graceful
  fallback to no icon.
- Item id inputs in mail/grant forms use the shared autocomplete (001/007) with
  name + quality + icon preview.

## Acceptance criteria

- Hovering an item name shows a tooltip matching the backend data.
- The detail header uses the tooltip styling and quality colour.
- If icons are adopted, they load from an allowed origin and degrade gracefully.
- `task web:check` and `task web:test` green.

## Implementation notes

- Build the tooltip from `AzerothItem`; do not scrape Wowhead, so custom items
  render correctly and no external dependency is introduced.
- If Wowhead integration is explored, keep it strictly additive (a link-out or
  an opt-in widget) and document the CSP change and failure mode.
- Icons require a serving strategy: either a static asset set or a backend
  media endpoint; note the CSP change in `web/Caddyfile`.

## Tests

- RTL tests for the tooltip content and quality colouring with representative
  items (weapon, armor, container, consumable).

## Dependencies

- 001 (autocomplete). Investigation can start immediately and is the main risk
  in this plan.
