import { useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";

import { azerothAccountLinksDeleteMutation, azerothAccountLinksListQueryKey } from "@/api";
import { isApiError } from "@/api/errors";
import { ConfirmDialog } from "@/components/common/confirm-dialog";
import { Button } from "@/components/ui/button";

export function DeleteLinkButton({
  userId,
  label = "Delete",
  onDeleted,
}: {
  userId: string;
  label?: string;
  onDeleted?: () => void;
}) {
  const queryClient = useQueryClient();
  const mutation = useMutation(azerothAccountLinksDeleteMutation());

  return (
    <ConfirmDialog
      trigger={
        <Button variant="outline" size="sm">
          {label}
        </Button>
      }
      title="Delete account link?"
      description={`This unlinks community user ${userId} from its account.`}
      confirmLabel="Delete"
      destructive
      onConfirm={async () => {
        try {
          await mutation.mutateAsync({ path: { user_id: userId } });
          toast.success("Link deleted");
          await queryClient.invalidateQueries({ queryKey: azerothAccountLinksListQueryKey() });
          onDeleted?.();
        } catch (error) {
          toast.error(isApiError(error) ? `${error.message} (${error.code})` : String(error));
        }
      }}
    />
  );
}
