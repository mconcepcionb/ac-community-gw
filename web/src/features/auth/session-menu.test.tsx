import { QueryClientProvider } from "@tanstack/react-query";
import { render, screen } from "@testing-library/react";
import { HttpResponse, http } from "msw";
import { describe, expect, it } from "vitest";

import { createQueryClient } from "@/app/query-client";
import { server } from "@/test/msw";
import { SessionMenu } from "./session-menu";

const meUrl = "http://localhost:8080/api/v1/me";

function renderMenu() {
  const queryClient = createQueryClient();
  return render(
    <QueryClientProvider client={queryClient}>
      <SessionMenu />
    </QueryClientProvider>,
  );
}

describe("SessionMenu", () => {
  it("offers login when the session is anonymous", async () => {
    server.use(
      http.get(meUrl, () =>
        HttpResponse.json(
          { error: { code: "unauthorized", message: "authentication required" } },
          { status: 401 },
        ),
      ),
    );

    renderMenu();

    expect(await screen.findByRole("button", { name: /login with discord/i })).toBeInTheDocument();
  });

  it("shows the principal and logout when authenticated", async () => {
    server.use(
      http.get(meUrl, () =>
        HttpResponse.json({
          user_id: "u1",
          discord_id: "42",
          roles: ["member"],
          permissions: ["store.catalog.read"],
        }),
      ),
    );

    renderMenu();

    expect(await screen.findByText("42")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: /logout/i })).toBeInTheDocument();
  });
});
