import { QueryClientProvider } from "@tanstack/react-query";
import { createMemoryHistory, createRouter, RouterProvider } from "@tanstack/react-router";
import { render, screen } from "@testing-library/react";
import { HttpResponse, http } from "msw";
import { describe, expect, it } from "vitest";

import { createQueryClient } from "@/app/query-client";
import { routeTree } from "@/routeTree.gen";
import { server } from "@/test/msw";

const meUrl = "http://localhost:8080/api/v1/me";
const usersUrl = "http://localhost:8080/api/v1/identity/users";

function authorized(permissions: string[]) {
  return http.get(meUrl, () =>
    HttpResponse.json({ user_id: "u1", discord_id: "1", roles: [], permissions }),
  );
}

function renderUsers() {
  const queryClient = createQueryClient();
  const router = createRouter({
    routeTree,
    context: { queryClient },
    history: createMemoryHistory({ initialEntries: ["/admin/users"] }),
  });

  render(
    <QueryClientProvider client={queryClient}>
      <RouterProvider router={router} />
    </QueryClientProvider>,
  );
}

describe("UsersPage", () => {
  it("lists users when authorized", async () => {
    server.use(
      authorized(["gw.identity.user.read"]),
      http.get(usersUrl, () =>
        HttpResponse.json({
          users: [
            {
              user_id: "11111111-1111-1111-1111-111111111111",
              discord_id: "42",
              username: "thrall",
              global_name: "Thrall",
              display_name: "Thrall#42",
              created_at: "2024-01-02T03:04:05Z",
            },
          ],
        }),
      ),
    );

    renderUsers();
    expect(await screen.findByText("thrall")).toBeInTheDocument();
  });

  it("shows the empty state", async () => {
    server.use(
      authorized(["gw.identity.user.read"]),
      http.get(usersUrl, () => HttpResponse.json({ users: [] })),
    );

    renderUsers();
    expect(await screen.findByText("No users")).toBeInTheDocument();
  });
});
