import { useMutation, useQueryClient } from "@tanstack/react-query";
import type { ReactNode } from "react";
import { toast } from "sonner";

import { azerothAccountsListQueryKey, azerothAccountsUnbanMutation } from "@/api";
import { isApiError } from "@/api/errors";
import { ConfirmDialog } from "@/components/common/confirm-dialog";

export function UnbanAccountButton({
  username,
  trigger,
}: {
  username: string;
  trigger: ReactNode;
}) {
  const queryClient = useQueryClient();
  const mutation = useMutation(azerothAccountsUnbanMutation());

  return (
    <ConfirmDialog
      trigger={trigger}
      title="Unban account?"
      description={`Lift the ban on ${username}.`}
      confirmLabel="Unban"
      onConfirm={async () => {
        try {
          await mutation.mutateAsync({ path: { username } });
          toast.success("Account unbanned");
          await queryClient.invalidateQueries({ queryKey: azerothAccountsListQueryKey() });
        } catch (error) {
          toast.error(isApiError(error) ? `${error.message} (${error.code})` : String(error));
        }
      }}
    />
  );
}
