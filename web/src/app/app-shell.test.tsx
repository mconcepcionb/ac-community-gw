import { QueryClientProvider } from "@tanstack/react-query";
import { createMemoryHistory, createRouter, RouterProvider } from "@tanstack/react-router";
import { render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";

import { routeTree } from "@/routeTree.gen";
import { createQueryClient } from "./query-client";

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

describe("app shell", () => {
  it("renders the portal layout at /", async () => {
    renderAt("/");

    expect(await screen.findByText("Community users")).toBeInTheDocument();
    expect(screen.getByRole("link", { name: "ac-community-gw" })).toBeInTheDocument();
  });

  it("renders the console layout at /admin", async () => {
    renderAt("/admin");

    expect(await screen.findByText("Operations console")).toBeInTheDocument();
    expect(screen.getByText("console")).toBeInTheDocument();
  });
});
