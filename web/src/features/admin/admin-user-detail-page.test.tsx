import { QueryClientProvider } from "@tanstack/react-query";
import { createMemoryHistory, createRouter, RouterProvider } from "@tanstack/react-router";
import { render, screen } from "@testing-library/react";
import { HttpResponse, http } from "msw";
import { describe, expect, it } from "vitest";

import { createQueryClient } from "@/app/query-client";
import { routeTree } from "@/routeTree.gen";
import { server } from "@/test/msw";

const meUrl = "http://localhost:8080/api/v1/me";
const userId = "11111111-1111-1111-1111-111111111111";

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

describe("AdminUserDetailPage", () => {
  it("renders the 360 view", async () => {
    server.use(
      http.get(meUrl, () =>
        HttpResponse.json({
          user_id: "u1",
          discord_id: "1",
          roles: [],
          permissions: ["azeroth.admin.users.read"],
        }),
      ),
      http.get(`http://localhost:8080/api/v1/admin/users/${userId}`, () =>
        HttpResponse.json({
          user_id: userId,
          discord_id: "123",
          username: "alice",
          display_name: "Alice",
          created_at: "2024-01-02T03:04:05Z",
          roles: ["member"],
          account_username: "ADMIN",
          account_id: 42,
          wallet: 250,
          characters: [
            {
              guid: 7,
              name: "Thrall",
              level: 80,
              race: 2,
              class: 7,
              guild: "Horde",
              online: true,
              money: 100,
            },
          ],
          orders: [
            {
              order_id: "22222222-2222-2222-2222-222222222222",
              sku: "starter",
              points: 100,
              character: "Thrall",
              status: "delivered",
              created_at: "2024-02-02T03:04:05Z",
            },
          ],
        }),
      ),
    );

    renderAt(`/admin/users/${userId}`);

    expect(await screen.findByText("member")).toBeInTheDocument();
    expect(screen.getAllByText("Thrall").length).toBeGreaterThan(0);
    expect(screen.getByText(/ADMIN/)).toBeInTheDocument();
    expect(screen.getByText("250 points")).toBeInTheDocument();
    expect(screen.getByText("starter")).toBeInTheDocument();
  });
});
