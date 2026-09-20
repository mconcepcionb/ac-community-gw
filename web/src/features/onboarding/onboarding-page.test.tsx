import { QueryClientProvider } from "@tanstack/react-query";
import { createMemoryHistory, createRouter, RouterProvider } from "@tanstack/react-router";
import { render, screen } from "@testing-library/react";
import { HttpResponse, http } from "msw";
import { describe, expect, it } from "vitest";

import { createQueryClient } from "@/app/query-client";
import { routeTree } from "@/routeTree.gen";
import { server } from "@/test/msw";

const meUrl = "http://localhost:8080/api/v1/me";
const accountUrl = "http://localhost:8080/api/v1/azeroth/me/account";

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

describe("OnboardingPage", () => {
  it("offers to create an account when none is linked", async () => {
    server.use(
      http.get(meUrl, () =>
        HttpResponse.json({
          user_id: "u1",
          discord_id: "1",
          roles: [],
          permissions: ["azeroth.account.self"],
        }),
      ),
      http.get(accountUrl, () => HttpResponse.json({ linked: false })),
    );

    renderAt("/onboarding");

    expect(await screen.findByText("Create a new account")).toBeInTheDocument();
    expect(screen.getByText("I already have an account")).toBeInTheDocument();
  });

  it("shows the linked account", async () => {
    server.use(
      http.get(meUrl, () =>
        HttpResponse.json({
          user_id: "u1",
          discord_id: "1",
          roles: [],
          permissions: ["azeroth.account.self"],
        }),
      ),
      http.get(accountUrl, () =>
        HttpResponse.json({ linked: true, account_username: "ADMIN", account_id: 42 }),
      ),
    );

    renderAt("/onboarding");

    expect(await screen.findByText("Account linked")).toBeInTheDocument();
  });
});
