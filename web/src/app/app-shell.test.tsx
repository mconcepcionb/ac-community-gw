import { QueryClientProvider } from "@tanstack/react-query";
import { createMemoryHistory, createRouter, RouterProvider } from "@tanstack/react-router";
import { render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";

import { routeTree } from "@/routeTree.gen";
import { createQueryClient } from "./query-client";

describe("app shell", () => {
  it("renders the root layout with header, outlet and footer at /", async () => {
    const queryClient = createQueryClient();
    const router = createRouter({
      routeTree,
      context: { queryClient },
      history: createMemoryHistory({ initialEntries: ["/"] }),
    });

    render(
      <QueryClientProvider client={queryClient}>
        <RouterProvider router={router} />
      </QueryClientProvider>,
    );

    expect(await screen.findByText("Community users")).toBeInTheDocument();
    expect(screen.getByRole("link", { name: "ac-community-gw" })).toBeInTheDocument();
  });
});
