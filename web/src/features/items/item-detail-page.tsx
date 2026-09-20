import { ErrorState } from "@/components/common/error-state";
import { LoadingState } from "@/components/common/loading-state";
import { PageHeader } from "@/components/common/page-header";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { isValidEntry, useItem } from "./use-items";

export function ItemDetailPage({ entry }: { entry: number }) {
  const query = useItem(entry);

  if (!isValidEntry(entry)) {
    return (
      <div className="mx-auto max-w-5xl p-8">
        <PageHeader title="Item" description={String(entry)} />
        <p className="text-sm text-muted-foreground">Invalid item id.</p>
      </div>
    );
  }

  if (query.isPending) {
    return <LoadingState label="Loading item…" />;
  }

  if (query.isError) {
    return (
      <div className="mx-auto max-w-5xl p-8">
        <PageHeader title="Item" description={String(entry)} />
        <ErrorState error={query.error} onRetry={() => void query.refetch()} />
      </div>
    );
  }

  const item = query.data;
  const fields = [
    { label: "Entry", value: item?.entry },
    { label: "Quality", value: item?.quality_name },
    { label: "Class", value: item?.class_name },
    { label: "Subclass", value: item?.subclass_name },
    { label: "Inventory", value: item?.inventory_type_name },
    { label: "Item level", value: item?.item_level },
    { label: "Required level", value: item?.required_level },
    { label: "Binding", value: item?.bonding_name },
    { label: "Armor", value: item?.armor },
    { label: "Buy price", value: item?.buy_price },
    { label: "Sell price", value: item?.sell_price },
  ];

  return (
    <div className="mx-auto max-w-5xl p-8">
      <PageHeader
        title={item?.name ?? `Item ${entry}`}
        description={item?.description || "Item detail."}
      />

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

      {item?.stats && item.stats.length > 0 ? (
        <section className="mt-8">
          <h2 className="mb-3 text-lg font-semibold">Stats</h2>
          <ul className="space-y-1 text-sm">
            {item.stats.map((stat) => (
              <li key={`${stat.type}-${stat.name}`}>
                {stat.name}: <span className="font-medium">{stat.value}</span>
              </li>
            ))}
          </ul>
        </section>
      ) : null}

      {item?.damage && item.damage.length > 0 ? (
        <section className="mt-8">
          <h2 className="mb-3 text-lg font-semibold">Damage</h2>
          <ul className="space-y-1 text-sm">
            {item.damage.map((damage) => (
              <li key={`${damage.type}-${damage.min}`}>
                {damage.type_name}: {damage.min}–{damage.max} ({(damage.dps ?? 0).toFixed(1)} dps)
              </li>
            ))}
          </ul>
        </section>
      ) : null}

      {item?.spells && item.spells.length > 0 ? (
        <section className="mt-8">
          <h2 className="mb-3 text-lg font-semibold">Spells</h2>
          <ul className="space-y-1 text-sm">
            {item.spells.map((spell) => (
              <li key={spell.id}>
                {spell.trigger_name}: spell {spell.id}
              </li>
            ))}
          </ul>
        </section>
      ) : null}
    </div>
  );
}
