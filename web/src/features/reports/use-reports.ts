import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";

import { reportsCreateMutation, reportsMineOptions, reportsMineQueryKey } from "@/api";

/** useMyReports lists the signed-in user's own reports. */
export function useMyReports() {
  return useQuery(reportsMineOptions());
}

/** useSubmitReport submits a report and refreshes the list. */
export function useSubmitReport() {
  const queryClient = useQueryClient();
  return useMutation({
    ...reportsCreateMutation(),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: reportsMineQueryKey() }),
  });
}
