import { QueryClientProvider } from "@tanstack/react-query";
import { createMemoryHistory, createRouter, RouterProvider } from "@tanstack/react-router";
import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { HttpResponse, http } from "msw";
import { beforeEach, describe, expect, it } from "vitest";

import { createQueryClient } from "@/app/query-client";
import { routeTree } from "@/routeTree.gen";
import { server } from "@/test/msw";

const meUrl = "http://localhost:8080/api/v1/me";
const productsUrl = "http://localhost:8080/api/v1/store/products";

let createBody: unknown;

function renderPage() {
  const queryClient = createQueryClient();
  const router = createRouter({
    routeTree,
    context: { queryClient },
    history: createMemoryHistory({ initialEntries: ["/admin/store"] }),
  });

  render(
    <QueryClientProvider client={queryClient}>
      <RouterProvider router={router} />
    </QueryClientProvider>,
  );
}

describe("StoreProductsPage", () => {
  beforeEach(() => {
    createBody = undefined;
  });
  it("lists products and creates one", async () => {
    server.use(
      http.get(meUrl, () =>
        HttpResponse.json({
          user_id: "u1",
          discord_id: "1",
          roles: [],
          permissions: ["store.catalog.read", "store.admin.products"],
        }),
      ),
      http.get(productsUrl, () =>
        HttpResponse.json({
          products: [
            {
              sku: "bag-16",
              name: "Traveler's Backpack",
              price_points: 500,
              money: 0,
              active: true,
              items: [],
            },
          ],
        }),
      ),
      http.post(productsUrl, async ({ request }) => {
        createBody = await request.json();
        return HttpResponse.json(
          {
            sku: "new-sku",
            name: "New product",
            price_points: 0,
            money: 100,
            active: true,
            items: [],
          },
          { status: 201 },
        );
      }),
    );

    renderPage();
    expect(await screen.findByText("Traveler's Backpack")).toBeInTheDocument();

    const user = userEvent.setup();
    await user.click(screen.getByRole("button", { name: "Create product" }));
    await user.type(await screen.findByLabelText("SKU"), "new-sku");
    await user.type(screen.getByLabelText("Name"), "New product");
    const money = screen.getByLabelText("Money (copper)");
    await user.clear(money);
    await user.type(money, "100");
    await user.click(screen.getByRole("button", { name: "Create" }));

    await waitFor(() => expect(createBody).toBeDefined());
    expect(createBody).toMatchObject({ sku: "new-sku", name: "New product", money: 100 });
  });
});
