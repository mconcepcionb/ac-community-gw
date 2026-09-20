import { zodResolver } from "@hookform/resolvers/zod";
import { useMutation, useQueryClient } from "@tanstack/react-query";
import { useState } from "react";
import { useForm } from "react-hook-form";
import { toast } from "sonner";
import { z } from "zod";

import {
  azerothCharactersBanMutation,
  azerothCharactersGetQueryKey,
  azerothCharactersUnbanMutation,
} from "@/api";
import { isApiError } from "@/api/errors";
import { TextField } from "@/components/common/form-controls";
import { PermissionGate } from "@/components/common/permission-gate";
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

/** CharacterBanActions bans or unbans a character from the console. */
export function CharacterBanActions({ name }: { name: string }) {
  const queryClient = useQueryClient();
  const [open, setOpen] = useState(false);
  const form = useForm<FormValues>({
    resolver: zodResolver(schema),
    defaultValues: { duration: "1d", reason: "" },
  });
  const ban = useMutation(azerothCharactersBanMutation());
  const unban = useMutation(azerothCharactersUnbanMutation());

  const invalidate = () =>
    queryClient.invalidateQueries({
      queryKey: azerothCharactersGetQueryKey({ path: { name } }),
    });

  const reportError = (error: unknown) =>
    toast.error(isApiError(error) ? `${error.message} (${error.code})` : String(error));

  const submit = form.handleSubmit(async (values) => {
    try {
      await ban.mutateAsync({ path: { name }, body: values });
      toast.success("Character banned");
      setOpen(false);
      form.reset({ duration: "1d", reason: "" });
      await invalidate();
    } catch (error) {
      reportError(error);
    }
  });

  const onUnban = async () => {
    try {
      await unban.mutateAsync({ path: { name } });
      toast.success("Character unbanned");
      await invalidate();
    } catch (error) {
      reportError(error);
    }
  };

  return (
    <PermissionGate permission="azeroth.admin.characters.ban">
      <div className="flex gap-2">
        <Dialog open={open} onOpenChange={setOpen}>
          <DialogTrigger asChild>
            <Button variant="destructive" size="sm">
              Ban
            </Button>
          </DialogTrigger>
          <DialogContent>
            <DialogHeader>
              <DialogTitle>Ban character</DialogTitle>
              <DialogDescription>Ban {name} from the game server.</DialogDescription>
            </DialogHeader>
            <Form {...form}>
              <form onSubmit={submit} className="space-y-4">
                <TextField
                  control={form.control}
                  name="duration"
                  label="Duration"
                  placeholder="1d"
                />
                <TextField control={form.control} name="reason" label="Reason" />
                <DialogFooter>
                  <Button type="button" variant="outline" onClick={() => setOpen(false)}>
                    Cancel
                  </Button>
                  <Button type="submit" variant="destructive" disabled={ban.isPending}>
                    Confirm ban
                  </Button>
                </DialogFooter>
              </form>
            </Form>
          </DialogContent>
        </Dialog>
        <Button
          variant="outline"
          size="sm"
          onClick={() => void onUnban()}
          disabled={unban.isPending}
        >
          Unban
        </Button>
      </div>
    </PermissionGate>
  );
}
