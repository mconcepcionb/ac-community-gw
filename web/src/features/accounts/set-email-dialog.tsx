import { zodResolver } from "@hookform/resolvers/zod";
import { useMutation, useQueryClient } from "@tanstack/react-query";
import { type ReactNode, useState } from "react";
import { useForm } from "react-hook-form";
import { toast } from "sonner";
import { z } from "zod";

import { azerothAccountsListQueryKey, azerothAccountsSetEmailMutation } from "@/api";
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
  email: z.union([z.literal(""), z.string().email("Invalid email")]),
});

type FormValues = z.infer<typeof schema>;

export function SetEmailDialog({
  username,
  currentEmail,
  trigger,
}: {
  username: string;
  currentEmail?: string;
  trigger: ReactNode;
}) {
  const [open, setOpen] = useState(false);
  const queryClient = useQueryClient();
  const form = useForm<FormValues>({
    resolver: zodResolver(schema),
    defaultValues: { email: currentEmail ?? "" },
  });
  const mutation = useMutation(azerothAccountsSetEmailMutation());

  const handleOpenChange = (next: boolean) => {
    if (next) {
      form.reset({ email: currentEmail ?? "" });
    }
    setOpen(next);
  };

  const submit = form.handleSubmit(async (values) => {
    try {
      await mutation.mutateAsync({ path: { username }, body: { username, email: values.email } });
      toast.success("Email updated");
      setOpen(false);
      await queryClient.invalidateQueries({ queryKey: azerothAccountsListQueryKey() });
    } catch (error) {
      toast.error(isApiError(error) ? `${error.message} (${error.code})` : String(error));
    }
  });

  return (
    <Dialog open={open} onOpenChange={handleOpenChange}>
      <DialogTrigger asChild>{trigger}</DialogTrigger>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Set email</DialogTitle>
          <DialogDescription>Change the email of {username}.</DialogDescription>
        </DialogHeader>
        <Form {...form}>
          <form onSubmit={submit} className="space-y-4">
            <TextField control={form.control} name="email" label="Email" />
            <DialogFooter>
              <Button type="button" variant="outline" onClick={() => setOpen(false)}>
                Cancel
              </Button>
              <Button type="submit" disabled={mutation.isPending}>
                Save
              </Button>
            </DialogFooter>
          </form>
        </Form>
      </DialogContent>
    </Dialog>
  );
}
