import { zodResolver } from "@hookform/resolvers/zod";
import { useMutation, useQueryClient } from "@tanstack/react-query";
import { type ReactNode, useState } from "react";
import { useForm } from "react-hook-form";
import { toast } from "sonner";
import { z } from "zod";

import { storeOrdersCreateMutation, storeOrdersListQueryKey, storeWalletGetQueryKey } from "@/api";
import { isApiError } from "@/api/errors";
import { ConfirmDialog } from "@/components/common/confirm-dialog";
import { TextField } from "@/components/common/form-controls";
import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog";
import { Form } from "@/components/ui/form";

const schema = z.object({
  sku: z.string().min(1, "Required"),
  character: z.string().min(1, "Required"),
});

type FormValues = z.infer<typeof schema>;

export function PurchaseDialog({ trigger, sku }: { trigger: ReactNode; sku?: string }) {
  const [open, setOpen] = useState(false);
  const queryClient = useQueryClient();
  const form = useForm<FormValues>({
    resolver: zodResolver(schema),
    defaultValues: { sku: sku ?? "", character: "" },
  });
  const mutation = useMutation(storeOrdersCreateMutation());

  const submit = form.handleSubmit(async (values) => {
    try {
      await mutation.mutateAsync({ body: values });
      toast.success("Purchase delivered");
      form.reset({ sku: sku ?? "", character: "" });
      setOpen(false);
      await queryClient.invalidateQueries({ queryKey: storeWalletGetQueryKey() });
      await queryClient.invalidateQueries({ queryKey: storeOrdersListQueryKey() });
    } catch (error) {
      toast.error(isApiError(error) ? `${error.message} (${error.code})` : String(error));
    }
  });

  return (
    <Dialog open={open} onOpenChange={setOpen}>
      <DialogTrigger asChild>{trigger}</DialogTrigger>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Purchase a product</DialogTitle>
          <DialogDescription>
            Points are spent immediately and the reward is delivered in-game.
          </DialogDescription>
        </DialogHeader>
        <Form {...form}>
          <form onSubmit={submit} className="space-y-4">
            <TextField control={form.control} name="sku" label="SKU" />
            <TextField control={form.control} name="character" label="Character" />
            <DialogFooter>
              <Button type="button" variant="outline" onClick={() => setOpen(false)}>
                Cancel
              </Button>
              <ConfirmDialog
                trigger={
                  <Button type="button" disabled={mutation.isPending}>
                    Review purchase
                  </Button>
                }
                title="Confirm purchase?"
                description="This spends points and delivers the reward. It cannot be undone."
                confirmLabel="Buy"
                onConfirm={async () => {
                  if (!(await form.trigger())) {
                    return false;
                  }
                  await submit();
                  return true;
                }}
              />
            </DialogFooter>
          </form>
        </Form>
      </DialogContent>
    </Dialog>
  );
}
