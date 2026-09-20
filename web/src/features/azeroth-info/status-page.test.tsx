import { QueryClientProvider } from "@tanstack/react-query";
import { createMemoryHistory, createRouter, RouterProvider } from "@tanstack/react-router";
import { render, screen } from "@testing-library/react";
import { HttpResponse, http } from "msw";
import { describe, expect, it } from "vitest";

import { createQueryClient } from "@/app/query-client";
import { routeTree } from "@/routeTree.gen";
import { server } from "@/test/msw";

const meUrl = "http://localhost:8080/api/v1/me";
const statusUrl = "http://localhost:8080/api/v1/public/status";

function authorized() {
  return http.get(meUrl, () =>
    HttpResponse.json({
      user_id: "u1",
      discord_id: "1",
      roles: [],
      permissions: ["azeroth.info.public.read"],
    }),
  );
}

function renderStatus() {
  const queryClient = createQueryClient();
  const router = createRouter({
    routeTree,
    context: { queryClient },
    history: createMemoryHistory({ initialEntries: ["/status"] }),
  });

  render(
    <QueryClientProvider client={queryClient}>
      <RouterProvider router={router} />
    </QueryClientProvider>,
  );
}

describe("StatusPage", () => {
  it("renders the status metrics", async () => {
    server.use(
      authorized(),
      http.get(statusUrl, () =>
        HttpResponse.json({
          output: "AzerothCore rev. fake",
          version: "AzerothCore rev. fake",
          connected_players: 3,
          characters_in_world: 5,
          connection_peak: 10,
          queue: 0,
          uptime: "1 Hour(s)",
        }),
      ),
    );

    renderStatus();
    expect(await screen.findByText("Connected players")).toBeInTheDocument();
    expect(screen.getByText("3")).toBeInTheDocument();
  });

  it("shows an error state when the upstream fails", async () => {
    server.use(
      authorized(),
      http.get(statusUrl, () =>
        HttpResponse.json(
          { error: { code: "bad_gateway", message: "upstream integration failed" } },
          { status: 502 },
        ),
      ),
    );

    renderStatus();
    expect(await screen.findByRole("alert")).toBeInTheDocument();
  });
});
