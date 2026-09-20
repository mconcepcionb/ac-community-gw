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
      permissions: ["azeroth.item.list"],
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

const thunderfury = {
  entry: 19019,
  name: "Thunderfury",
  quality: 5,
  quality_name: "Legendary",
  quality_color: "#ff8000",
  class: 2,
  class_name: "Weapon",
  subclass: 7,
  subclass_name: "One-Handed Sword",
  inventory_type_name: "One-Hand",
  item_level: 80,
  required_level: 60,
  stats: [],
  damage: [],
  spells: [],
};

describe("ItemsPage", () => {
  it("lists items", async () => {
    server.use(
      authorized(),
      http.get("http://localhost:8080/api/v1/azeroth/items", () =>
        HttpResponse.json({ items: [thunderfury] }),
      ),
    );

    renderAt("/items");

    expect(await screen.findByRole("link", { name: "19019" })).toBeInTheDocument();
    expect(screen.getByText("Thunderfury")).toBeInTheDocument();
  });
});

describe("ItemDetailPage", () => {
  it("renders the item detail", async () => {
    server.use(
      authorized(),
      http.get("http://localhost:8080/api/v1/azeroth/items/19019", () =>
        HttpResponse.json(thunderfury),
      ),
    );

    renderAt("/items/19019");

    expect(await screen.findByText("Thunderfury")).toBeInTheDocument();
    expect(screen.getByText("Legendary")).toBeInTheDocument();
  });
});
