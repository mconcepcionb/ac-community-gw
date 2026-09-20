import { useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";

import { storeProductsDeleteMutation, storeProductsListQueryKey } from "@/api";
import { isApiError } from "@/api/errors";
import { ConfirmDialog } from "@/components/common/confirm-dialog";
import { Button } from "@/components/ui/button";

export function DeleteProductButton({ sku, onDeleted }: { sku: string; onDeleted?: () => void }) {
  const queryClient = useQueryClient();
  const mutation = useMutation(storeProductsDeleteMutation());

  return (
    <ConfirmDialog
      trigger={
        <Button variant="outline" size="sm">
          Deactivate
        </Button>
      }
      title="Deactivate product?"
      description={`Deactivates ${sku} so it can no longer be purchased.`}
      confirmLabel="Deactivate"
      destructive
      onConfirm={async () => {
        try {
          await mutation.mutateAsync({ path: { sku } });
          toast.success("Product deactivated");
          await queryClient.invalidateQueries({ queryKey: storeProductsListQueryKey() });
          onDeleted?.();
        } catch (error) {
          toast.error(isApiError(error) ? `${error.message} (${error.code})` : String(error));
        }
      }}
    />
  );
}
