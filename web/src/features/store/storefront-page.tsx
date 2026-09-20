import { Link } from "@tanstack/react-router";

import { PageHeader } from "@/components/common/page-header";
import { PermissionGate } from "@/components/common/permission-gate";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { PurchaseDialog } from "./purchase-dialog";
import { useProducts } from "./use-products";

/** StorefrontPage lists active products a player can buy with points. */
export function StorefrontPage() {
  const query = useProducts();
  const products = (query.data?.products ?? []).filter((product) => product.active !== false);

  return (
    <div className="mx-auto max-w-5xl p-8">
      <PageHeader title="Store" description="Spend points on rewards delivered in game." />

      {query.isPending ? <p className="text-sm text-muted-foreground">Loading…</p> : null}
      {query.isError ? <p className="text-sm text-muted-foreground">Store unavailable.</p> : null}

      <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
        {products.map((product) => (
          <Card key={product.sku}>
            <CardHeader>
              <CardTitle>
                <Link
                  to="/store/products/$sku"
                  params={{ sku: product.sku ?? "" }}
                  className="hover:underline"
                >
                  {product.name}
                </Link>
              </CardTitle>
              <CardDescription>{product.description || "—"}</CardDescription>
            </CardHeader>
            <CardContent className="space-y-2">
              <p className="text-sm font-semibold">{product.price_points ?? 0} points</p>
              <PermissionGate permission="gw.store.purchase">
                <PurchaseDialog
                  sku={product.sku ?? ""}
                  trigger={
                    <Button size="sm" variant="outline">
                      Buy
                    </Button>
                  }
                />
              </PermissionGate>
            </CardContent>
          </Card>
        ))}
      </div>
    </div>
  );
}
