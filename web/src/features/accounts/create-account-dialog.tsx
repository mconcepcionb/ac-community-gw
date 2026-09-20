import { zodResolver } from "@hookform/resolvers/zod";
import { useMutation, useQueryClient } from "@tanstack/react-query";
import { type ReactNode, useState } from "react";
import { useForm } from "react-hook-form";
import { toast } from "sonner";
import { z } from "zod";

import { azerothAccountsCreateMutation, azerothAccountsListQueryKey } from "@/api";
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
  username: z.string().min(1, "Required").max(32),
  password: z.string().min(1, "Required").max(64),
  email: z.union([z.literal(""), z.string().email("Invalid email")]),
});

type FormValues = z.infer<typeof schema>;

const emptyValues: FormValues = { username: "", password: "", email: "" };

export function CreateAccountDialog({ trigger }: { trigger: ReactNode }) {
  const [open, setOpen] = useState(false);
  const queryClient = useQueryClient();
  const form = useForm<FormValues>({ resolver: zodResolver(schema), defaultValues: emptyValues });
  const mutation = useMutation(azerothAccountsCreateMutation());

  const submit = form.handleSubmit(async (values) => {
    try {
      await mutation.mutateAsync({ body: values });
      toast.success("Account created");
      form.reset(emptyValues);
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
          <DialogTitle>Create account</DialogTitle>
          <DialogDescription>Creates an AzerothCore login account.</DialogDescription>
        </DialogHeader>
        <Form {...form}>
          <form onSubmit={submit} className="space-y-4">
            <TextField control={form.control} name="username" label="Username" />
            <TextField control={form.control} name="password" label="Password" type="password" />
            <TextField control={form.control} name="email" label="Email" />
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
