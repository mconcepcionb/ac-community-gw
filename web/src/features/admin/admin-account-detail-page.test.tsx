import { QueryClientProvider } from "@tanstack/react-query";
import { createMemoryHistory, createRouter, RouterProvider } from "@tanstack/react-router";
import { render, screen } from "@testing-library/react";
import { HttpResponse, http } from "msw";
import { describe, expect, it } from "vitest";

import { createQueryClient } from "@/app/query-client";
import { routeTree } from "@/routeTree.gen";
import { server } from "@/test/msw";

const meUrl = "http://localhost:8080/api/v1/me";
const accountUrl = "http://localhost:8080/api/v1/azeroth/accounts/ADMIN";
const charactersUrl = "http://localhost:8080/api/v1/azeroth/characters";
const annotationsUrl = "http://localhost:8080/api/v1/admin/annotations";
const auditUrl = "http://localhost:8080/api/v1/admin/audit";

function renderDetail() {
  const queryClient = createQueryClient();
  const router = createRouter({
    routeTree,
    context: { queryClient },
    history: createMemoryHistory({ initialEntries: ["/admin/accounts/ADMIN"] }),
  });

  render(
    <QueryClientProvider client={queryClient}>
      <RouterProvider router={router} />
    </QueryClientProvider>,
  );
}

describe("AdminAccountDetailPage", () => {
  it("renders the account, its owner and characters", async () => {
    server.use(
      http.get(meUrl, () =>
        HttpResponse.json({
          user_id: "u1",
          discord_id: "1",
          roles: [],
          permissions: ["azeroth.account.list", "gw.notes.read", "gw.audit.read"],
        }),
      ),
      http.get(accountUrl, () =>
        HttpResponse.json({
          account: {
            id: 1,
            username: "ADMIN",
            email: "admin@example.test",
            gm_level: 3,
            expansion: 2,
            online: true,
            banned: false,
            last_ip: "127.0.0.1",
            last_login: null,
            claimed: true,
            claimed_by: {
              user_id: "11111111-1111-1111-1111-111111111111",
              display_name: "Thrall",
              discord_id: "42",
            },
          },
        }),
      ),
      http.get(charactersUrl, () =>
        HttpResponse.json({
          characters: [
            {
              guid: 7,
              name: "Thrall",
              level: 80,
              class_name: "Shaman",
              race_name: "Orc",
              guild: "Horde",
              online: true,
            },
          ],
        }),
      ),
      http.get(annotationsUrl, () => HttpResponse.json({ annotations: [] })),
      http.get(auditUrl, () => HttpResponse.json({ entries: [] })),
    );

    renderDetail();

    expect(await screen.findByRole("heading", { name: "ADMIN" })).toBeInTheDocument();
    expect(screen.getByText("admin@example.test")).toBeInTheDocument();
    expect(await screen.findByRole("link", { name: "View user" })).toBeInTheDocument();
    expect(await screen.findByRole("link", { name: "Thrall" })).toBeInTheDocument();
  });
});
