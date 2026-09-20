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
const userId = "11111111-1111-1111-1111-111111111111";
const charsUrl = `http://localhost:8080/api/v1/azeroth/users/${userId}/characters`;
const mailUrl = "http://localhost:8080/api/v1/azeroth/mail";

let mailBody: unknown;

function renderPage() {
  const queryClient = createQueryClient();
  const router = createRouter({
    routeTree,
    context: { queryClient },
    history: createMemoryHistory({ initialEntries: [`/identity/users/${userId}/characters`] }),
  });

  render(
    <QueryClientProvider client={queryClient}>
      <RouterProvider router={router} />
    </QueryClientProvider>,
  );
}

describe("UserCharactersPage", () => {
  beforeEach(() => {
    mailBody = undefined;
  });
  it("lists characters and sends mail", async () => {
    server.use(
      http.get(meUrl, () =>
        HttpResponse.json({
          user_id: "u1",
          discord_id: "1",
          roles: [],
          permissions: ["azeroth.character.list", "azeroth.mail.send"],
        }),
      ),
      http.get(charsUrl, () =>
        HttpResponse.json({
          characters: [
            {
              guid: 7,
              name: "Thrall",
              level: 80,
              class_name: "Shaman",
              race_name: "Orc",
              online: true,
            },
          ],
        }),
      ),
      http.post(mailUrl, async ({ request }) => {
        mailBody = await request.json();
        return HttpResponse.json({ recipient: "Thrall", results: ["ok"] });
      }),
    );

    renderPage();

    expect(await screen.findByText("Thrall")).toBeInTheDocument();

    const user = userEvent.setup();
    await user.type(screen.getByLabelText("Character"), "Thrall");
    const money = screen.getByLabelText("Money (copper)");
    await user.clear(money);
    await user.type(money, "10000");

    await user.click(screen.getByRole("button", { name: "Send mail" }));
    await user.click(await screen.findByRole("button", { name: "Send" }));

    await waitFor(() => expect(mailBody).toBeDefined());
    expect(mailBody).toMatchObject({ character: "Thrall", money: 10000 });
  });

  it("hides the mail form without the permission", async () => {
    server.use(
      http.get(meUrl, () =>
        HttpResponse.json({
          user_id: "u1",
          discord_id: "1",
          roles: [],
          permissions: ["azeroth.character.list"],
        }),
      ),
      http.get(charsUrl, () => HttpResponse.json({ characters: [] })),
    );

    renderPage();

    expect(await screen.findByText(/azeroth.mail.send permission/)).toBeInTheDocument();
  });
});
