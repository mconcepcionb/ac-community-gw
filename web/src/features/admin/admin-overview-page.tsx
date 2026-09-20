import { useQuery } from "@tanstack/react-query";
import { Link } from "@tanstack/react-router";

import { storeAdminOrdersListOptions } from "@/api";
import { PageHeader } from "@/components/common/page-header";
import { PermissionGate } from "@/components/common/permission-gate";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { useAzerothStatus } from "@/features/azeroth-info/use-status";

const links = [
  { to: "/admin/accounts", label: "Accounts", permission: "azeroth.account.list" },
  { to: "/admin/characters", label: "Characters", permission: "azeroth.character.list" },
  { to: "/admin/users", label: "Community users", permission: "gw.identity.user.read" },
  { to: "/admin/items", label: "Items", permission: "azeroth.item.list" },
  { to: "/admin/store", label: "Store", permission: "gw.store.admin.products" },
  { to: "/admin/online", label: "Online players", permission: "azeroth.admin.players.read" },
] as const;

/** AdminOverviewPage is the staff operations overview. */
export function AdminOverviewPage() {
  return (
    <div className="mx-auto max-w-6xl p-8">
      <PageHeader title="Operations console" description="Staff tools for the community gateway." />

      <div className="grid gap-4 lg:grid-cols-2">
        <PermissionGate permission="azeroth.info.public.read">
          <StatusCard />
        </PermissionGate>
        <PermissionGate permission="gw.store.admin.orders.read">
          <RecentOrdersCard />
        </PermissionGate>
      </div>

      <section className="mt-8">
        <h2 className="mb-3 text-lg font-semibold">Areas</h2>
        <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
          {links.map((link) => (
            <PermissionGate key={link.to} permission={link.permission}>
              <Link to={link.to} className="block">
                <Card className="h-full transition-colors hover:border-foreground/30">
                  <CardHeader>
                    <CardTitle>{link.label}</CardTitle>
                  </CardHeader>
                </Card>
              </Link>
            </PermissionGate>
          ))}
        </div>
      </section>
    </div>
  );
}

function StatusCard() {
  const query = useAzerothStatus();

  return (
    <Card>
      <CardHeader>
        <CardTitle>Server status</CardTitle>
      </CardHeader>
      <CardContent className="space-y-1 text-sm">
        {query.isPending ? <p className="text-muted-foreground">Loading…</p> : null}
        {query.isError ? <p className="text-muted-foreground">Status unavailable.</p> : null}
        {query.data ? (
          <>
            <p>
              Connected players:{" "}
              <span className="font-semibold">{query.data.connected_players ?? 0}</span>
            </p>
            <p>
              Queue: <span className="font-semibold">{query.data.queue ?? 0}</span>
            </p>
            <p>
              Uptime: <span className="font-semibold">{query.data.uptime ?? "—"}</span>
            </p>
          </>
        ) : null}
      </CardContent>
    </Card>
  );
}

function RecentOrdersCard() {
  const query = useQuery(storeAdminOrdersListOptions({ query: { limit: 5 } }));
  const orders = query.data?.orders ?? [];

  return (
    <Card>
      <CardHeader>
        <CardTitle>Recent orders</CardTitle>
      </CardHeader>
      <CardContent className="space-y-1 text-sm">
        {query.isPending ? <p className="text-muted-foreground">Loading…</p> : null}
        {query.isError ? <p className="text-muted-foreground">Orders unavailable.</p> : null}
        {!query.isPending && !query.isError && orders.length === 0 ? (
          <p className="text-muted-foreground">No orders.</p>
        ) : null}
        {orders.map((order) => (
          <p key={order.order_id}>
            {order.sku} — {order.status}
            {order.character ? ` (${order.character})` : ""}
          </p>
        ))}
      </CardContent>
    </Card>
  );
}
