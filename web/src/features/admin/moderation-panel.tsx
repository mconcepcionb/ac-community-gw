import { zodResolver } from "@hookform/resolvers/zod";
import { useMutation } from "@tanstack/react-query";
import { useState } from "react";
import { useForm } from "react-hook-form";
import { toast } from "sonner";
import { z } from "zod";

import {
  azerothCharactersBanMutation,
  azerothCharactersUnbanMutation,
  azerothPlayersKickMutation,
  azerothPlayersMuteMutation,
  azerothPlayersUnmuteMutation,
} from "@/api";
import { isApiError } from "@/api/errors";
import { ConfirmDialog } from "@/components/common/confirm-dialog";
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
import { Input } from "@/components/ui/input";

function toastError(error: unknown) {
  toast.error(isApiError(error) ? `${error.message} (${error.code})` : String(error));
}

const kickSchema = z.object({ reason: z.string() });
const muteSchema = z.object({
  duration: z.string().min(1, "Required"),
  reason: z.string().min(1, "Required"),
});

function KickDialog({ name }: { name: string }) {
  const [open, setOpen] = useState(false);
  const form = useForm({ resolver: zodResolver(kickSchema), defaultValues: { reason: "" } });
  const mutation = useMutation(azerothPlayersKickMutation());

  const submit = form.handleSubmit(async (values) => {
    try {
      await mutation.mutateAsync({ path: { name }, body: { reason: values.reason } });
      toast.success("Player kicked");
      setOpen(false);
    } catch (error) {
      toastError(error);
    }
  });

  return (
    <Dialog open={open} onOpenChange={setOpen}>
      <DialogTrigger asChild>
        <Button variant="outline" size="sm" disabled={!name}>
          Kick
        </Button>
      </DialogTrigger>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Kick {name}</DialogTitle>
          <DialogDescription>Disconnect the player from the server.</DialogDescription>
        </DialogHeader>
        <Form {...form}>
          <form onSubmit={submit} className="space-y-4">
            <TextField control={form.control} name="reason" label="Reason (optional)" />
            <DialogFooter>
              <Button type="button" variant="outline" onClick={() => setOpen(false)}>
                Cancel
              </Button>
              <Button type="submit" variant="destructive" disabled={mutation.isPending}>
                Confirm kick
              </Button>
            </DialogFooter>
          </form>
        </Form>
      </DialogContent>
    </Dialog>
  );
}

function MuteDialog({ name }: { name: string }) {
  const [open, setOpen] = useState(false);
  const form = useForm({
    resolver: zodResolver(muteSchema),
    defaultValues: { duration: "1h", reason: "" },
  });
  const mutation = useMutation(azerothPlayersMuteMutation());

  const submit = form.handleSubmit(async (values) => {
    try {
      await mutation.mutateAsync({ path: { name }, body: values });
      toast.success("Player muted");
      setOpen(false);
    } catch (error) {
      toastError(error);
    }
  });

  return (
    <Dialog open={open} onOpenChange={setOpen}>
      <DialogTrigger asChild>
        <Button variant="outline" size="sm" disabled={!name}>
          Mute
        </Button>
      </DialogTrigger>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Mute {name}</DialogTitle>
          <DialogDescription>Temporarily mute the player in chat.</DialogDescription>
        </DialogHeader>
        <Form {...form}>
          <form onSubmit={submit} className="space-y-4">
            <TextField control={form.control} name="duration" label="Duration" placeholder="1h" />
            <TextField control={form.control} name="reason" label="Reason" />
            <DialogFooter>
              <Button type="button" variant="outline" onClick={() => setOpen(false)}>
                Cancel
              </Button>
              <Button type="submit" disabled={mutation.isPending}>
                Confirm mute
              </Button>
            </DialogFooter>
          </form>
        </Form>
      </DialogContent>
    </Dialog>
  );
}

function BanCharacterDialog({ name }: { name: string }) {
  const [open, setOpen] = useState(false);
  const form = useForm({
    resolver: zodResolver(muteSchema),
    defaultValues: { duration: "1d", reason: "" },
  });
  const mutation = useMutation(azerothCharactersBanMutation());

  const submit = form.handleSubmit(async (values) => {
    try {
      await mutation.mutateAsync({ path: { name }, body: values });
      toast.success("Character banned");
      setOpen(false);
    } catch (error) {
      toastError(error);
    }
  });

  return (
    <Dialog open={open} onOpenChange={setOpen}>
      <DialogTrigger asChild>
        <Button variant="outline" size="sm" disabled={!name}>
          Ban character
        </Button>
      </DialogTrigger>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Ban {name}</DialogTitle>
          <DialogDescription>Ban the character from the game.</DialogDescription>
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

function UnmuteButton({ name }: { name: string }) {
  const mutation = useMutation(azerothPlayersUnmuteMutation());
  return (
    <ConfirmDialog
      trigger={
        <Button variant="outline" size="sm" disabled={!name}>
          Unmute
        </Button>
      }
      title="Unmute player?"
      description={`Remove the mute on ${name}.`}
      confirmLabel="Unmute"
      onConfirm={async () => {
        try {
          await mutation.mutateAsync({ path: { name } });
          toast.success("Player unmuted");
        } catch (error) {
          toastError(error);
        }
      }}
    />
  );
}

function UnbanCharacterButton({ name }: { name: string }) {
  const mutation = useMutation(azerothCharactersUnbanMutation());
  return (
    <ConfirmDialog
      trigger={
        <Button variant="outline" size="sm" disabled={!name}>
          Unban character
        </Button>
      }
      title="Unban character?"
      description={`Lift the ban on ${name}.`}
      confirmLabel="Unban"
      onConfirm={async () => {
        try {
          await mutation.mutateAsync({ path: { name } });
          toast.success("Character unbanned");
        } catch (error) {
          toastError(error);
        }
      }}
    />
  );
}

export function ModerationPanel() {
  const [name, setName] = useState("");
  const trimmed = name.trim();

  return (
    <div className="space-y-4">
      <div className="max-w-sm">
        <Input
          placeholder="Player or character name"
          aria-label="Player name"
          value={name}
          onChange={(event) => setName(event.target.value)}
        />
      </div>
      <div className="flex flex-wrap gap-2">
        <PermissionGate permission="azeroth.admin.players.kick">
          <KickDialog name={trimmed} />
        </PermissionGate>
        <PermissionGate permission="azeroth.admin.players.mute">
          <MuteDialog name={trimmed} />
          <UnmuteButton name={trimmed} />
        </PermissionGate>
        <PermissionGate permission="azeroth.admin.characters.ban">
          <BanCharacterDialog name={trimmed} />
          <UnbanCharacterButton name={trimmed} />
        </PermissionGate>
      </div>
    </div>
  );
}
