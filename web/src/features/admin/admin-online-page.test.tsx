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
const onlineUrl = "http://localhost:8080/api/v1/azeroth/online";
const charactersUrl = "http://localhost:8080/api/v1/azeroth/characters";
const kickUrl = "http://localhost:8080/api/v1/azeroth/players/Thrall/kick";
const announceUrl = "http://localhost:8080/api/v1/azeroth/announce";

let kickBody: unknown;
let announceBody: unknown;

function renderPage() {
  const queryClient = createQueryClient();
  const router = createRouter({
    routeTree,
    context: { queryClient },
    history: createMemoryHistory({ initialEntries: ["/admin/online"] }),
  });

  render(
    <QueryClientProvider client={queryClient}>
      <RouterProvider router={router} />
    </QueryClientProvider>,
  );
}

describe("AdminOnlinePage", () => {
  beforeEach(() => {
    kickBody = undefined;
    announceBody = undefined;
  });

  it("lists online players, kicks one and sends an announcement", async () => {
    server.use(
      http.get(meUrl, () =>
        HttpResponse.json({
          user_id: "u1",
          discord_id: "1",
          roles: [],
          permissions: [
            "azeroth.admin.players.read",
            "azeroth.admin.players.kick",
            "azeroth.admin.announce",
          ],
        }),
      ),
      http.get(onlineUrl, () =>
        HttpResponse.json({
          players: [
            {
              guid: 1,
              name: "Thrall",
              level: 80,
              class_name: "Shaman",
              race_name: "Orc",
              guild: "Horde",
              account_id: 5,
            },
          ],
        }),
      ),
      http.get(charactersUrl, () =>
        HttpResponse.json({ characters: [{ guid: 1, name: "Thrall", class_name: "Shaman" }] }),
      ),
      http.post(kickUrl, async ({ request }) => {
        kickBody = await request.json();
        return HttpResponse.json({ result: "kicked" });
      }),
      http.post(announceUrl, async ({ request }) => {
        announceBody = await request.json();
        return HttpResponse.json({ result: "announced" });
      }),
    );

    renderPage();
    expect(await screen.findByText("Thrall")).toBeInTheDocument();

    const user = userEvent.setup();
    await user.type(screen.getByLabelText("Player or character name"), "Thrall");
    await user.click(screen.getAllByRole("button", { name: "Kick" })[0]);
    await user.type(await screen.findByLabelText("Reason (optional)"), "afk");
    await user.click(screen.getByRole("button", { name: "Confirm kick" }));

    await waitFor(() => expect(kickBody).toBeDefined());
    expect(kickBody).toMatchObject({ reason: "afk" });

    await user.type(screen.getByLabelText("Message"), "Server restart in 5 minutes");
    await user.click(screen.getByRole("button", { name: "Announce" }));
    await user.click(await screen.findByRole("button", { name: "Send" }));

    await waitFor(() => expect(announceBody).toBeDefined());
    expect(announceBody).toMatchObject({ message: "Server restart in 5 minutes" });
  });
});
