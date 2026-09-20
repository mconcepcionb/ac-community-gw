import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";

import {
  adminNotesCreateMutation,
  adminNotesDeleteMutation,
  adminNotesListOptions,
  adminNotesListQueryKey,
  adminNotesUpdateMutation,
} from "@/api";

export interface AnnotationsQuery {
  targetType: string;
  targetId: string;
  limit?: number;
  offset?: number;
}

/** useAnnotations lists the staff annotations of a target, newest first. */
export function useAnnotations({ targetType, targetId, limit = 50, offset = 0 }: AnnotationsQuery) {
  return useQuery({
    ...adminNotesListOptions({
      query: { target_type: targetType, target_id: targetId, limit, offset },
    }),
    enabled: targetType !== "" && targetId !== "",
  });
}

function useInvalidate(targetType: string, targetId: string) {
  const queryClient = useQueryClient();
  return () =>
    queryClient.invalidateQueries({
      queryKey: adminNotesListQueryKey({ query: { target_type: targetType, target_id: targetId } }),
    });
}

/** useCreateAnnotation creates a staff annotation. */
export function useCreateAnnotation(targetType: string, targetId: string) {
  const invalidate = useInvalidate(targetType, targetId);
  return useMutation({ ...adminNotesCreateMutation(), onSuccess: invalidate });
}

/** useUpdateAnnotation edits a staff annotation. */
export function useUpdateAnnotation(targetType: string, targetId: string) {
  const invalidate = useInvalidate(targetType, targetId);
  return useMutation({ ...adminNotesUpdateMutation(), onSuccess: invalidate });
}

/** useDeleteAnnotation removes a staff annotation. */
export function useDeleteAnnotation(targetType: string, targetId: string) {
  const invalidate = useInvalidate(targetType, targetId);
  return useMutation({ ...adminNotesDeleteMutation(), onSuccess: invalidate });
}
