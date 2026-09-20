import { Link } from "@tanstack/react-router";
import { ErrorState } from "@/components/common/error-state";
import { LoadingState } from "@/components/common/loading-state";
import { PageHeader } from "@/components/common/page-header";
import { StatusBadge } from "@/components/common/status-badge";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { useAccount } from "@/features/accounts/use-accounts";
import { AnnotationsPanel } from "@/features/annotations/annotations-panel";
import { EntityHistory } from "@/features/audit/entity-history";
import { useCharacters } from "@/features/characters/use-characters";

/** AdminAccountDetailPage aggregates everything known about one account. */
export function AdminAccountDetailPage({ username }: { username: string }) {
  const query = useAccount(username);

  if (query.isPending) {
    return <LoadingState label="Loading account…" />;
  }
  if (query.isError || !query.data) {
    return (
      <div className="mx-auto max-w-5xl p-8">
        <PageHeader title="Account" description={username} />
        <ErrorState error={query.error} onRetry={() => void query.refetch()} />
      </div>
    );
  }

  const account = query.data.account;
  const owner = account?.claimed_by;
  const claim = query.data.claim;

  const fields = [
    { label: "ID", value: account?.id },
    { label: "Username", value: account?.username },
    { label: "Email", value: account?.email },
    { label: "GM level", value: account?.gm_level },
    { label: "Expansion", value: account?.expansion },
    { label: "Last IP", value: account?.last_ip },
    { label: "Last login", value: account?.last_login ?? "never" },
  ];

  return (
    <div className="mx-auto max-w-5xl p-8">
      <PageHeader
        title={account?.username ?? username}
        description="Account detail, ownership and staff notes."
        actions={
          <div className="flex gap-2">
            <StatusBadge tone={account?.online ? "positive" : "neutral"}>
              {account?.online ? "Online" : "Offline"}
            </StatusBadge>
            <StatusBadge tone={account?.banned ? "negative" : "positive"}>
              {account?.banned ? "Banned" : "Active"}
            </StatusBadge>
          </div>
        }
      />

      <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
        {fields.map((field) => (
          <Card key={field.label}>
            <CardHeader>
              <CardTitle className="text-sm font-medium text-muted-foreground">
                {field.label}
              </CardTitle>
            </CardHeader>
            <CardContent className="break-all text-lg font-semibold">
              {field.value ?? "—"}
            </CardContent>
          </Card>
        ))}
      </div>

      {account?.banned && account.ban_reason ? (
        <p className="mt-4 text-sm text-destructive">Ban reason: {account.ban_reason}</p>
      ) : null}

      <section className="mt-8">
        <h2 className="mb-3 text-lg font-semibold">Owner</h2>
        {owner ? (
          <div className="text-sm">
            {owner.display_name || owner.user_id}
            {owner.discord_id ? ` · Discord ${owner.discord_id}` : ""}
            {owner.user_id ? (
              <Link
                to="/admin/users/$userId"
                params={{ userId: owner.user_id }}
                className="ml-2 text-blue-400 underline"
              >
                View user
              </Link>
            ) : null}
          </div>
        ) : (
          <p className="text-sm text-muted-foreground">Unclaimed.</p>
        )}
      </section>

      {claim ? (
        <section className="mt-8">
          <h2 className="mb-3 text-lg font-semibold">Pending claim</h2>
          <p className="text-sm">
            user {claim.user_id} · expires {claim.expires_at} · attempts {claim.attempts}
          </p>
        </section>
      ) : null}

      <AccountCharacters username={username} />

      <section className="mt-8">
        <h2 className="mb-3 text-lg font-semibold">Annotations</h2>
        <AnnotationsPanel targetType="account" targetId={username} />
      </section>

      <section className="mt-8">
        <h2 className="mb-3 text-lg font-semibold">History</h2>
        <EntityHistory targetType="account" targetId={username} />
      </section>
    </div>
  );
}

function AccountCharacters({ username }: { username: string }) {
  const query = useCharacters({ account: username });
  const characters = query.data?.characters ?? [];

  return (
    <section className="mt-8">
      <h2 className="mb-3 text-lg font-semibold">Characters</h2>
      {query.isPending ? <p className="text-sm text-muted-foreground">Loading…</p> : null}
      {query.isError ? (
        <p className="text-sm text-muted-foreground">Characters unavailable.</p>
      ) : null}
      {!query.isPending && !query.isError && characters.length === 0 ? (
        <p className="text-sm text-muted-foreground">No characters.</p>
      ) : null}
      {characters.length > 0 ? (
        <div className="overflow-x-auto rounded-md border border-border">
          <table className="w-full text-left text-sm">
            <thead className="border-b border-border text-muted-foreground">
              <tr>
                <th className="px-3 py-2 font-medium">Name</th>
                <th className="px-3 py-2 font-medium">Level</th>
                <th className="px-3 py-2 font-medium">Class</th>
                <th className="px-3 py-2 font-medium">Race</th>
                <th className="px-3 py-2 font-medium">Guild</th>
                <th className="px-3 py-2 font-medium">Online</th>
              </tr>
            </thead>
            <tbody>
              {characters.map((character) => (
                <tr key={character.guid} className="border-b border-border/50 last:border-0">
                  <td className="px-3 py-2">
                    <Link
                      to="/admin/characters/$name"
                      params={{ name: character.name ?? "" }}
                      className="text-blue-400 underline"
                    >
                      {character.name}
                    </Link>
                  </td>
                  <td className="px-3 py-2">{character.level}</td>
                  <td className="px-3 py-2">{character.class_name}</td>
                  <td className="px-3 py-2">{character.race_name}</td>
                  <td className="px-3 py-2">{character.guild || "—"}</td>
                  <td className="px-3 py-2">
                    <StatusBadge tone={character.online ? "positive" : "neutral"}>
                      {character.online ? "Online" : "Offline"}
                    </StatusBadge>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      ) : null}
    </section>
  );
}
