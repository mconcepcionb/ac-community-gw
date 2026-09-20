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
let createBody: unknown;

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

function mockSession(permissions: string[]) {
  server.use(
    http.get(meUrl, () =>
      HttpResponse.json({ user_id: "u1", discord_id: "1", roles: [], permissions }),
    ),
  );
}

function mockAccounts() {
  server.use(
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
  );
}

describe("AdminAccountsPage", () => {
  beforeEach(() => {
    banBody = undefined;
    createBody = undefined;
  });

  it("lists accounts and creates one", async () => {
    mockSession(["azeroth.account.list", "azeroth.account.manage"]);
    mockAccounts();
    server.use(
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

  it("bans an account", async () => {
    mockSession(["azeroth.account.list", "azeroth.admin.accounts.ban"]);
    mockAccounts();
    server.use(
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

  it("shows only Unban for a banned account", async () => {
    mockSession(["azeroth.account.list", "azeroth.admin.accounts.ban"]);
    server.use(
      http.get(accountsUrl, () =>
        HttpResponse.json({
          accounts: [
            {
              id: 2,
              username: "BAD",
              email: "bad@example.test",
              gm_level: 0,
              online: false,
              banned: true,
              ban_reason: "cheating",
              last_login: null,
            },
          ],
        }),
      ),
    );

    renderPage();
    expect(await screen.findByRole("button", { name: "Unban" })).toBeInTheDocument();
    expect(screen.queryByRole("button", { name: "Ban" })).not.toBeInTheDocument();
    expect(screen.getByText("cheating")).toBeInTheDocument();
  });

  it("reflects the current GM level in the control", async () => {
    mockSession(["azeroth.account.list", "azeroth.admin.accounts.gmlevel"]);
    mockAccounts();

    renderPage();
    expect(await screen.findByRole("button", { name: "3 — Administrator" })).toBeInTheDocument();
  });
});
