import type { ReactNode } from "react";

import { cn } from "@/lib/utils";

interface RowActionsProps {
  children: ReactNode;
  className?: string;
}

/** RowActions lays out per-row actions with consistent spacing and wrapping. */
export function RowActions({ children, className }: RowActionsProps) {
  return <div className={cn("flex flex-wrap items-center gap-2", className)}>{children}</div>;
}
