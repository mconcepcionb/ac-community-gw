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
const walletUrl = "http://localhost:8080/api/v1/store/wallet";
const ordersUrl = "http://localhost:8080/api/v1/store/orders";

let purchaseBody: unknown;

function renderPage() {
  const queryClient = createQueryClient();
  const router = createRouter({
    routeTree,
    context: { queryClient },
    history: createMemoryHistory({ initialEntries: ["/wallet"] }),
  });

  render(
    <QueryClientProvider client={queryClient}>
      <RouterProvider router={router} />
    </QueryClientProvider>,
  );
}

describe("WalletPage", () => {
  beforeEach(() => {
    purchaseBody = undefined;
  });
  it("shows the balance and purchases a product", async () => {
    server.use(
      http.get(meUrl, () =>
        HttpResponse.json({
          user_id: "u1",
          discord_id: "1",
          roles: [],
          permissions: ["gw.store.wallet.read", "gw.store.orders.read", "gw.store.purchase"],
        }),
      ),
      http.get(walletUrl, () => HttpResponse.json({ user_id: "u1", balance: 1000 })),
      http.get(ordersUrl, () =>
        HttpResponse.json({
          orders: [
            {
              order_id: "o1",
              user_id: "u1",
              sku: "bag-16",
              price_points: 500,
              character: "Thrall",
              status: "delivered",
              created_at: "2024-01-02T03:04:05Z",
            },
          ],
        }),
      ),
      http.post(ordersUrl, async ({ request }) => {
        purchaseBody = await request.json();
        return HttpResponse.json(
          {
            order_id: "o2",
            user_id: "u1",
            sku: "bag-16",
            price_points: 500,
            character: "Thrall",
            status: "delivered",
            created_at: "2024-01-02T03:04:05Z",
          },
          { status: 201 },
        );
      }),
    );

    renderPage();
    expect(await screen.findByText("1000")).toBeInTheDocument();
    expect(screen.getByText("delivered")).toBeInTheDocument();

    const user = userEvent.setup();
    await user.click(screen.getByRole("button", { name: "Purchase" }));
    await user.type(await screen.findByLabelText("SKU"), "bag-16");
    await user.type(screen.getByLabelText("Character"), "Thrall");
    await user.click(screen.getByRole("button", { name: "Review purchase" }));
    await user.click(await screen.findByRole("button", { name: "Buy" }));

    await waitFor(() => expect(purchaseBody).toBeDefined());
    expect(purchaseBody).toMatchObject({ sku: "bag-16", character: "Thrall" });
  });
});
