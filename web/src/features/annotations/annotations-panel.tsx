import { useState } from "react";
import { toast } from "sonner";

import type { AdminAnnotation } from "@/api";
import { isApiError } from "@/api/errors";
import { ConfirmDialog } from "@/components/common/confirm-dialog";
import { EmptyState } from "@/components/common/empty-state";
import { ErrorState } from "@/components/common/error-state";
import { LoadingState } from "@/components/common/loading-state";
import { PermissionGate } from "@/components/common/permission-gate";
import { Button } from "@/components/ui/button";
import { Textarea } from "@/components/ui/textarea";
import {
  useAnnotations,
  useCreateAnnotation,
  useDeleteAnnotation,
  useUpdateAnnotation,
} from "./use-annotations";

const report = (error: unknown) =>
  toast.error(isApiError(error) ? `${error.message} (${error.code})` : String(error));

function AnnotationItem({
  annotation,
  targetType,
  targetId,
}: {
  annotation: AdminAnnotation;
  targetType: string;
  targetId: string;
}) {
  const [editing, setEditing] = useState(false);
  const [body, setBody] = useState(annotation.body ?? "");
  const update = useUpdateAnnotation(targetType, targetId);
  const remove = useDeleteAnnotation(targetType, targetId);

  const save = async () => {
    if (!annotation.id) {
      return;
    }
    try {
      await update.mutateAsync({ path: { id: annotation.id }, body: { body } });
      toast.success("Annotation updated");
      setEditing(false);
    } catch (error) {
      report(error);
    }
  };

  const onDelete = async () => {
    if (!annotation.id) {
      return true;
    }
    try {
      await remove.mutateAsync({ path: { id: annotation.id } });
      toast.success("Annotation deleted");
    } catch (error) {
      report(error);
    }
    return true;
  };

  return (
    <li className="rounded-md border border-border p-3">
      {editing ? (
        <div className="space-y-2">
          <Textarea
            aria-label="Edit annotation"
            value={body}
            onChange={(event) => setBody(event.target.value)}
          />
          <div className="flex gap-2">
            <Button size="sm" onClick={() => void save()} disabled={update.isPending}>
              Save
            </Button>
            <Button size="sm" variant="outline" onClick={() => setEditing(false)}>
              Cancel
            </Button>
          </div>
        </div>
      ) : (
        <>
          <p className="whitespace-pre-wrap text-sm">{annotation.body}</p>
          <div className="mt-2 flex flex-wrap items-center justify-between gap-2 text-xs text-muted-foreground">
            <span>
              {annotation.author_id} · {annotation.created_at}
            </span>
            <PermissionGate permission="gw.notes.write">
              <div className="flex gap-2">
                <Button size="xs" variant="outline" onClick={() => setEditing(true)}>
                  Edit
                </Button>
                <ConfirmDialog
                  trigger={
                    <Button size="xs" variant="outline">
                      Delete
                    </Button>
                  }
                  title="Delete this annotation?"
                  description="This cannot be undone."
                  confirmLabel="Delete"
                  destructive
                  onConfirm={onDelete}
                />
              </div>
            </PermissionGate>
          </div>
        </>
      )}
    </li>
  );
}

/** AnnotationsPanel shows and manages the staff annotations of a target. */
export function AnnotationsPanel({
  targetType,
  targetId,
}: {
  targetType: string;
  targetId: string;
}) {
  const query = useAnnotations({ targetType, targetId });
  const create = useCreateAnnotation(targetType, targetId);
  const [body, setBody] = useState("");

  const annotations = query.data?.annotations ?? [];

  const add = async () => {
    try {
      await create.mutateAsync({ body: { target_type: targetType, target_id: targetId, body } });
      toast.success("Annotation added");
      setBody("");
    } catch (error) {
      report(error);
    }
  };

  return (
    <div className="space-y-3">
      <PermissionGate permission="gw.notes.write">
        <div className="space-y-2">
          <Textarea
            aria-label="New annotation"
            placeholder="Add a note for other staff…"
            value={body}
            onChange={(event) => setBody(event.target.value)}
          />
          <Button
            size="sm"
            onClick={() => void add()}
            disabled={create.isPending || body.trim() === ""}
          >
            Add note
          </Button>
        </div>
      </PermissionGate>

      {query.isPending ? <LoadingState label="Loading annotations…" /> : null}
      {query.isError ? (
        <ErrorState error={query.error} onRetry={() => void query.refetch()} />
      ) : null}
      {query.data && annotations.length === 0 ? <EmptyState title="No annotations" /> : null}
      {annotations.length > 0 ? (
        <ul className="space-y-2">
          {annotations.map((annotation) => (
            <AnnotationItem
              key={annotation.id}
              annotation={annotation}
              targetType={targetType}
              targetId={targetId}
            />
          ))}
        </ul>
      ) : null}
    </div>
  );
}
