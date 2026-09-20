import { QueryClientProvider } from "@tanstack/react-query";
import { render, screen } from "@testing-library/react";
import { HttpResponse, http } from "msw";
import { describe, expect, it } from "vitest";

import { createQueryClient } from "@/app/query-client";
import { server } from "@/test/msw";
import { ProfilePage } from "./profile-page";

const meUrl = "http://localhost:8080/api/v1/me";

describe("ProfilePage", () => {
  it("shows the full profile of the signed-in user", async () => {
    server.use(
      http.get(meUrl, () =>
        HttpResponse.json({
          user_id: "u1",
          discord_id: "42",
          username: "alice",
          global_name: "Alice",
          display_name: "Alice",
          avatar: "abc",
          created_at: "2024-01-02T03:04:05Z",
          roles: ["member"],
          permissions: ["store.catalog.read"],
        }),
      ),
    );

    render(
      <QueryClientProvider client={createQueryClient()}>
        <ProfilePage />
      </QueryClientProvider>,
    );

    expect((await screen.findAllByText("Alice")).length).toBeGreaterThan(0);
    expect(screen.getAllByText("42").length).toBeGreaterThan(0);
    expect(screen.getByText("member")).toBeInTheDocument();
    expect(screen.getByText("store.catalog.read")).toBeInTheDocument();
  });
});
