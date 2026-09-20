/**
 * Games the gateway serves. Per-game route subtrees and navigation are derived
 * from this list, so adding a game adds a section rather than rewriting links.
 */
export const GAMES = [{ id: "azeroth", label: "AzerothCore" }] as const;

/** Game is one entry of {@link GAMES}. */
export type Game = (typeof GAMES)[number];
