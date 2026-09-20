import { QueryClientProvider } from "@tanstack/react-query";
import { createMemoryHistory, createRouter, RouterProvider } from "@tanstack/react-router";
import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { HttpResponse, http } from "msw";
import { beforeEach, describe, expect, it } from "vitest";

import { createQueryClient } from "@/app/query-client";
import { routeTree } from "@/routeTree.gen";
import { server } from "@/test/msw";

const meUrl = "http://localhost:8080/api/v1/me";
const rolesUrl = "http://localhost:8080/api/v1/admin/roles";
const replaceUrl = "http://localhost:8080/api/v1/admin/roles/moderator/permissions";

let replaceBody: unknown;

function renderPage() {
  const queryClient = createQueryClient();
  const router = createRouter({
    routeTree,
    context: { queryClient },
    history: createMemoryHistory({ initialEntries: ["/admin/roles"] }),
  });

  render(
    <QueryClientProvider client={queryClient}>
      <RouterProvider router={router} />
    </QueryClientProvider>,
  );
}

describe("AdminRolesPage", () => {
  beforeEach(() => {
    replaceBody = undefined;
  });

  it("saves a role's permission matrix", async () => {
    server.use(
      http.get(meUrl, () =>
        HttpResponse.json({
          user_id: "u1",
          discord_id: "1",
          roles: [],
          permissions: ["gw.identity.roles.manage"],
        }),
      ),
      http.get(rolesUrl, () =>
        HttpResponse.json({
          roles: ["moderator"],
          permissions: [
            { name: "gw.report.create", owner: "reports" },
            { name: "gw.report.read", owner: "reports" },
          ],
          grants: [{ role: "moderator", permission: "gw.report.create" }],
          mappings: [],
        }),
      ),
      http.put(replaceUrl, async ({ request }) => {
        replaceBody = await request.json();
        return new HttpResponse(null, { status: 204 });
      }),
    );

    renderPage();
    expect(await screen.findByText("moderator")).toBeInTheDocument();

    const user = userEvent.setup();
    await user.click(screen.getByRole("checkbox", { name: "gw.report.read" }));
    await user.click(screen.getByRole("button", { name: "Save" }));

    await waitFor(() => expect(replaceBody).toBeDefined());
    expect(replaceBody).toMatchObject({
      permissions: expect.arrayContaining(["gw.report.create", "gw.report.read"]),
    });
  });
});
