import { useNavigate } from "@tanstack/react-router";

import { ErrorState } from "@/components/common/error-state";
import { LoadingState } from "@/components/common/loading-state";
import { PageHeader } from "@/components/common/page-header";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { DeleteLinkButton } from "./delete-link-button";
import { useAccountLink } from "./use-account-links";

export function AccountLinkDetailPage({ userId }: { userId: string }) {
  const query = useAccountLink(userId);
  const navigate = useNavigate();

  if (userId === "") {
    return (
      <div className="mx-auto max-w-5xl p-8">
        <PageHeader title="Account link" description="—" />
        <p className="text-sm text-muted-foreground">Invalid user id.</p>
      </div>
    );
  }

  if (query.isPending) {
    return <LoadingState label="Loading link…" />;
  }

  if (query.isError) {
    return (
      <div className="mx-auto max-w-5xl p-8">
        <PageHeader title="Account link" description={userId} />
        <ErrorState error={query.error} onRetry={() => void query.refetch()} />
      </div>
    );
  }

  const link = query.data;
  const fields = [
    { label: "User ID", value: link?.user_id },
    { label: "Account", value: link?.account_username },
    { label: "Account ID", value: link?.account_id },
    { label: "Linked at", value: link?.linked_at },
  ];

  return (
    <div className="mx-auto max-w-5xl p-8">
      <PageHeader
        title="Account link"
        description={link?.account_username ?? userId}
        actions={
          <DeleteLinkButton
            userId={userId}
            label="Delete link"
            onDeleted={() => void navigate({ to: "/account-links" })}
          />
        }
      />

      <div className="grid gap-4 sm:grid-cols-2">
        {fields.map((field) => (
          <Card key={field.label}>
            <CardHeader>
              <CardTitle className="text-sm font-medium text-muted-foreground">
                {field.label}
              </CardTitle>
            </CardHeader>
            <CardContent className="text-lg font-semibold">{field.value ?? "—"}</CardContent>
          </Card>
        ))}
      </div>
    </div>
  );
}
