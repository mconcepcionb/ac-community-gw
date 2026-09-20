interface LoadingStateProps {
  label?: string;
}

export function LoadingState({ label = "Loading…" }: LoadingStateProps) {
  return <div className="p-8 text-center text-sm text-muted-foreground">{label}</div>;
}
