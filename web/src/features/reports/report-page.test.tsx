import { QueryClientProvider } from "@tanstack/react-query";
import { createMemoryHistory, createRouter, RouterProvider } from "@tanstack/react-router";
import { render, screen } from "@testing-library/react";
import { HttpResponse, http } from "msw";
import { describe, expect, it } from "vitest";

import { createQueryClient } from "@/app/query-client";
import { routeTree } from "@/routeTree.gen";
import { server } from "@/test/msw";

const meUrl = "http://localhost:8080/api/v1/me";
const mineUrl = "http://localhost:8080/api/v1/reports/mine";

function renderAt(path: string) {
  const queryClient = createQueryClient();
  const router = createRouter({
    routeTree,
    context: { queryClient },
    history: createMemoryHistory({ initialEntries: [path] }),
  });

  render(
    <QueryClientProvider client={queryClient}>
      <RouterProvider router={router} />
    </QueryClientProvider>,
  );
}

describe("ReportPage", () => {
  it("renders the report form and my reports", async () => {
    server.use(
      http.get(meUrl, () =>
        HttpResponse.json({
          user_id: "u1",
          discord_id: "1",
          roles: [],
          permissions: ["report.create"],
        }),
      ),
      http.get(mineUrl, () =>
        HttpResponse.json({
          reports: [
            {
              id: "11111111-1111-1111-1111-111111111111",
              reporter_id: "u1",
              target: "Thrall",
              category: "abuse",
              message: "spamming",
              status: "open",
              created_at: "2024-01-02T03:04:05Z",
            },
          ],
        }),
      ),
    );

    renderAt("/report");

    expect(await screen.findByText("New report")).toBeInTheDocument();
    expect(await screen.findByText("spamming")).toBeInTheDocument();
  });
});
