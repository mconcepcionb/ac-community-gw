import { QueryClientProvider } from "@tanstack/react-query";
import { createMemoryHistory, createRouter, RouterProvider } from "@tanstack/react-router";
import { render, screen } from "@testing-library/react";
import { HttpResponse, http } from "msw";
import { describe, expect, it } from "vitest";

import { createQueryClient } from "@/app/query-client";
import { routeTree } from "@/routeTree.gen";
import { server } from "@/test/msw";

const meUrl = "http://localhost:8080/api/v1/me";
const myCharactersUrl = "http://localhost:8080/api/v1/azeroth/me/characters";

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

describe("MyCharactersPage", () => {
  it("lists my characters", async () => {
    server.use(
      http.get(meUrl, () =>
        HttpResponse.json({
          user_id: "u1",
          discord_id: "1",
          roles: [],
          permissions: ["azeroth.character.self"],
        }),
      ),
      http.get(myCharactersUrl, () =>
        HttpResponse.json({
          characters: [
            {
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
            },
          ],
        }),
      ),
    );

    renderAt("/characters");

    expect(await screen.findByText("Thrall")).toBeInTheDocument();
    expect(screen.getByText("Shaman")).toBeInTheDocument();
  });
});
