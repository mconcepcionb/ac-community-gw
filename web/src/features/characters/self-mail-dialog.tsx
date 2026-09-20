import { zodResolver } from "@hookform/resolvers/zod";
import { useMutation, useQueryClient } from "@tanstack/react-query";
import { type ReactNode, useState } from "react";
import { useForm } from "react-hook-form";
import { toast } from "sonner";
import { z } from "zod";

import type { AzerothMailItem } from "@/api";
import { azerothMeCharactersListQueryKey, azerothMeMailSendMutation } from "@/api";
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
  subject: z.string(),
  body: z.string(),
  money: z.string().refine((value) => value === "" || /^\d+$/.test(value), "Non-negative integer"),
  items: z.string(),
});

type FormValues = z.infer<typeof schema>;

const emptyValues: FormValues = { subject: "", body: "", money: "", items: "" };

function parseItems(value: string): AzerothMailItem[] {
  return value
    .split(",")
    .map((part) => part.trim())
    .filter((part) => part !== "")
    .map((part) => {
      const [id, count] = part.split(":").map((piece) => Number(piece.trim()));
      return { id, count };
    });
}

/** SelfMailDialog mails items or money to one of the user's own characters. */
export function SelfMailDialog({ character, trigger }: { character: string; trigger: ReactNode }) {
  const [open, setOpen] = useState(false);
  const queryClient = useQueryClient();
  const form = useForm<FormValues>({ resolver: zodResolver(schema), defaultValues: emptyValues });
  const mutation = useMutation(azerothMeMailSendMutation());

  const submit = form.handleSubmit(async (values) => {
    const items = parseItems(values.items);
    const money = values.money === "" ? 0 : Number(values.money);
    if (items.length === 0 && money <= 0) {
      toast.error("Provide items or money");
      return;
    }
    try {
      await mutation.mutateAsync({
        body: {
          character,
          subject: values.subject,
          body: values.body,
          money,
          items,
        },
      });
      toast.success("Mail sent");
      form.reset(emptyValues);
      setOpen(false);
      await queryClient.invalidateQueries({ queryKey: azerothMeCharactersListQueryKey() });
    } catch (error) {
      toast.error(isApiError(error) ? `${error.message} (${error.code})` : String(error));
    }
  });

  return (
    <Dialog open={open} onOpenChange={setOpen}>
      <DialogTrigger asChild>{trigger}</DialogTrigger>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Mail {character}</DialogTitle>
          <DialogDescription>Delivers items and/or money to your character.</DialogDescription>
        </DialogHeader>
        <Form {...form}>
          <form onSubmit={submit} className="space-y-4">
            <TextField control={form.control} name="subject" label="Subject" />
            <TextField control={form.control} name="body" label="Body" />
            <TextField control={form.control} name="money" label="Money (copper)" type="number" />
            <TextField
              control={form.control}
              name="items"
              label="Items (id:count, comma separated)"
            />
            <DialogFooter>
              <Button type="button" variant="outline" onClick={() => setOpen(false)}>
                Cancel
              </Button>
              <Button type="submit" disabled={mutation.isPending}>
                Send
              </Button>
            </DialogFooter>
          </form>
        </Form>
      </DialogContent>
    </Dialog>
  );
}
