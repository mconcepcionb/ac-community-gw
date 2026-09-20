import { useQuery } from "@tanstack/react-query";
import type { ColumnDef } from "@tanstack/react-table";
import { useState } from "react";

import type { StoreOrder } from "@/api";
import { storeAdminOrdersListOptions } from "@/api";
import { DataTable } from "@/components/common/data-table";
import { PageHeader } from "@/components/common/page-header";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { useDebouncedValue } from "@/hooks/use-debounced-value";

const PAGE_SIZE = 25;

const columns: ColumnDef<StoreOrder, unknown>[] = [
  { accessorKey: "order_id", header: "Order" },
  { accessorKey: "user_id", header: "User" },
  { accessorKey: "sku", header: "SKU" },
  { accessorKey: "price_points", header: "Points" },
  { accessorKey: "character", header: "Character" },
  { accessorKey: "status", header: "Status" },
  { accessorKey: "created_at", header: "Created" },
];

/** AdminOrdersPage lists every store order with an optional status filter. */
export function AdminOrdersPage() {
  const [statusInput, setStatusInput] = useState("");
  const [page, setPage] = useState(0);
  const status = useDebouncedValue(statusInput, 300);
  const limit = PAGE_SIZE;
  const offset = page * limit;

  const query = useQuery(
    storeAdminOrdersListOptions({ query: { status: status || undefined, limit, offset } }),
  );
  const orders = query.data?.orders ?? [];

  return (
    <div className="mx-auto max-w-6xl p-8">
      <PageHeader title="Orders" description="Every store order, newest first." />

      <div className="mb-4 max-w-xs">
        <Input
          placeholder="Status: pending, delivered, failed"
          aria-label="Filter by status"
          value={statusInput}
          onChange={(event) => {
            setStatusInput(event.target.value);
            setPage(0);
          }}
        />
      </div>

      <DataTable
        columns={columns}
        data={orders}
        loading={query.isPending}
        error={query.isError ? query.error : undefined}
        onRetry={() => void query.refetch()}
        emptyMessage="No orders"
      />

      <div className="mt-4 flex items-center justify-between text-sm text-muted-foreground">
        <span>
          Showing {orders.length === 0 ? 0 : offset + 1}–{offset + orders.length}
        </span>
        <div className="flex gap-2">
          <Button
            variant="outline"
            size="sm"
            disabled={offset === 0}
            onClick={() => setPage((current) => Math.max(0, current - 1))}
          >
            Previous
          </Button>
          <Button
            variant="outline"
            size="sm"
            disabled={orders.length < limit}
            onClick={() => setPage((current) => current + 1)}
          >
            Next
          </Button>
        </div>
      </div>
    </div>
  );
}
