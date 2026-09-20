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
const accountsUrl = "http://localhost:8080/api/v1/azeroth/accounts";
const banUrl = "http://localhost:8080/api/v1/azeroth/accounts/ADMIN/ban";

let banBody: unknown;

function renderPage() {
  const queryClient = createQueryClient();
  const router = createRouter({
    routeTree,
    context: { queryClient },
    history: createMemoryHistory({ initialEntries: ["/admin/accounts"] }),
  });

  render(
    <QueryClientProvider client={queryClient}>
      <RouterProvider router={router} />
    </QueryClientProvider>,
  );
}

describe("AdminAccountsPage", () => {
  beforeEach(() => {
    banBody = undefined;
  });
  it("lists accounts and bans one", async () => {
    server.use(
      http.get(meUrl, () =>
        HttpResponse.json({
          user_id: "u1",
          discord_id: "1",
          roles: [],
          permissions: ["azeroth.account.list", "azeroth.admin.accounts.ban"],
        }),
      ),
      http.get(accountsUrl, () =>
        HttpResponse.json({
          accounts: [{ id: 1, username: "ADMIN", gm_level: 3, online: false, banned: false }],
        }),
      ),
      http.post(banUrl, async ({ request }) => {
        banBody = await request.json();
        return HttpResponse.json({ result: "banned" });
      }),
    );

    renderPage();
    expect(await screen.findByText("ADMIN")).toBeInTheDocument();

    const user = userEvent.setup();
    await user.click(screen.getByRole("button", { name: "Ban" }));
    await user.type(await screen.findByLabelText("Reason"), "cheating");
    await user.click(screen.getByRole("button", { name: "Confirm ban" }));

    await waitFor(() => expect(banBody).toBeDefined());
    expect(banBody).toMatchObject({ username: "ADMIN", duration: "1d", reason: "cheating" });
  });
});
