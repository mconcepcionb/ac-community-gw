import { QueryClientProvider } from "@tanstack/react-query";
import { createMemoryHistory, createRouter, RouterProvider } from "@tanstack/react-router";
import { render, screen } from "@testing-library/react";
import { HttpResponse, http } from "msw";
import { describe, expect, it } from "vitest";

import { createQueryClient } from "@/app/query-client";
import { routeTree } from "@/routeTree.gen";
import { server } from "@/test/msw";

const meUrl = "http://localhost:8080/api/v1/me";

function authorized() {
  return http.get(meUrl, () =>
    HttpResponse.json({
      user_id: "u1",
      discord_id: "1",
      roles: [],
      permissions: ["azeroth.character.list"],
    }),
  );
}

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

const thrall = {
  guid: 7,
  name: "Thrall",
  race: 2,
  race_name: "Orc",
  class: 7,
  class_name: "Shaman",
  gender: 0,
  level: 80,
  online: true,
  guild: "Horde",
  money: 999,
  total_time: 1234,
  logout_time: null,
};

describe("CharactersPage", () => {
  it("lists characters for an account", async () => {
    server.use(
      authorized(),
      http.get("http://localhost:8080/api/v1/azeroth/characters", () =>
        HttpResponse.json({ characters: [thrall] }),
      ),
    );

    renderAt("/characters?account=ADMIN");

    expect(await screen.findByRole("link", { name: "Thrall" })).toBeInTheDocument();
    expect(screen.getByText("Shaman")).toBeInTheDocument();
  });

  it("prompts for an account when none is given", async () => {
    server.use(authorized());
    renderAt("/characters");

    expect(await screen.findByText("Choose an account")).toBeInTheDocument();
  });
});

describe("CharacterDetailPage", () => {
  it("renders the character detail", async () => {
    server.use(
      authorized(),
      http.get("http://localhost:8080/api/v1/azeroth/characters/Thrall", () =>
        HttpResponse.json(thrall),
      ),
    );

    renderAt("/characters/Thrall");

    expect(await screen.findByText("80")).toBeInTheDocument();
    expect(screen.getByText("Shaman")).toBeInTheDocument();
  });
});
