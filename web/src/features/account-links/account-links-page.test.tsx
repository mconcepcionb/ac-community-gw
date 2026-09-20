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
const linksUrl = "http://localhost:8080/api/v1/azeroth/account-links";
const userId = "11111111-1111-1111-1111-111111111111";

let createBody: unknown;

function renderPage() {
  const queryClient = createQueryClient();
  const router = createRouter({
    routeTree,
    context: { queryClient },
    history: createMemoryHistory({ initialEntries: ["/account-links"] }),
  });

  render(
    <QueryClientProvider client={queryClient}>
      <RouterProvider router={router} />
    </QueryClientProvider>,
  );
}

describe("AccountLinksPage", () => {
  beforeEach(() => {
    createBody = undefined;
  });
  it("lists links and creates one", async () => {
    server.use(
      http.get(meUrl, () =>
        HttpResponse.json({
          user_id: "u1",
          discord_id: "1",
          roles: [],
          permissions: ["azeroth.account.read", "azeroth.account.link"],
        }),
      ),
      http.get(linksUrl, () =>
        HttpResponse.json({
          links: [
            {
              user_id: userId,
              account_username: "ADMIN",
              account_id: 5,
              linked_at: "2024-01-02T03:04:05Z",
            },
          ],
        }),
      ),
      http.post(linksUrl, async ({ request }) => {
        createBody = await request.json();
        return HttpResponse.json(
          {
            user_id: userId,
            account_username: "NEW",
            account_id: 6,
            linked_at: "2024-01-02T03:04:05Z",
          },
          { status: 201 },
        );
      }),
    );

    renderPage();
    expect(await screen.findByText("ADMIN")).toBeInTheDocument();

    const user = userEvent.setup();
    await user.click(screen.getByRole("button", { name: "Create link" }));
    await user.type(await screen.findByLabelText("Account username"), "NEW");
    await user.type(screen.getByLabelText("Discord id"), "42");
    await user.click(screen.getByRole("button", { name: "Create" }));

    await waitFor(() => expect(createBody).toBeDefined());
    expect(createBody).toMatchObject({ account_username: "NEW", discord_id: "42" });
  });
});
