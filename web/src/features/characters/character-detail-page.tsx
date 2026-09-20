import { ErrorState } from "@/components/common/error-state";
import { LoadingState } from "@/components/common/loading-state";
import { PageHeader } from "@/components/common/page-header";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { useCharacter } from "./use-characters";

export function CharacterDetailPage({ name }: { name: string }) {
  const query = useCharacter(name);

  if (query.isPending) {
    return <LoadingState label="Loading character…" />;
  }

  if (query.isError) {
    return (
      <div className="mx-auto max-w-5xl p-8">
        <PageHeader title="Character" description={name} />
        <ErrorState error={query.error} onRetry={() => void query.refetch()} />
      </div>
    );
  }

  const character = query.data;
  const fields = [
    { label: "GUID", value: character?.guid },
    { label: "Level", value: character?.level },
    { label: "Class", value: character?.class_name },
    { label: "Race", value: character?.race_name },
    { label: "Guild", value: character?.guild },
    { label: "Online", value: character?.online ? "yes" : "no" },
    { label: "Money", value: character?.money },
    { label: "Last logout", value: character?.logout_time ?? "never" },
  ];

  return (
    <div className="mx-auto max-w-5xl p-8">
      <PageHeader title={character?.name ?? name} description="Character detail." />

      <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
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
