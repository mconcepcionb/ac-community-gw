import type { ReactNode } from "react";

import { Badge } from "@/components/ui/badge";
import { cn } from "@/lib/utils";

export type StatusTone = "positive" | "negative" | "warning" | "neutral";

const toneClasses: Record<StatusTone, string> = {
  positive: "border-emerald-500/30 bg-emerald-500/10 text-emerald-600 dark:text-emerald-400",
  negative: "border-destructive/40 bg-destructive/10 text-destructive",
  warning: "border-amber-500/30 bg-amber-500/10 text-amber-600 dark:text-amber-400",
  neutral: "border-border bg-muted text-muted-foreground",
};

interface StatusBadgeProps {
  tone?: StatusTone;
  children: ReactNode;
  className?: string;
}

/** StatusBadge renders a state with colour and text, never colour alone. */
export function StatusBadge({ tone = "neutral", children, className }: StatusBadgeProps) {
  return (
    <Badge variant="outline" className={cn(toneClasses[tone], className)}>
      {children}
    </Badge>
  );
}
