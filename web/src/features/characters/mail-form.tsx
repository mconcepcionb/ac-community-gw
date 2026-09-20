import { zodResolver } from "@hookform/resolvers/zod";
import { useMutation } from "@tanstack/react-query";
import { useState } from "react";
import { useFieldArray, useForm } from "react-hook-form";
import { toast } from "sonner";
import { z } from "zod";

import { azerothMailSendMutation } from "@/api";
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
import { Input } from "@/components/ui/input";
import { Textarea } from "@/components/ui/textarea";

const schema = z
  .object({
    character: z.string().min(1, "Character is required").max(32),
    subject: z.string().max(200),
    body: z.string().max(2000),
    money: z.number().int().min(0),
    items: z.array(
      z.object({
        id: z.number().int().positive("Positive id"),
        count: z.number().int().positive("Positive count"),
      }),
    ),
  })
  .refine((values) => values.money > 0 || values.items.length > 0, {
    message: "Provide items or money",
    path: ["money"],
  });

type FormValues = z.infer<typeof schema>;

const emptyValues = (character: string): FormValues => ({
  character,
  subject: "",
  body: "",
  money: 0,
  items: [],
});

export function MailForm({ defaultCharacter = "" }: { defaultCharacter?: string }) {
  const [formError, setFormError] = useState<string | null>(null);

  const form = useForm<FormValues>({
    resolver: zodResolver(schema),
    defaultValues: emptyValues(defaultCharacter),
  });
  const { fields, append, remove } = useFieldArray({ control: form.control, name: "items" });
  const mutation = useMutation(azerothMailSendMutation());

  const send = async (values: FormValues) => {
    setFormError(null);
    try {
      await mutation.mutateAsync({
        body: {
          character: values.character,
          subject: values.subject,
          body: values.body,
          money: values.money,
          items: values.items,
        },
      });
      toast.success("Mail sent");
      form.reset(emptyValues(defaultCharacter));
    } catch (error) {
      const message = isApiError(error) ? `${error.message} (${error.code})` : String(error);
      setFormError(message);
      toast.error(message);
    }
  };

  const submit = form.handleSubmit(send);

  return (
    <Form {...form}>
      <form onSubmit={submit} className="space-y-4">
        <div className="grid gap-4 sm:grid-cols-2">
          <FormField
            control={form.control}
            name="character"
            render={({ field }) => (
              <FormItem>
                <FormLabel>Character</FormLabel>
                <FormControl>
                  <Input {...field} />
                </FormControl>
                <FormMessage />
              </FormItem>
            )}
          />
          <FormField
            control={form.control}
            name="money"
            render={({ field }) => (
              <FormItem>
                <FormLabel>Money (copper)</FormLabel>
                <FormControl>
                  <Input
                    type="number"
                    value={field.value}
                    onChange={(event) => field.onChange(event.target.valueAsNumber || 0)}
                  />
                </FormControl>
                <FormMessage />
              </FormItem>
            )}
          />
        </div>

        <FormField
          control={form.control}
          name="subject"
          render={({ field }) => (
            <FormItem>
              <FormLabel>Subject</FormLabel>
              <FormControl>
                <Input {...field} />
              </FormControl>
              <FormMessage />
            </FormItem>
          )}
        />

        <FormField
          control={form.control}
          name="body"
          render={({ field }) => (
            <FormItem>
              <FormLabel>Body</FormLabel>
              <FormControl>
                <Textarea {...field} />
              </FormControl>
              <FormMessage />
            </FormItem>
          )}
        />

        <div>
          <div className="mb-2 flex items-center justify-between">
            <span className="text-sm font-medium">Items</span>
            <Button
              type="button"
              variant="outline"
              size="sm"
              onClick={() => append({ id: 0, count: 1 })}
            >
              Add item
            </Button>
          </div>
          {fields.length === 0 ? (
            <p className="text-sm text-muted-foreground">No items.</p>
          ) : (
            <div className="space-y-2">
              {fields.map((item, index) => (
                <div key={item.id} className="flex items-end gap-2">
                  <FormField
                    control={form.control}
                    name={`items.${index}.id`}
                    render={({ field }) => (
                      <FormItem className="flex-1">
                        <FormLabel>Item id</FormLabel>
                        <FormControl>
                          <Input
                            type="number"
                            value={field.value}
                            onChange={(event) => field.onChange(event.target.valueAsNumber || 0)}
                          />
                        </FormControl>
                        <FormMessage />
                      </FormItem>
                    )}
                  />
                  <FormField
                    control={form.control}
                    name={`items.${index}.count`}
                    render={({ field }) => (
                      <FormItem className="flex-1">
                        <FormLabel>Count</FormLabel>
                        <FormControl>
                          <Input
                            type="number"
                            value={field.value}
                            onChange={(event) => field.onChange(event.target.valueAsNumber || 0)}
                          />
                        </FormControl>
                        <FormMessage />
                      </FormItem>
                    )}
                  />
                  <Button type="button" variant="ghost" onClick={() => remove(index)}>
                    Remove
                  </Button>
                </div>
              ))}
            </div>
          )}
        </div>

        {formError ? (
          <p role="alert" className="text-sm text-destructive">
            {formError}
          </p>
        ) : null}

        <ConfirmDialog
          trigger={
            <Button type="button" disabled={mutation.isPending}>
              Send mail
            </Button>
          }
          title="Send in-game mail?"
          description="This delivers items and/or money to the character and cannot be undone."
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
