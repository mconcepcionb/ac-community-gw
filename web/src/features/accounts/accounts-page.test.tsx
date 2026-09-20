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

let createBody: unknown;

function renderPage() {
  const queryClient = createQueryClient();
  const router = createRouter({
    routeTree,
    context: { queryClient },
    history: createMemoryHistory({ initialEntries: ["/accounts"] }),
  });

  render(
    <QueryClientProvider client={queryClient}>
      <RouterProvider router={router} />
    </QueryClientProvider>,
  );
}

describe("AccountsPage", () => {
  beforeEach(() => {
    createBody = undefined;
  });
  it("lists accounts and creates one", async () => {
    server.use(
      http.get(meUrl, () =>
        HttpResponse.json({
          user_id: "u1",
          discord_id: "1",
          roles: [],
          permissions: ["azeroth.account.list", "azeroth.account.manage"],
        }),
      ),
      http.get(accountsUrl, () =>
        HttpResponse.json({
          accounts: [
            {
              id: 1,
              username: "ADMIN",
              email: "admin@example.test",
              gm_level: 3,
              online: false,
              banned: false,
              last_login: null,
            },
          ],
        }),
      ),
      http.post(accountsUrl, async ({ request }) => {
        createBody = await request.json();
        return HttpResponse.json({ result: "Account created" });
      }),
    );

    renderPage();
    expect(await screen.findByText("ADMIN")).toBeInTheDocument();

    const user = userEvent.setup();
    await user.click(screen.getByRole("button", { name: "Create account" }));
    await user.type(await screen.findByLabelText("Username"), "NEWUSER");
    await user.type(screen.getByLabelText("Password"), "secret");
    await user.click(screen.getByRole("button", { name: "Create" }));

    await waitFor(() => expect(createBody).toBeDefined());
    expect(createBody).toMatchObject({ username: "NEWUSER", password: "secret" });
  });
});
