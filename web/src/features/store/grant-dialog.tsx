import { zodResolver } from "@hookform/resolvers/zod";
import { useMutation, useQueryClient } from "@tanstack/react-query";
import { type ReactNode, useState } from "react";
import { useForm } from "react-hook-form";
import { toast } from "sonner";
import { z } from "zod";

import { storeWalletGetQueryKey, storeWalletsGrantMutation } from "@/api";
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

const schema = z
  .object({
    user_id: z.string(),
    discord_id: z.string(),
    points: z
      .string()
      .refine((value) => /^\d+$/.test(value) && Number(value) > 0, "Positive points"),
    reason: z.string(),
  })
  .refine((values) => values.user_id || values.discord_id, {
    message: "Provide user_id or discord_id",
    path: ["user_id"],
  });

type FormValues = z.infer<typeof schema>;

export function GrantDialog({ trigger }: { trigger: ReactNode }) {
  const [open, setOpen] = useState(false);
  const queryClient = useQueryClient();
  const form = useForm<FormValues>({
    resolver: zodResolver(schema),
    defaultValues: { user_id: "", discord_id: "", points: "", reason: "" },
  });
  const mutation = useMutation(storeWalletsGrantMutation());

  const submit = form.handleSubmit(async (values) => {
    try {
      await mutation.mutateAsync({
        body: {
          user_id: values.user_id || undefined,
          discord_id: values.discord_id || undefined,
          points: Number(values.points),
          reason: values.reason,
        },
      });
      toast.success("Points granted");
      form.reset({ user_id: "", discord_id: "", points: "", reason: "" });
      setOpen(false);
      await queryClient.invalidateQueries({ queryKey: storeWalletGetQueryKey() });
    } catch (error) {
      toast.error(isApiError(error) ? `${error.message} (${error.code})` : String(error));
    }
  });

  return (
    <Dialog open={open} onOpenChange={setOpen}>
      <DialogTrigger asChild>{trigger}</DialogTrigger>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Grant wallet points</DialogTitle>
          <DialogDescription>Add points to a community user's wallet.</DialogDescription>
        </DialogHeader>
        <Form {...form}>
          <form onSubmit={submit} className="space-y-4">
            <TextField control={form.control} name="user_id" label="User id" />
            <TextField control={form.control} name="discord_id" label="Discord id" />
            <TextField control={form.control} name="points" label="Points" type="number" />
            <TextField control={form.control} name="reason" label="Reason" />
            <DialogFooter>
              <Button type="button" variant="outline" onClick={() => setOpen(false)}>
                Cancel
              </Button>
              <ConfirmDialog
                trigger={
                  <Button type="button" disabled={mutation.isPending}>
                    Review grant
                  </Button>
                }
                title="Confirm grant?"
                description="This adds points to the wallet. It cannot be undone."
                confirmLabel="Grant"
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
