import { Link } from "@tanstack/react-router";
import type { ColumnDef } from "@tanstack/react-table";
import { useMemo, useState } from "react";

import type { StoreProduct } from "@/api";
import { DataTable } from "@/components/common/data-table";
import { PageHeader } from "@/components/common/page-header";
import { PermissionGate } from "@/components/common/permission-gate";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { DeleteProductButton } from "./delete-product-button";
import { ProductFormDialog } from "./product-form-dialog";
import { useProducts } from "./use-products";

const PAGE_SIZE = 10;

const columns: ColumnDef<StoreProduct, unknown>[] = [
  {
    accessorKey: "sku",
    header: "SKU",
    cell: ({ row }) => (
      <Link
        to="/admin/store/$sku"
        params={{ sku: row.original.sku ?? "" }}
        className="text-blue-400 underline"
      >
        {row.original.sku}
      </Link>
    ),
  },
  { accessorKey: "name", header: "Name" },
  { accessorKey: "price_points", header: "Points" },
  { accessorKey: "money", header: "Money" },
  {
    accessorKey: "active",
    header: "Active",
    cell: ({ row }) => (row.original.active ? "yes" : "no"),
  },
  {
    id: "actions",
    header: "",
    cell: ({ row }) => (
      <PermissionGate permission="gw.store.admin.products">
        <div className="flex gap-2">
          <ProductFormDialog
            mode="update"
            product={row.original}
            trigger={
              <Button variant="outline" size="sm">
                Edit
              </Button>
            }
          />
          <DeleteProductButton sku={row.original.sku ?? ""} />
        </div>
      </PermissionGate>
    ),
  },
];

export function StoreProductsPage() {
  const query = useProducts();
  const products = query.data?.products ?? [];
  const [search, setSearch] = useState("");
  const [page, setPage] = useState(0);

  const filtered = useMemo(() => {
    const needle = search.trim().toLowerCase();
    if (!needle) {
      return products;
    }
    return products.filter(
      (product) =>
        (product.sku ?? "").toLowerCase().includes(needle) ||
        (product.name ?? "").toLowerCase().includes(needle),
    );
  }, [products, search]);

  const pageItems = filtered.slice(page * PAGE_SIZE, page * PAGE_SIZE + PAGE_SIZE);

  return (
    <div className="mx-auto max-w-5xl p-8">
      <PageHeader
        title="Store products"
        description="Catalog of purchasable products."
        actions={
          <PermissionGate permission="gw.store.admin.products">
            <ProductFormDialog mode="create" trigger={<Button>Create product</Button>} />
          </PermissionGate>
        }
      />

      <div className="mb-4 max-w-sm">
        <Input
          placeholder="Search products…"
          aria-label="Search products"
          value={search}
          onChange={(event) => {
            setSearch(event.target.value);
            setPage(0);
          }}
        />
      </div>

      <DataTable
        columns={columns}
        data={pageItems}
        loading={query.isPending}
        error={query.isError ? query.error : undefined}
        onRetry={() => void query.refetch()}
        emptyMessage="No products"
      />

      <div className="mt-4 flex items-center justify-between text-sm text-muted-foreground">
        <span>
          Showing {filtered.length === 0 ? 0 : page * PAGE_SIZE + 1}–
          {page * PAGE_SIZE + pageItems.length} of {filtered.length}
        </span>
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
            disabled={(page + 1) * PAGE_SIZE >= filtered.length}
            onClick={() => setPage((current) => current + 1)}
          >
            Next
          </Button>
        </div>
      </div>
    </div>
  );
}
