interface EmptyStateProps {
  title?: string;
  description?: string;
}

export function EmptyState({ title = "Nothing here yet", description }: EmptyStateProps) {
  return (
    <div className="rounded-md border border-dashed border-border p-8 text-center">
      <p className="text-sm font-medium">{title}</p>
      {description ? <p className="mt-1 text-sm text-muted-foreground">{description}</p> : null}
    </div>
  );
}
