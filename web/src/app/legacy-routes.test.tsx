import { QueryClientProvider } from "@tanstack/react-query";
import { createMemoryHistory, createRouter, RouterProvider } from "@tanstack/react-router";
import { render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";

import { routeTree } from "@/routeTree.gen";
import { NotFound } from "./not-found";
import { createQueryClient } from "./query-client";

const retiredPaths = [
  "/accounts",
  "/account-links",
  "/items",
  "/identity/users",
  "/azeroth/status",
  "/store/products",
  "/store/wallet",
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

describe("retired routes", () => {
  it.each(retiredPaths)("%s no longer resolves", async (path) => {
    renderAt(path);

    expect(await screen.findByText("Page not found")).toBeInTheDocument();
  });
});
