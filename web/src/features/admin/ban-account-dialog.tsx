import { zodResolver } from "@hookform/resolvers/zod";
import { useMutation, useQueryClient } from "@tanstack/react-query";
import { type ReactNode, useState } from "react";
import { useForm } from "react-hook-form";
import { toast } from "sonner";
import { z } from "zod";

import { azerothAccountsBanMutation, azerothAccountsListQueryKey } from "@/api";
import { isApiError } from "@/api/errors";
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
  duration: z.string().min(1, "Required"),
  reason: z.string().min(1, "Required"),
});

type FormValues = z.infer<typeof schema>;

export function BanAccountDialog({ username, trigger }: { username: string; trigger: ReactNode }) {
  const [open, setOpen] = useState(false);
  const queryClient = useQueryClient();
  const form = useForm<FormValues>({
    resolver: zodResolver(schema),
    defaultValues: { duration: "1d", reason: "" },
  });
  const mutation = useMutation(azerothAccountsBanMutation());

  const submit = form.handleSubmit(async (values) => {
    try {
      await mutation.mutateAsync({ path: { username }, body: { username, ...values } });
      toast.success("Account banned");
      form.reset({ duration: "1d", reason: "" });
      setOpen(false);
      await queryClient.invalidateQueries({ queryKey: azerothAccountsListQueryKey() });
    } catch (error) {
      toast.error(isApiError(error) ? `${error.message} (${error.code})` : String(error));
    }
  });

  return (
    <Dialog open={open} onOpenChange={setOpen}>
      <DialogTrigger asChild>{trigger}</DialogTrigger>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Ban account</DialogTitle>
          <DialogDescription>Ban {username} from the game server.</DialogDescription>
        </DialogHeader>
        <Form {...form}>
          <form onSubmit={submit} className="space-y-4">
            <TextField control={form.control} name="duration" label="Duration" placeholder="1d" />
            <TextField control={form.control} name="reason" label="Reason" />
            <DialogFooter>
              <Button type="button" variant="outline" onClick={() => setOpen(false)}>
                Cancel
              </Button>
              <Button type="submit" variant="destructive" disabled={mutation.isPending}>
                Confirm ban
              </Button>
            </DialogFooter>
          </form>
        </Form>
      </DialogContent>
    </Dialog>
  );
}
