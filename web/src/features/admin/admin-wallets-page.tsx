import { PageHeader } from "@/components/common/page-header";
import { Button } from "@/components/ui/button";
import { GrantDialog } from "@/features/store/grant-dialog";

/** AdminWalletsPage grants points to a community user's wallet. */
export function AdminWalletsPage() {
  return (
    <div className="mx-auto max-w-5xl p-8">
      <PageHeader
        title="Wallets"
        description="Grant points to a community user's wallet."
        actions={<GrantDialog trigger={<Button>Grant points</Button>} />}
      />
      <p className="max-w-2xl text-sm text-muted-foreground">
        Enter a community user id or Discord id to add points. For a balance and order history, open
        the user's 360 view.
      </p>
    </div>
  );
}
