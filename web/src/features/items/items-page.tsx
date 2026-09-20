import { getRouteApi, Link } from "@tanstack/react-router";
import type { ColumnDef } from "@tanstack/react-table";
import { useEffect, useState } from "react";

import type { AzerothItem } from "@/api";
import { DataTable } from "@/components/common/data-table";
import { PageHeader } from "@/components/common/page-header";
import { Pagination } from "@/components/common/pagination";
import { Input } from "@/components/ui/input";
import { useDebouncedValue } from "@/hooks/use-debounced-value";
import { useItems } from "./use-items";

const route = getRouteApi("/admin/items/");

const columns: ColumnDef<AzerothItem, unknown>[] = [
  {
    accessorKey: "entry",
    header: "Entry",
    cell: ({ row }) => (
      <Link
        to="/admin/items/$entry"
        params={{ entry: String(row.original.entry ?? 0) }}
        className="text-blue-400 underline"
      >
        {row.original.entry}
      </Link>
    ),
  },
  {
    accessorKey: "name",
    header: "Name",
    cell: ({ row }) => (
      <span style={{ color: row.original.quality_color }}>{row.original.name}</span>
    ),
  },
  { accessorKey: "quality_name", header: "Quality" },
  { accessorKey: "class_name", header: "Class" },
  { accessorKey: "subclass_name", header: "Subclass" },
  { accessorKey: "item_level", header: "iLvl" },
  { accessorKey: "required_level", header: "Req" },
];

export function ItemsPage() {
  const search = route.useSearch();
  const navigate = route.useNavigate();
  const limit = search.limit ?? 50;
  const offset = search.offset ?? 0;

  const [filterInput, setFilterInput] = useState(search.filter ?? "");
  const [classInput, setClassInput] = useState(
    search.class !== undefined ? String(search.class) : "",
  );
  const filter = useDebouncedValue(filterInput, 300);
  const classValue = useDebouncedValue(classInput, 300);
  const classId = classValue === "" ? undefined : Number(classValue);

  useEffect(() => {
    const nextClass = Number.isFinite(classId) ? classId : undefined;
    if ((search.filter ?? "") === filter && search.class === nextClass) {
      return;
    }
    navigate({
      search: {
        ...search,
        filter: filter || undefined,
        class: nextClass,
        offset: undefined,
      },
      replace: true,
    });
  }, [filter, classId, search, navigate]);

  const query = useItems({ filter: search.filter, classId: search.class, limit, offset });
  const items = query.data?.items ?? [];

  return (
    <div className="mx-auto max-w-5xl p-8">
      <PageHeader title="Items" description="AzerothCore item catalog." />

      <div className="mb-4 flex flex-wrap gap-3">
        <Input
          className="max-w-xs"
          placeholder="Filter by name"
          aria-label="Filter by name"
          value={filterInput}
          onChange={(event) => setFilterInput(event.target.value)}
        />
        <Input
          className="max-w-[8rem]"
          type="number"
          placeholder="Class id"
          aria-label="Class id"
          value={classInput}
          onChange={(event) => setClassInput(event.target.value)}
        />
      </div>

      <DataTable
        columns={columns}
        data={items}
        loading={query.isPending}
        error={query.isError ? query.error : undefined}
        onRetry={() => void query.refetch()}
        emptyMessage="No items"
      />

      <Pagination
        limit={limit}
        offset={offset}
        count={items.length}
        itemLabel="items"
        onOffsetChange={(next) =>
          navigate({ search: { ...search, offset: next === 0 ? undefined : next } })
        }
      />
    </div>
  );
}
