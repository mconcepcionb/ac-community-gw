import { zodResolver } from "@hookform/resolvers/zod";
import { useForm } from "react-hook-form";
import { toast } from "sonner";
import { z } from "zod";

import { isApiError } from "@/api/errors";
import { TextField } from "@/components/common/form-controls";
import { PageHeader } from "@/components/common/page-header";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Form } from "@/components/ui/form";
import { useMyReports, useSubmitReport } from "./use-reports";

const schema = z.object({
  target: z.string().min(1, "Required").max(64),
  category: z.string().min(1, "Required").max(64),
  message: z.string().min(1, "Required").max(1000),
});

type FormValues = z.infer<typeof schema>;

const emptyValues: FormValues = { target: "", category: "", message: "" };

/** ReportPage lets a player report another player and see their own reports. */
export function ReportPage() {
  const query = useMyReports();
  const submit = useSubmitReport();
  const form = useForm<FormValues>({ resolver: zodResolver(schema), defaultValues: emptyValues });
  const reports = query.data?.reports ?? [];

  const onSubmit = form.handleSubmit(async (values) => {
    try {
      await submit.mutateAsync({ body: values });
      toast.success("Report submitted");
      form.reset(emptyValues);
    } catch (error) {
      toast.error(isApiError(error) ? `${error.message} (${error.code})` : String(error));
    }
  });

  return (
    <div className="mx-auto max-w-3xl p-8">
      <PageHeader title="Report a player" description="Reports are reviewed by staff." />

      <Card className="mb-8">
        <CardHeader>
          <CardTitle>New report</CardTitle>
          <CardDescription>Describe what happened.</CardDescription>
        </CardHeader>
        <CardContent>
          <Form {...form}>
            <form onSubmit={onSubmit} className="space-y-4">
              <TextField control={form.control} name="target" label="Character or account" />
              <TextField control={form.control} name="category" label="Category" />
              <TextField control={form.control} name="message" label="Message" />
              <Button type="submit" disabled={submit.isPending}>
                Submit report
              </Button>
            </form>
          </Form>
        </CardContent>
      </Card>

      <h2 className="mb-3 text-lg font-semibold">My reports</h2>
      {query.isPending ? <p className="text-sm text-muted-foreground">Loading…</p> : null}
      {query.isError ? <p className="text-sm text-muted-foreground">Unavailable.</p> : null}
      {query.data && reports.length === 0 ? (
        <p className="text-sm text-muted-foreground">No reports.</p>
      ) : null}
      <ul className="space-y-2 text-sm">
        {reports.map((report) => (
          <li key={report.id} className="rounded border border-border p-3">
            <div className="flex items-center justify-between">
              <span className="font-medium">
                {report.target} · {report.category}
              </span>
              <span className="text-muted-foreground">{report.status}</span>
            </div>
            <p className="mt-1 text-muted-foreground">{report.message}</p>
          </li>
        ))}
      </ul>
    </div>
  );
}
