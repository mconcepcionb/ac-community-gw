import { zodResolver } from "@hookform/resolvers/zod";
import { useMutation, useQueryClient } from "@tanstack/react-query";
import { type ReactNode, useEffect, useState } from "react";
import { useFieldArray, useForm } from "react-hook-form";
import { toast } from "sonner";
import { z } from "zod";

import {
  type StoreProduct,
  storeProductsCreateMutation,
  storeProductsGetQueryKey,
  storeProductsListQueryKey,
  storeProductsUpdateMutation,
} from "@/api";
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
import { Switch } from "@/components/ui/switch";
import { Textarea } from "@/components/ui/textarea";

const numberString = z
  .string()
  .refine((value) => value === "" || /^\d+$/.test(value), "Must be a non-negative integer");

const schema = z
  .object({
    sku: z.string(),
    name: z.string(),
    description: z.string(),
    price_points: numberString,
    money: numberString,
    active: z.boolean(),
    items: z.array(z.object({ item_id: numberString, count: numberString })),
  })
  .refine((values) => values.items.length > 0 || Number(values.money || "0") > 0, {
    message: "Provide items or money",
    path: ["money"],
  });

type FormValues = z.infer<typeof schema>;

function initialValues(product?: StoreProduct): FormValues {
  return {
    sku: product?.sku ?? "",
    name: product?.name ?? "",
    description: product?.description ?? "",
    price_points: String(product?.price_points ?? 0),
    money: String(product?.money ?? 0),
    active: product?.active ?? true,
    items: (product?.items ?? []).map((item) => ({
      item_id: String(item.item_id ?? 0),
      count: String(item.count ?? 0),
    })),
  };
}

export function ProductFormDialog({
  mode,
  product,
  trigger,
}: {
  mode: "create" | "update";
  product?: StoreProduct;
  trigger: ReactNode;
}) {
  const [open, setOpen] = useState(false);
  const queryClient = useQueryClient();
  const originalSku = product?.sku;
  const form = useForm<FormValues>({
    resolver: zodResolver(schema),
    defaultValues: initialValues(product),
  });
  const { fields, append, remove } = useFieldArray({ control: form.control, name: "items" });
  const createMutation = useMutation(storeProductsCreateMutation());
  const updateMutation = useMutation(storeProductsUpdateMutation());
  const pending = createMutation.isPending || updateMutation.isPending;

  // Reset the form whenever the selected product changes or the dialog opens.
  useEffect(() => {
    form.reset(initialValues(product));
  }, [product, form]);

  const handleOpenChange = (next: boolean) => {
    if (next) {
      form.reset(initialValues(product));
    }
    setOpen(next);
  };

  const submit = form.handleSubmit(async (values) => {
    try {
      const body = {
        sku: mode === "update" ? (originalSku ?? values.sku) : values.sku,
        name: values.name,
        description: values.description,
        price_points: Number(values.price_points || "0"),
        money: Number(values.money || "0"),
        active: values.active,
        items: values.items.map((item) => ({
          item_id: Number(item.item_id),
          count: Number(item.count),
        })),
      };
      if (mode === "create") {
        await createMutation.mutateAsync({ body });
      } else {
        await updateMutation.mutateAsync({ path: { sku: originalSku ?? values.sku }, body });
      }
      toast.success(mode === "create" ? "Product created" : "Product updated");
      await queryClient.invalidateQueries({ queryKey: storeProductsListQueryKey() });
      for (const sku of new Set([originalSku, values.sku].filter(Boolean))) {
        await queryClient.invalidateQueries({
          queryKey: storeProductsGetQueryKey({ path: { sku: sku as string } }),
        });
      }
      setOpen(false);
    } catch (error) {
      toast.error(isApiError(error) ? `${error.message} (${error.code})` : String(error));
    }
  });

  return (
    <Dialog open={open} onOpenChange={handleOpenChange}>
      <DialogTrigger asChild>{trigger}</DialogTrigger>
      <DialogContent className="max-h-[90vh] overflow-y-auto">
        <DialogHeader>
          <DialogTitle>{mode === "create" ? "Create product" : "Edit product"}</DialogTitle>
          <DialogDescription>
            Products deliver items and/or money. Leave the SKU empty to derive it from a single
            item.
          </DialogDescription>
        </DialogHeader>
        <Form {...form}>
          <form onSubmit={submit} className="space-y-4">
            <TextField
              control={form.control}
              name="sku"
              label="SKU"
              placeholder="item-19019"
              disabled={mode === "update"}
            />
            <TextField control={form.control} name="name" label="Name" />
            <FormField
              control={form.control}
              name="description"
              render={({ field }) => (
                <FormItem>
                  <FormLabel>Description</FormLabel>
                  <FormControl>
                    <Textarea rows={2} {...field} />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              )}
            />
            <div className="grid gap-4 sm:grid-cols-2">
              <TextField
                control={form.control}
                name="price_points"
                label="Price (points)"
                type="number"
              />
              <TextField control={form.control} name="money" label="Money (copper)" type="number" />
            </div>
            <FormField
              control={form.control}
              name="active"
              render={({ field }) => (
                <FormItem className="flex flex-row items-center justify-between rounded-md border border-border p-3">
                  <FormLabel>Active</FormLabel>
                  <FormControl>
                    <Switch checked={field.value} onCheckedChange={field.onChange} />
                  </FormControl>
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
                  onClick={() => append({ item_id: "0", count: "1" })}
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
                      <TextField
                        control={form.control}
                        name={`items.${index}.item_id`}
                        label="Item id"
                        type="number"
                      />
                      <TextField
                        control={form.control}
                        name={`items.${index}.count`}
                        label="Count"
                        type="number"
                      />
                      <Button type="button" variant="ghost" onClick={() => remove(index)}>
                        Remove
                      </Button>
                    </div>
                  ))}
                </div>
              )}
            </div>

            <DialogFooter>
              <Button type="button" variant="outline" onClick={() => setOpen(false)}>
                Cancel
              </Button>
              <Button type="submit" disabled={pending}>
                {mode === "create" ? "Create" : "Save"}
              </Button>
            </DialogFooter>
          </form>
        </Form>
      </DialogContent>
    </Dialog>
  );
}
