import { Link } from "@tanstack/react-router";

import type { AzerothEquipmentSlot } from "@/api";
import { ErrorState } from "@/components/common/error-state";
import { LoadingState } from "@/components/common/loading-state";
import { PageHeader } from "@/components/common/page-header";
import { PermissionGate } from "@/components/common/permission-gate";
import { StatusBadge } from "@/components/common/status-badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { CharacterBanActions } from "@/features/admin/character-ban-actions";
import { AnnotationsPanel } from "@/features/annotations/annotations-panel";
import { EntityHistory } from "@/features/audit/entity-history";
import { AdminMailDialog } from "./admin-mail-dialog";
import { useCharacter } from "./use-characters";

const slotNames = [
  "Head",
  "Neck",
  "Shoulder",
  "Shirt",
  "Chest",
  "Waist",
  "Legs",
  "Feet",
  "Wrist",
  "Hands",
  "Ring",
  "Ring",
  "Trinket",
  "Trinket",
  "Back",
  "Main Hand",
  "Off Hand",
  "Ranged",
  "Tabard",
];

function Equipment({ equipment }: { equipment: AzerothEquipmentSlot[] }) {
  if (equipment.length === 0) {
    return <p className="text-sm text-muted-foreground">No equipped items.</p>;
  }
  return (
    <ul className="space-y-1 text-sm">
      {equipment.map((slot) => (
        <li
          key={slot.slot ?? slot.entry}
          className="flex flex-wrap items-center justify-between gap-2 rounded border border-border px-3 py-2"
        >
          <span className="text-muted-foreground">
            {slot.slot !== undefined ? (slotNames[slot.slot] ?? `Slot ${slot.slot}`) : "Slot"}
          </span>
          <span>
            <Link
              to="/admin/azeroth/items/$entry"
              params={{ entry: String(slot.entry) }}
              className="text-blue-400 underline"
            >
              {slot.name || `Item ${slot.entry}`}
            </Link>
            {slot.count && slot.count > 1 ? (
              <span className="ml-2 text-xs text-muted-foreground">×{slot.count}</span>
            ) : null}
          </span>
        </li>
      ))}
    </ul>
  );
}

/** CharacterDetailPage is the admin view of one AzerothCore character. */
export function CharacterDetailPage({ name }: { name: string }) {
  const query = useCharacter(name);

  if (query.isPending) {
    return <LoadingState label="Loading character…" />;
  }

  if (query.isError) {
    return (
      <div className="mx-auto max-w-5xl p-8">
        <PageHeader title="Character" description={name} />
        <ErrorState error={query.error} onRetry={() => void query.refetch()} />
      </div>
    );
  }

  const character = query.data;
  const fields = [
    { label: "GUID", value: character?.guid },
    { label: "Level", value: character?.level },
    { label: "Class", value: character?.class_name },
    { label: "Race", value: character?.race_name },
    { label: "Guild", value: character?.guild },
    { label: "Online", value: character?.online ? "yes" : "no" },
    { label: "Money", value: character?.money },
    { label: "Last logout", value: character?.logout_time ?? "never" },
  ];

  return (
    <div className="mx-auto max-w-5xl p-8">
      <PageHeader
        title={character?.name ?? name}
        description="Character detail, equipment and staff tools."
        actions={
          <div className="flex flex-wrap items-center gap-2">
            <StatusBadge tone={character?.online ? "positive" : "neutral"}>
              {character?.online ? "Online" : "Offline"}
            </StatusBadge>
            <StatusBadge tone={character?.banned ? "negative" : "positive"}>
              {character?.banned ? "Banned" : "Active"}
            </StatusBadge>
            <PermissionGate permission="azeroth.admin.mail.send">
              <AdminMailDialog
                character={name}
                trigger={
                  <Button variant="outline" size="sm">
                    Send mail
                  </Button>
                }
              />
            </PermissionGate>
            <CharacterBanActions name={name} banned={character?.banned ?? false} />
          </div>
        }
      />

      {character?.banned && character.ban_reason ? (
        <p className="mb-4 text-sm text-destructive">Ban reason: {character.ban_reason}</p>
      ) : null}

      <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
        {fields.map((field) => (
          <Card key={field.label}>
            <CardHeader>
              <CardTitle className="text-sm font-medium text-muted-foreground">
                {field.label}
              </CardTitle>
            </CardHeader>
            <CardContent className="text-lg font-semibold">{field.value ?? "—"}</CardContent>
          </Card>
        ))}
      </div>

      <section className="mt-8">
        <h2 className="mb-3 text-lg font-semibold">Equipment</h2>
        <Equipment equipment={character?.equipment ?? []} />
      </section>

      <section className="mt-8">
        <h2 className="mb-3 text-lg font-semibold">Annotations</h2>
        <AnnotationsPanel targetType="character" targetId={name} />
      </section>

      <section className="mt-8">
        <h2 className="mb-3 text-lg font-semibold">History</h2>
        <EntityHistory targetType="character" targetId={name} />
      </section>
    </div>
  );
}
