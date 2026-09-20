import { useQuery } from "@tanstack/react-query";
import { Link } from "@tanstack/react-router";

import {
  azerothMeAccountGetOptions,
  azerothMeCharactersListOptions,
  storeOrdersListOptions,
  storeWalletGetOptions,
} from "@/api";
import { PageHeader } from "@/components/common/page-header";
import { PermissionGate } from "@/components/common/permission-gate";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { useAzerothStatus } from "@/features/azeroth-info/use-status";

/** HomePage is the player portal dashboard. */
export function HomePage() {
  return (
    <div className="mx-auto max-w-5xl p-8">
      <PageHeader
        title="Community portal"
        description="Your characters, store and community status."
      />

      <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
        <PermissionGate permission="azeroth.account.self">
          <AccountCard />
        </PermissionGate>
        <PermissionGate permission="azeroth.character.self">
          <CharactersCard />
        </PermissionGate>
        <PermissionGate permission="store.wallet.read">
          <WalletCard />
        </PermissionGate>
        <PermissionGate permission="store.orders.read">
          <OrdersCard />
        </PermissionGate>
        <PermissionGate permission="azeroth.info.public.read">
          <StatusCard />
        </PermissionGate>
      </div>

      <section className="mt-8">
        <h2 className="mb-3 text-lg font-semibold">Shortcuts</h2>
        <div className="flex flex-wrap gap-2">
          <PermissionGate permission="azeroth.character.self">
            <Button asChild variant="outline" size="sm">
              <Link to="/characters">My characters</Link>
            </Button>
          </PermissionGate>
          <PermissionGate permission="store.catalog.read">
            <Button asChild variant="outline" size="sm">
              <Link to="/store">Store</Link>
            </Button>
          </PermissionGate>
          <PermissionGate permission="store.wallet.read">
            <Button asChild variant="outline" size="sm">
              <Link to="/wallet">Wallet</Link>
            </Button>
          </PermissionGate>
          <PermissionGate permission="azeroth.info.public.read">
            <Button asChild variant="outline" size="sm">
              <Link to="/status">Server status</Link>
            </Button>
          </PermissionGate>
        </div>
      </section>
    </div>
  );
}

function AccountCard() {
  const query = useQuery(azerothMeAccountGetOptions());

  return (
    <Card>
      <CardHeader>
        <CardTitle>Game account</CardTitle>
      </CardHeader>
      <CardContent className="space-y-2 text-sm">
        {query.isPending ? <p className="text-muted-foreground">Loading…</p> : null}
        {query.isError ? <p className="text-muted-foreground">Unavailable.</p> : null}
        {query.data?.linked ? (
          <p>
            Linked to <span className="font-semibold">{query.data.account_username}</span>.
          </p>
        ) : null}
        {query.data && !query.data.linked ? (
          <>
            <p className="text-muted-foreground">No game account linked.</p>
            <Button asChild size="sm">
              <Link to="/onboarding">Link account</Link>
            </Button>
          </>
        ) : null}
      </CardContent>
    </Card>
  );
}

function CharactersCard() {
  const query = useQuery(azerothMeCharactersListOptions());
  const characters = query.data?.characters ?? [];

  return (
    <Card>
      <CardHeader>
        <CardTitle>Characters</CardTitle>
      </CardHeader>
      <CardContent className="text-sm">
        {query.isPending ? <p className="text-muted-foreground">Loading…</p> : null}
        {query.isError ? <p className="text-muted-foreground">Unavailable.</p> : null}
        {query.data ? (
          <p>
            <span className="text-2xl font-semibold">{characters.length}</span>{" "}
            {characters.length === 1 ? "character" : "characters"}
          </p>
        ) : null}
      </CardContent>
    </Card>
  );
}

function WalletCard() {
  const query = useQuery(storeWalletGetOptions());

  return (
    <Card>
      <CardHeader>
        <CardTitle>Points</CardTitle>
      </CardHeader>
      <CardContent className="text-sm">
        {query.isPending ? <p className="text-muted-foreground">Loading…</p> : null}
        {query.isError ? <p className="text-muted-foreground">Unavailable.</p> : null}
        {query.data ? <p className="text-2xl font-semibold">{query.data.balance ?? 0}</p> : null}
      </CardContent>
    </Card>
  );
}

function OrdersCard() {
  const query = useQuery(storeOrdersListOptions({ query: { limit: 5 } }));
  const orders = query.data?.orders ?? [];

  return (
    <Card>
      <CardHeader>
        <CardTitle>Recent orders</CardTitle>
      </CardHeader>
      <CardContent className="space-y-1 text-sm">
        {query.isPending ? <p className="text-muted-foreground">Loading…</p> : null}
        {query.isError ? <p className="text-muted-foreground">Unavailable.</p> : null}
        {query.data && orders.length === 0 ? (
          <p className="text-muted-foreground">No orders.</p>
        ) : null}
        {orders.map((order) => (
          <p key={order.order_id}>
            {order.sku} — {order.status}
          </p>
        ))}
      </CardContent>
    </Card>
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
        {query.isError ? <p className="text-muted-foreground">Unavailable.</p> : null}
        {query.data ? (
          <>
            <p>
              Players: <span className="font-semibold">{query.data.connected_players ?? 0}</span>
            </p>
            <p>
              Queue: <span className="font-semibold">{query.data.queue ?? 0}</span>
            </p>
          </>
        ) : null}
      </CardContent>
    </Card>
  );
}
