import { Link } from "@tanstack/react-router";

import { Button } from "@/components/ui/button";

/** ForbiddenState is shown when the principal lacks access to an area. */
export function ForbiddenState() {
  return (
    <div className="mx-auto max-w-3xl p-8 text-center">
      <h1 className="text-xl font-semibold">Access denied</h1>
      <p className="mt-2 text-sm text-muted-foreground">
        You do not have permission to view this area.
      </p>
      <Button asChild variant="outline" size="sm" className="mt-4">
        <Link to="/">Back to the portal</Link>
      </Button>
    </div>
  );
}
