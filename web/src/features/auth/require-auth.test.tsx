import { QueryClientProvider } from "@tanstack/react-query";
import {
  createMemoryHistory,
  createRootRoute,
  createRoute,
  createRouter,
  RouterProvider,
} from "@tanstack/react-router";
import { render, screen, waitFor } from "@testing-library/react";
import { HttpResponse, http } from "msw";
import { describe, expect, it } from "vitest";

import { createQueryClient } from "@/app/query-client";
import { server } from "@/test/msw";

import { RequireAuth } from "./require-auth";

const meUrl = "http://localhost:8080/api/v1/me";

function renderGuarded() {
  const rootRoute = createRootRoute();
  const indexRoute = createRoute({
    getParentRoute: () => rootRoute,
    path: "/",
    component: () => (
      <RequireAuth>
        <span>secret content</span>
      </RequireAuth>
    ),
  });
  const loginRoute = createRoute({
    getParentRoute: () => rootRoute,
    path: "/login",
    validateSearch: (search: Record<string, unknown>) => ({
      return_to: typeof search.return_to === "string" ? search.return_to : undefined,
    }),
    component: () => <span>login page</span>,
  });
  const router = createRouter({
    routeTree: rootRoute.addChildren([indexRoute, loginRoute]),
    history: createMemoryHistory({ initialEntries: ["/"] }),
  });
  const queryClient = createQueryClient();
  render(
    <QueryClientProvider client={queryClient}>
      <RouterProvider router={router} />
    </QueryClientProvider>,
  );
}

describe("RequireAuth", () => {
  it("redirects anonymous visitors to /login", async () => {
    server.use(http.get(meUrl, () => new HttpResponse(null, { status: 401 })));
    renderGuarded();
    expect(await screen.findByText("login page")).toBeInTheDocument();
  });

  it("shows an error state instead of redirecting on a server error", async () => {
    server.use(http.get(meUrl, () => new HttpResponse(null, { status: 500 })));
    renderGuarded();
    await waitFor(() => expect(screen.getByText("Something went wrong")).toBeInTheDocument());
    expect(screen.queryByText("login page")).not.toBeInTheDocument();
  });

  it("renders children for an authenticated session", async () => {
    server.use(
      http.get(meUrl, () =>
        HttpResponse.json({ user_id: "u1", discord_id: "1", roles: [], permissions: [] }),
      ),
    );
    renderGuarded();
    expect(await screen.findByText("secret content")).toBeInTheDocument();
  });
});
