import { zodResolver } from "@hookform/resolvers/zod";
import { useMutation, useQueryClient } from "@tanstack/react-query";
import { type ReactNode, useState } from "react";
import { useForm } from "react-hook-form";
import { toast } from "sonner";
import { z } from "zod";

import { azerothAccountLinksCreateMutation, azerothAccountLinksListQueryKey } from "@/api";
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

const schema = z
  .object({
    account_username: z.string().min(1, "Required"),
    user_id: z.string(),
    discord_id: z.string(),
    discord_username: z.string(),
  })
  .refine((values) => values.user_id || values.discord_id || values.discord_username, {
    message: "Provide user_id, discord_id or discord_username",
    path: ["user_id"],
  });

type FormValues = z.infer<typeof schema>;

const emptyValues: FormValues = {
  account_username: "",
  user_id: "",
  discord_id: "",
  discord_username: "",
};

export function CreateLinkDialog({ trigger }: { trigger: ReactNode }) {
  const [open, setOpen] = useState(false);
  const queryClient = useQueryClient();
  const form = useForm<FormValues>({ resolver: zodResolver(schema), defaultValues: emptyValues });
  const mutation = useMutation(azerothAccountLinksCreateMutation());

  const submit = form.handleSubmit(async (values) => {
    try {
      await mutation.mutateAsync({
        body: {
          account_username: values.account_username,
          user_id: values.user_id || undefined,
          discord_id: values.discord_id || undefined,
          discord_username: values.discord_username || undefined,
        },
      });
      toast.success("Link created");
      form.reset(emptyValues);
      setOpen(false);
      await queryClient.invalidateQueries({ queryKey: azerothAccountLinksListQueryKey() });
    } catch (error) {
      toast.error(isApiError(error) ? `${error.message} (${error.code})` : String(error));
    }
  });

  return (
    <Dialog open={open} onOpenChange={setOpen}>
      <DialogTrigger asChild>{trigger}</DialogTrigger>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Link an account</DialogTitle>
          <DialogDescription>
            Identify the community user by id, Discord id or Discord username.
          </DialogDescription>
        </DialogHeader>
        <Form {...form}>
          <form onSubmit={submit} className="space-y-4">
            <TextField control={form.control} name="account_username" label="Account username" />
            <TextField control={form.control} name="user_id" label="User id" />
            <TextField control={form.control} name="discord_id" label="Discord id" />
            <TextField control={form.control} name="discord_username" label="Discord username" />
            <DialogFooter>
              <Button type="button" variant="outline" onClick={() => setOpen(false)}>
                Cancel
              </Button>
              <Button type="submit" disabled={mutation.isPending}>
                Create
              </Button>
            </DialogFooter>
          </form>
        </Form>
      </DialogContent>
    </Dialog>
  );
}
