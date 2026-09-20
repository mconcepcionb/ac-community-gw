import { zodResolver } from "@hookform/resolvers/zod";
import { useMutation } from "@tanstack/react-query";
import { useForm } from "react-hook-form";
import { toast } from "sonner";
import { z } from "zod";

import { azerothAnnounceSendMutation } from "@/api";
import { isApiError } from "@/api/errors";
import { ConfirmDialog } from "@/components/common/confirm-dialog";
import { Button } from "@/components/ui/button";
import {
  Form,
  FormControl,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from "@/components/ui/form";
import { Textarea } from "@/components/ui/textarea";

const schema = z.object({
  message: z.string().min(1, "Required").max(200),
});

type FormValues = z.infer<typeof schema>;

export function AnnounceForm() {
  const form = useForm<FormValues>({
    resolver: zodResolver(schema),
    defaultValues: { message: "" },
  });
  const mutation = useMutation(azerothAnnounceSendMutation());

  const submit = form.handleSubmit(async (values) => {
    try {
      await mutation.mutateAsync({ body: values });
      toast.success("Announcement sent");
      form.reset({ message: "" });
    } catch (error) {
      toast.error(isApiError(error) ? `${error.message} (${error.code})` : String(error));
    }
  });

  return (
    <Form {...form}>
      <form onSubmit={(event) => event.preventDefault()} className="max-w-xl space-y-4">
        <FormField
          control={form.control}
          name="message"
          render={({ field }) => (
            <FormItem>
              <FormLabel>Message</FormLabel>
              <FormControl>
                <Textarea rows={3} {...field} />
              </FormControl>
              <FormMessage />
            </FormItem>
          )}
        />
        <ConfirmDialog
          trigger={
            <Button type="button" disabled={mutation.isPending}>
              Announce
            </Button>
          }
          title="Broadcast announcement?"
          description="Sends a server-wide announcement to every online player."
          confirmLabel="Send"
          onConfirm={async () => {
            if (!(await form.trigger())) {
              return false;
            }
            await submit();
            return true;
          }}
        />
      </form>
    </Form>
  );
}
