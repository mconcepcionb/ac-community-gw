import type { ColumnDef } from "@tanstack/react-table";

import type { StoreOrder } from "@/api";
import { DataTable } from "@/components/common/data-table";
import { ErrorState } from "@/components/common/error-state";
import { PageHeader } from "@/components/common/page-header";
import { PermissionGate } from "@/components/common/permission-gate";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { PurchaseDialog } from "./purchase-dialog";
import { useOrders, useWallet } from "./use-wallet";

const columns: ColumnDef<StoreOrder, unknown>[] = [
  { accessorKey: "order_id", header: "Order" },
  { accessorKey: "sku", header: "SKU" },
  { accessorKey: "character", header: "Character" },
  { accessorKey: "status", header: "Status" },
  { accessorKey: "price_points", header: "Points" },
  { accessorKey: "created_at", header: "Created" },
];

export function WalletPage() {
  const wallet = useWallet();
  const orders = useOrders();
  const orderItems = orders.data?.orders ?? [];

  return (
    <div className="mx-auto max-w-5xl p-8">
      <PageHeader
        title="Wallet"
        description="Your points balance and order history."
        actions={
          <PermissionGate permission="store.purchase">
            <PurchaseDialog trigger={<Button>Purchase</Button>} />
          </PermissionGate>
        }
      />

      <Card className="mb-8 max-w-xs">
        <CardHeader>
          <CardTitle className="text-sm font-medium text-muted-foreground">Balance</CardTitle>
        </CardHeader>
        <CardContent className="text-3xl font-semibold">
          {wallet.isError ? (
            <ErrorState error={wallet.error} onRetry={() => void wallet.refetch()} />
          ) : wallet.isPending ? (
            "…"
          ) : (
            (wallet.data?.balance ?? 0)
          )}
        </CardContent>
      </Card>

      <h2 className="mb-3 text-lg font-semibold">Orders</h2>
      <DataTable
        columns={columns}
        data={orderItems}
        loading={orders.isPending}
        error={orders.isError ? orders.error : undefined}
        onRetry={() => void orders.refetch()}
        emptyMessage="No orders"
      />
    </div>
  );
}
