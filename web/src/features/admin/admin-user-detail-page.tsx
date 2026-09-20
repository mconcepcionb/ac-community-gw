import { useQuery, useQueryClient } from "@tanstack/react-query";
import { Link } from "@tanstack/react-router";

import { azerothAdminUsersGetOptions, azerothAdminUsersGetQueryKey } from "@/api";
import { ErrorState } from "@/components/common/error-state";
import { LoadingState } from "@/components/common/loading-state";
import { PageHeader } from "@/components/common/page-header";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { CreateLinkDialog } from "@/features/account-links/create-link-dialog";
import { DeleteLinkButton } from "@/features/account-links/delete-link-button";

/** AdminUserDetailPage is the staff 360 view of a community user. */
export function AdminUserDetailPage({ userId }: { userId: string }) {
  const queryClient = useQueryClient();
  const query = useQuery(azerothAdminUsersGetOptions({ path: { id: userId } }));

  if (query.isPending) {
    return <LoadingState label="Loading user…" />;
  }
  if (query.isError || !query.data) {
    return (
      <div className="mx-auto max-w-5xl p-8">
        <PageHeader title="Community user" description={userId} />
        <ErrorState error={query.error} onRetry={() => void query.refetch()} />
      </div>
    );
  }

  const user = query.data;
  const name = user.display_name || user.username || user.discord_id || userId;
  const characters = user.characters ?? [];
  const orders = user.orders ?? [];
  const roles = user.roles ?? [];

  const invalidate = () =>
    queryClient.invalidateQueries({
      queryKey: azerothAdminUsersGetQueryKey({ path: { id: userId } }),
    });

  const fields = [
    { label: "Display name", value: user.display_name },
    { label: "Username", value: user.username },
    { label: "Global name", value: user.global_name },
    { label: "Discord ID", value: user.discord_id },
    { label: "Community user ID", value: user.user_id },
    { label: "Member since", value: user.created_at },
  ];

  return (
    <div className="mx-auto max-w-5xl p-8">
      <PageHeader title={name} description="Community user 360 view." />

      <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
        {fields.map((field) => (
          <Card key={field.label}>
            <CardHeader>
              <CardTitle className="text-sm font-medium text-muted-foreground">
                {field.label}
              </CardTitle>
            </CardHeader>
            <CardContent className="break-all text-lg font-semibold">
              {field.value || "—"}
            </CardContent>
          </Card>
        ))}
      </div>

      <section className="mt-8">
        <h2 className="mb-3 text-lg font-semibold">Roles</h2>
        {roles.length === 0 ? (
          <p className="text-sm text-muted-foreground">No roles.</p>
        ) : (
          <ul className="flex flex-wrap gap-2">
            {roles.map((role) => (
              <li key={role} className="rounded border border-border px-2 py-1 text-xs">
                {role}
              </li>
            ))}
          </ul>
        )}
      </section>

      <section className="mt-8">
        <h2 className="mb-3 text-lg font-semibold">Linked account</h2>
        {user.account_username ? (
          <div className="flex items-center gap-3">
            <span className="text-sm">
              {user.account_username}
              {user.account_id ? ` (#${user.account_id})` : ""}
            </span>
            <DeleteLinkButton userId={userId} label="Unlink" onDeleted={() => void invalidate()} />
          </div>
        ) : (
          <div className="flex items-center gap-3">
            <span className="text-sm text-muted-foreground">Not linked.</span>
            <CreateLinkDialog
              trigger={
                <Button variant="outline" size="sm">
                  Link account
                </Button>
              }
            />
          </div>
        )}
      </section>

      <section className="mt-8">
        <h2 className="mb-3 text-lg font-semibold">Characters</h2>
        {characters.length === 0 ? (
          <p className="text-sm text-muted-foreground">No characters.</p>
        ) : (
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
                  <th className="px-3 py-2 font-medium">Money</th>
                </tr>
              </thead>
              <tbody>
                {characters.map((character) => (
                  <tr key={character.guid} className="border-b border-border/50 last:border-0">
                    <td className="px-3 py-2">{character.name}</td>
                    <td className="px-3 py-2">{character.level}</td>
                    <td className="px-3 py-2">{character.class}</td>
                    <td className="px-3 py-2">{character.race}</td>
                    <td className="px-3 py-2">{character.guild || "—"}</td>
                    <td className="px-3 py-2">{character.online ? "yes" : "no"}</td>
                    <td className="px-3 py-2">{character.money}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </section>

      <section className="mt-8">
        <h2 className="mb-3 text-lg font-semibold">Wallet</h2>
        <p className="text-sm">{user.wallet ?? 0} points</p>
      </section>

      <section className="mt-8">
        <h2 className="mb-3 text-lg font-semibold">Recent orders</h2>
        {orders.length === 0 ? (
          <p className="text-sm text-muted-foreground">No orders.</p>
        ) : (
          <div className="overflow-x-auto rounded-md border border-border">
            <table className="w-full text-left text-sm">
              <thead className="border-b border-border text-muted-foreground">
                <tr>
                  <th className="px-3 py-2 font-medium">Order</th>
                  <th className="px-3 py-2 font-medium">SKU</th>
                  <th className="px-3 py-2 font-medium">Points</th>
                  <th className="px-3 py-2 font-medium">Character</th>
                  <th className="px-3 py-2 font-medium">Status</th>
                  <th className="px-3 py-2 font-medium">Created</th>
                </tr>
              </thead>
              <tbody>
                {orders.map((order) => (
                  <tr key={order.order_id} className="border-b border-border/50 last:border-0">
                    <td className="px-3 py-2 font-mono text-xs">{order.order_id}</td>
                    <td className="px-3 py-2">{order.sku}</td>
                    <td className="px-3 py-2">{order.points}</td>
                    <td className="px-3 py-2">{order.character}</td>
                    <td className="px-3 py-2">{order.status}</td>
                    <td className="px-3 py-2">{order.created_at}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </section>

      <section className="mt-8">
        <h2 className="mb-3 text-lg font-semibold">Audit</h2>
        <Button asChild variant="outline" size="sm">
          <Link to="/admin/audit" search={{ actor: userId }}>
            View audit entries
          </Link>
        </Button>
      </section>
    </div>
  );
}
