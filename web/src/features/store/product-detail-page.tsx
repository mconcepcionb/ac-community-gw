import { useNavigate } from "@tanstack/react-router";

import { ErrorState } from "@/components/common/error-state";
import { LoadingState } from "@/components/common/loading-state";
import { PageHeader } from "@/components/common/page-header";
import { PermissionGate } from "@/components/common/permission-gate";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { DeleteProductButton } from "./delete-product-button";
import { ProductFormDialog } from "./product-form-dialog";
import { PurchaseDialog } from "./purchase-dialog";
import { useProduct } from "./use-products";

export function ProductDetailPage({ sku }: { sku: string }) {
  const query = useProduct(sku);
  const navigate = useNavigate();

  if (sku === "") {
    return (
      <div className="mx-auto max-w-5xl p-8">
        <PageHeader title="Product" description="—" />
        <p className="text-sm text-muted-foreground">Invalid product SKU.</p>
      </div>
    );
  }

  if (query.isPending) {
    return <LoadingState label="Loading product…" />;
  }

  if (query.isError) {
    return (
      <div className="mx-auto max-w-5xl p-8">
        <PageHeader title="Product" description={sku} />
        <ErrorState error={query.error} onRetry={() => void query.refetch()} />
      </div>
    );
  }

  const product = query.data;
  const fields = [
    { label: "SKU", value: product?.sku },
    { label: "Name", value: product?.name },
    { label: "Price (points)", value: product?.price_points },
    { label: "Money (copper)", value: product?.money },
    { label: "Active", value: product?.active ? "yes" : "no" },
  ];

  return (
    <div className="mx-auto max-w-5xl p-8">
      <PageHeader
        title={product?.name ?? sku}
        description={product?.description || "Product detail."}
        actions={
          <div className="flex gap-2">
            <PermissionGate permission="store.purchase">
              <PurchaseDialog sku={sku} trigger={<Button>Buy</Button>} />
            </PermissionGate>
            <PermissionGate permission="store.admin.products">
              <ProductFormDialog
                mode="update"
                product={product}
                trigger={<Button variant="outline">Edit</Button>}
              />
              <DeleteProductButton
                sku={sku}
                onDeleted={() => void navigate({ to: "/admin/store" })}
              />
            </PermissionGate>
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
            <CardContent className="text-lg font-semibold">{field.value ?? "—"}</CardContent>
          </Card>
        ))}
      </div>

      {product?.items && product.items.length > 0 ? (
        <section className="mt-8">
          <h2 className="mb-3 text-lg font-semibold">Items</h2>
          <ul className="space-y-1 text-sm">
            {product.items.map((item) => (
              <li key={`${item.item_id}-${item.count}`}>
                {item.name ?? `Item ${item.item_id}`} × {item.count}
              </li>
            ))}
          </ul>
        </section>
      ) : null}
    </div>
  );
}
