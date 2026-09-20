import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { useState } from "react";
import { toast } from "sonner";

import {
  azerothAdminAccountClaimsListOptions,
  reportsCloseMutation,
  reportsListOptions,
  reportsListQueryKey,
  storeAdminOrdersListOptions,
  storeAdminOrdersListQueryKey,
  storeAdminOrdersRefundMutation,
  storeAdminOrdersRetryMutation,
} from "@/api";
import { isApiError } from "@/api/errors";
import { ConfirmDialog } from "@/components/common/confirm-dialog";
import { PageHeader } from "@/components/common/page-header";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";

/** AdminModerationPage is the unified moderation and reconciliation queue. */
export function AdminModerationPage() {
  const queryClient = useQueryClient();
  const pending = useQuery(
    storeAdminOrdersListOptions({ query: { status: "pending", limit: 50 } }),
  );
  const failed = useQuery(storeAdminOrdersListOptions({ query: { status: "failed", limit: 20 } }));
  const claims = useQuery(azerothAdminAccountClaimsListOptions());
  const reports = useQuery(reportsListOptions({ query: { status: "open" } }));

  const refund = useMutation(storeAdminOrdersRefundMutation());
  const retry = useMutation(storeAdminOrdersRetryMutation());
  const closeReport = useMutation(reportsCloseMutation());

  const [reasons, setReasons] = useState<Record<string, string>>({});

  const report = (error: unknown) =>
    toast.error(isApiError(error) ? `${error.message} (${error.code})` : String(error));

  const refreshOrders = () =>
    queryClient.invalidateQueries({ queryKey: storeAdminOrdersListQueryKey() });

  const stuck = (pending.data?.orders ?? []).filter((order) => !order.output);
  const failedOrders = failed.data?.orders ?? [];
  const claimItems = claims.data?.claims ?? [];
  const reportItems = reports.data?.reports ?? [];

  const doRefund = async (id: string) => {
    try {
      await refund.mutateAsync({ path: { id }, body: { reason: reasons[id] ?? "" } });
      toast.success("Order refunded");
      await refreshOrders();
    } catch (error) {
      report(error);
    }
    return true;
  };

  const doRetry = async (id: string) => {
    try {
      await retry.mutateAsync({ path: { id } });
      toast.success("Order delivered");
      await refreshOrders();
    } catch (error) {
      report(error);
    }
    return true;
  };

  const doCloseReport = async (id: string) => {
    try {
      await closeReport.mutateAsync({ path: { id } });
      toast.success("Report closed");
      await queryClient.invalidateQueries({ queryKey: reportsListQueryKey() });
    } catch (error) {
      report(error);
    }
  };

  return (
    <div className="mx-auto max-w-5xl p-8">
      <PageHeader
        title="Moderation"
        description="Stuck orders, failed deliveries, account claims and open reports."
      />

      <section className="mb-8">
        <h2 className="mb-3 text-lg font-semibold">Stuck orders</h2>
        {pending.isPending ? <p className="text-sm text-muted-foreground">Loading…</p> : null}
        {stuck.length === 0 && pending.data ? (
          <p className="text-sm text-muted-foreground">Nothing stuck.</p>
        ) : null}
        <ul className="space-y-2">
          {stuck.map((order) => (
            <li key={order.order_id ?? ""} className="rounded border border-border p-3">
              <div className="flex flex-wrap items-center justify-between gap-2">
                <span className="text-sm">
                  <span className="font-medium">{order.sku}</span> → {order.character} ·{" "}
                  {order.price_points} points
                </span>
                <div className="flex items-center gap-2">
                  <Input
                    className="h-8 w-40"
                    placeholder="Reason"
                    aria-label={`Reason for ${order.order_id ?? ""}`}
                    value={reasons[order.order_id ?? ""] ?? ""}
                    onChange={(event) =>
                      setReasons((current) => ({
                        ...current,
                        [order.order_id ?? ""]: event.target.value,
                      }))
                    }
                  />
                  <ConfirmDialog
                    trigger={
                      <Button variant="outline" size="sm">
                        Refund
                      </Button>
                    }
                    title="Refund this order?"
                    description="The points are returned and the order is marked failed."
                    confirmLabel="Refund"
                    destructive
                    onConfirm={() => doRefund(order.order_id ?? "")}
                  />
                  <ConfirmDialog
                    trigger={
                      <Button variant="outline" size="sm">
                        Retry
                      </Button>
                    }
                    title="Retry delivery?"
                    description="The reward is delivered again and the order completed."
                    confirmLabel="Retry"
                    onConfirm={() => doRetry(order.order_id ?? "")}
                  />
                </div>
              </div>
            </li>
          ))}
        </ul>
      </section>

      <section className="mb-8">
        <h2 className="mb-3 text-lg font-semibold">Failed orders (refunded)</h2>
        {failedOrders.length === 0 ? (
          <p className="text-sm text-muted-foreground">None.</p>
        ) : (
          <ul className="space-y-1 text-sm">
            {failedOrders.map((order) => (
              <li key={order.order_id}>
                {order.sku} → {order.character} · refunded {order.price_points} points
              </li>
            ))}
          </ul>
        )}
      </section>

      <section className="mb-8">
        <h2 className="mb-3 text-lg font-semibold">Pending account claims</h2>
        {claimItems.length === 0 ? (
          <p className="text-sm text-muted-foreground">None.</p>
        ) : (
          <ul className="space-y-1 text-sm">
            {claimItems.map((claim) => (
              <li key={claim.user_id}>
                {claim.account_username} · user {claim.user_id} · expires {claim.expires_at}
              </li>
            ))}
          </ul>
        )}
      </section>

      <section>
        <h2 className="mb-3 text-lg font-semibold">Open reports</h2>
        {reportItems.length === 0 ? (
          <p className="text-sm text-muted-foreground">None.</p>
        ) : (
          <ul className="space-y-2">
            {reportItems.map((item) => (
              <li key={item.id} className="rounded border border-border p-3">
                <div className="flex items-start justify-between gap-2">
                  <div className="text-sm">
                    <p className="font-medium">
                      {item.target} · {item.category}
                    </p>
                    <p className="text-muted-foreground">{item.message}</p>
                  </div>
                  <ConfirmDialog
                    trigger={
                      <Button variant="outline" size="sm">
                        Close
                      </Button>
                    }
                    title="Close this report?"
                    description="Marks the report closed."
                    confirmLabel="Close"
                    onConfirm={() => {
                      void doCloseReport(item.id ?? "");
                      return true;
                    }}
                  />
                </div>
              </li>
            ))}
          </ul>
        )}
      </section>
    </div>
  );
}
