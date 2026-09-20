import { QueryClientProvider } from "@tanstack/react-query";
import { createMemoryHistory, createRouter, RouterProvider } from "@tanstack/react-router";
import { render, screen } from "@testing-library/react";
import { HttpResponse, http } from "msw";
import { describe, expect, it } from "vitest";

import { routeTree } from "@/routeTree.gen";
import { server } from "@/test/msw";
import { createQueryClient } from "./query-client";

const meUrl = "http://localhost:8080/api/v1/me";

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

function mockStaffSession() {
  server.use(
    http.get(meUrl, () =>
      HttpResponse.json({
        user_id: "u1",
        discord_id: "1",
        roles: ["admin"],
        permissions: ["azeroth.admin.accounts.ban"],
      }),
    ),
  );
}

describe("app shell", () => {
  it("renders the portal layout at / for anonymous visitors", async () => {
    renderAt("/");

    expect(await screen.findByText("Community users")).toBeInTheDocument();
    expect(screen.getByRole("link", { name: "ac-community-gw" })).toBeInTheDocument();
  });

  it("renders the console layout at /admin for staff", async () => {
    mockStaffSession();
    renderAt("/admin");

    expect(await screen.findByText("Operations console")).toBeInTheDocument();
    expect(screen.getByText("console")).toBeInTheDocument();
  });

  it("lands staff on /admin when they open the portal home", async () => {
    mockStaffSession();
    renderAt("/");

    expect(await screen.findByText("Operations console")).toBeInTheDocument();
  });

  it("sends a signed-in player from /admin to the forbidden state", async () => {
    server.use(
      http.get(meUrl, () =>
        HttpResponse.json({ user_id: "u2", discord_id: "2", roles: [], permissions: [] }),
      ),
    );
    renderAt("/admin");

    expect(await screen.findByText("Access denied")).toBeInTheDocument();
  });
});
