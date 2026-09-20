import { QueryClientProvider } from "@tanstack/react-query";
import { createMemoryHistory, createRouter, RouterProvider } from "@tanstack/react-router";
import { render, screen } from "@testing-library/react";
import { HttpResponse, http } from "msw";
import { describe, expect, it } from "vitest";

import { routeTree } from "@/routeTree.gen";
import { server } from "@/test/msw";
import { NotFound } from "./not-found";
import { createQueryClient } from "./query-client";

const meUrl = "http://localhost:8080/api/v1/me";

const retiredPaths = [
  "/accounts",
  "/account-links",
  "/items",
  "/identity/users",
  "/azeroth/status",
  "/store/products",
  "/store/wallet",
];

const retiredConsolePaths = [
  "/admin/accounts",
  "/admin/characters",
  "/admin/items",
  "/admin/online",
];

function renderAt(path: string) {
  const queryClient = createQueryClient();
  const router = createRouter({
    routeTree,
    context: { queryClient },
    history: createMemoryHistory({ initialEntries: [path] }),
    defaultNotFoundComponent: NotFound,
  });

  render(
    <QueryClientProvider client={queryClient}>
      <RouterProvider router={router} />
    </QueryClientProvider>,
  );
}

function mockStaffSession() {
  server.use(
    http.get(meUrl, () =>
      HttpResponse.json({
        user_id: "u1",
        discord_id: "1",
        roles: ["admin"],
        permissions: ["azeroth.admin.players.read"],
      }),
    ),
  );
}

describe("retired routes", () => {
  it.each(retiredPaths)("%s no longer resolves", async (path) => {
    renderAt(path);

    expect(await screen.findByText("Page not found")).toBeInTheDocument();
  });
});

describe("retired console routes", () => {
  it.each(retiredConsolePaths)("%s no longer resolves", async (path) => {
    mockStaffSession();
    renderAt(path);

    expect(await screen.findByText("Page not found")).toBeInTheDocument();
  });
});
