import { zodResolver } from "@hookform/resolvers/zod";
import { useMutation, useQueryClient } from "@tanstack/react-query";
import { type ReactNode, useEffect, useState } from "react";
import { useForm } from "react-hook-form";
import { toast } from "sonner";
import { z } from "zod";

import { azerothAccountsListQueryKey, azerothAccountsSetGmlevelMutation } from "@/api";
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
import {
  Form,
  FormControl,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from "@/components/ui/form";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";

const levelValues = ["0", "1", "2", "3", "4"] as const;

const schema = z.object({
  level: z.enum(levelValues),
  realm: z.string(),
});

type FormValues = z.infer<typeof schema>;

function levelValue(level: number): FormValues["level"] {
  const value = String(level);
  return (levelValues as readonly string[]).includes(value)
    ? (value as FormValues["level"])
    : "0";
}

const levels = [
  { value: "0", label: "0 — Player" },
  { value: "1", label: "1 — Moderator" },
  { value: "2", label: "2 — Game Master" },
  { value: "3", label: "3 — Administrator" },
  { value: "4", label: "4 — Console" },
];

export function SetGmLevelDialog({
  username,
  currentLevel = 0,
  currentRealm = "",
  trigger,
}: {
  username: string;
  currentLevel?: number;
  currentRealm?: string;
  trigger: ReactNode;
}) {
  const [open, setOpen] = useState(false);
  const queryClient = useQueryClient();
  const form = useForm<FormValues>({
    resolver: zodResolver(schema),
    defaultValues: { level: levelValue(currentLevel), realm: currentRealm },
  });
  const mutation = useMutation(azerothAccountsSetGmlevelMutation());

  useEffect(() => {
    if (open) {
      form.reset({ level: levelValue(currentLevel), realm: currentRealm });
    }
  }, [open, currentLevel, currentRealm, form]);

  const submit = form.handleSubmit(async (values) => {
    try {
      await mutation.mutateAsync({
        path: { username },
        body: { username, level: Number(values.level), realm: values.realm },
      });
      toast.success("GM level updated");
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
          <DialogTitle>Set GM level</DialogTitle>
          <DialogDescription>Change the GM level of {username}.</DialogDescription>
        </DialogHeader>
        <Form {...form}>
          <form onSubmit={submit} className="space-y-4">
            <FormField
              control={form.control}
              name="level"
              render={({ field }) => (
                <FormItem>
                  <FormLabel>GM level</FormLabel>
                  <Select value={field.value} onValueChange={field.onChange}>
                    <FormControl>
                      <SelectTrigger>
                        <SelectValue />
                      </SelectTrigger>
                    </FormControl>
                    <SelectContent>
                      {levels.map((level) => (
                        <SelectItem key={level.value} value={level.value}>
                          {level.label}
                        </SelectItem>
                      ))}
                    </SelectContent>
                  </Select>
                  <FormMessage />
                </FormItem>
              )}
            />
            <TextField control={form.control} name="realm" label="Realm (-1 = all)" />
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
