import { PageHeader } from "@/components/common/page-header";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { useSession } from "@/features/auth/use-session";

function avatarUrl(discordId?: string, avatar?: string): string | undefined {
  if (!discordId || !avatar) {
    return undefined;
  }
  return `https://cdn.discordapp.com/avatars/${discordId}/${avatar}.png?size=128`;
}

function formatDate(value?: string): string {
  if (!value) {
    return "—";
  }
  const date = new Date(value);
  return Number.isNaN(date.getTime()) ? value : date.toLocaleString();
}

/** ProfilePage shows every piece of information known about the signed-in user. */
export function ProfilePage() {
  const { principal } = useSession();

  const name =
    principal?.display_name ||
    principal?.global_name ||
    principal?.username ||
    principal?.discord_id ||
    "—";
  const avatar = avatarUrl(principal?.discord_id, principal?.avatar);
  const roles = principal?.roles ?? [];
  const permissions = principal?.permissions ?? [];

  const fields = [
    { label: "Display name", value: principal?.display_name },
    { label: "Global name", value: principal?.global_name },
    { label: "Username", value: principal?.username },
    { label: "Discord ID", value: principal?.discord_id },
    { label: "Community user ID", value: principal?.user_id },
    { label: "Member since", value: formatDate(principal?.created_at) },
  ];

  return (
    <div className="mx-auto max-w-5xl p-8">
      <PageHeader title="Profile" description="Your account information." />

      <div className="mb-8 flex items-center gap-4">
        {avatar ? (
          <img src={avatar} alt="" className="h-16 w-16 rounded-full" />
        ) : (
          <div className="h-16 w-16 rounded-full bg-neutral-800" />
        )}
        <div>
          <p className="text-lg font-semibold">{name}</p>
          <p className="text-sm text-muted-foreground">{principal?.discord_id}</p>
        </div>
      </div>

      <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
        {fields.map((field) => (
          <Card key={field.label}>
            <CardHeader>
              <CardTitle className="text-sm font-medium text-muted-foreground">
                {field.label}
              </CardTitle>
            </CardHeader>
            <CardContent className="break-all text-lg font-semibold">
              {field.value || "—"}
            </CardContent>
          </Card>
        ))}
      </div>

      <section className="mt-8">
        <h2 className="mb-3 text-lg font-semibold">Roles</h2>
        {roles.length === 0 ? (
          <p className="text-sm text-muted-foreground">No roles.</p>
        ) : (
          <ul className="flex flex-wrap gap-2">
            {roles.map((role) => (
              <li key={role} className="rounded border border-border px-2 py-1 text-xs">
                {role}
              </li>
            ))}
          </ul>
        )}
      </section>

      <section className="mt-8">
        <h2 className="mb-3 text-lg font-semibold">Permissions</h2>
        {permissions.length === 0 ? (
          <p className="text-sm text-muted-foreground">No permissions.</p>
        ) : (
          <ul className="flex flex-wrap gap-2">
            {permissions.map((permission) => (
              <li key={permission} className="rounded border border-border px-2 py-1 text-xs">
                {permission}
              </li>
            ))}
          </ul>
        )}
      </section>
    </div>
  );
}
