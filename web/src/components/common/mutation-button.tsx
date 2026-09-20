import type { ComponentProps } from "react";
import { toast } from "sonner";

import { isApiError } from "@/api/errors";
import { Button } from "@/components/ui/button";

interface MutationButtonProps extends Omit<ComponentProps<typeof Button>, "onClick"> {
  onAction: () => Promise<void> | void;
  pending?: boolean;
  pendingLabel?: string;
}

/**
 * MutationButton wraps Button with a pending state and surfaces any ApiError
 * through a toast, so feature code can focus on the happy path.
 */
export function MutationButton({
  onAction,
  pending = false,
  pendingLabel = "Working…",
  children,
  disabled,
  ...props
}: MutationButtonProps) {
  return (
    <Button
      {...props}
      disabled={pending || disabled}
      onClick={async () => {
        try {
          await onAction();
        } catch (error) {
          toast.error(isApiError(error) ? `${error.message} (${error.code})` : String(error));
        }
      }}
    >
      {pending ? pendingLabel : children}
    </Button>
  );
}
