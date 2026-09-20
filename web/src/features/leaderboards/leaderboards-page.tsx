import { useQuery } from "@tanstack/react-query";
import { Link } from "@tanstack/react-router";
import { useMemo, useState } from "react";

import { azerothPublicLeaderboardsGetOptions } from "@/api";
import { LoadingState } from "@/components/common/loading-state";
import { PageHeader } from "@/components/common/page-header";
import { Button } from "@/components/ui/button";
import { useMyCharacters } from "@/features/characters/use-my-characters";

const boards = [
  { id: "progression", label: "Progression" },
  { id: "wealth", label: "Wealth" },
  { id: "playtime", label: "Playtime" },
  { id: "pvp", label: "PvP" },
] as const;

const PAGE_SIZE = 25;

function metric(
  board: string,
  entry: { level?: number; money?: number; total_time?: number; arena_points?: number },
) {
  switch (board) {
    case "wealth":
      return `${entry.money ?? 0} copper`;
    case "playtime":
      return `${Math.round((entry.total_time ?? 0) / 3600)}h`;
    case "pvp":
      return `${entry.arena_points ?? 0} arena`;
    default:
      return `level ${entry.level ?? 0}`;
  }
}

/** LeaderboardsPage ranks opted-in characters on one board. */
export function LeaderboardsPage({ board }: { board: string }) {
  const [page, setPage] = useState(0);
  const query = useQuery(
    azerothPublicLeaderboardsGetOptions({
      path: { board },
      query: { limit: PAGE_SIZE, offset: page * PAGE_SIZE },
    }),
  );
  const mine = useMyCharacters({});
  const owned = useMemo(
    () => new Set((mine.data?.characters ?? []).map((character) => character.name ?? "")),
    [mine.data],
  );

  if (query.isPending) {
    return <LoadingState label="Loading leaderboard…" />;
  }

  const entries = query.data?.entries ?? [];

  return (
    <div className="mx-auto max-w-5xl p-8">
      <PageHeader title="Leaderboards" description="Opted-in characters, ranked." />

      <div className="mb-4 flex flex-wrap gap-2">
        {boards.map((item) => (
          <Button
            key={item.id}
            asChild
            variant={item.id === board ? "default" : "outline"}
            size="sm"
          >
            <Link to="/azeroth/leaderboards/$board" params={{ board: item.id }}>
              {item.label}
            </Link>
          </Button>
        ))}
      </div>

      {query.isError ? (
        <p className="text-sm text-muted-foreground">Leaderboard unavailable.</p>
      ) : null}

      <div className="overflow-x-auto rounded-md border border-border">
        <table className="w-full text-left text-sm">
          <thead className="border-b border-border text-muted-foreground">
            <tr>
              <th className="px-3 py-2 font-medium">#</th>
              <th className="px-3 py-2 font-medium">Name</th>
              <th className="px-3 py-2 font-medium">Class</th>
              <th className="px-3 py-2 font-medium">Guild</th>
              <th className="px-3 py-2 font-medium">Score</th>
            </tr>
          </thead>
          <tbody>
            {entries.map((entry) => {
              const isMine = entry.name ? owned.has(entry.name) : false;
              return (
                <tr
                  key={`${entry.rank}-${entry.name}`}
                  className={`border-b border-border/50 last:border-0 ${isMine ? "bg-muted/50" : ""}`}
                >
                  <td className="px-3 py-2">{entry.rank}</td>
                  <td className="px-3 py-2 font-medium">
                    {entry.name}
                    {isMine ? (
                      <span className="ml-2 text-xs text-muted-foreground">you</span>
                    ) : null}
                  </td>
                  <td className="px-3 py-2">{entry.class_name}</td>
                  <td className="px-3 py-2">{entry.guild || "—"}</td>
                  <td className="px-3 py-2">{metric(board, entry)}</td>
                </tr>
              );
            })}
          </tbody>
        </table>
      </div>

      {entries.length === 0 && query.data ? (
        <p className="mt-3 text-sm text-muted-foreground">No ranked characters yet.</p>
      ) : null}

      <div className="mt-4 flex items-center justify-between text-sm text-muted-foreground">
        <span>Page {page + 1}</span>
        <div className="flex gap-2">
          <Button
            variant="outline"
            size="sm"
            disabled={page === 0}
            onClick={() => setPage((current) => Math.max(0, current - 1))}
          >
            Previous
          </Button>
          <Button
            variant="outline"
            size="sm"
            disabled={entries.length < PAGE_SIZE}
            onClick={() => setPage((current) => current + 1)}
          >
            Next
          </Button>
        </div>
      </div>
    </div>
  );
}
