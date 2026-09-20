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

function authorized(permissions: string[] = ["azeroth.character.list"]) {
  return http.get(meUrl, () =>
    HttpResponse.json({ user_id: "u1", discord_id: "1", roles: [], permissions }),
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

let banBody: unknown;
let mailBody: unknown;

describe("CharactersPage", () => {
  beforeEach(() => {
    banBody = undefined;
    mailBody = undefined;
  });

  it("lists characters for an account", async () => {
    server.use(
      authorized(),
      http.get("http://localhost:8080/api/v1/azeroth/accounts", () =>
        HttpResponse.json({ accounts: [{ id: 1, username: "ADMIN" }] }),
      ),
      http.get("http://localhost:8080/api/v1/azeroth/characters", () =>
        HttpResponse.json({ characters: [thrall], total: 1 }),
      ),
    );

    renderAt("/admin/characters?account=ADMIN");

    expect(await screen.findByRole("link", { name: "Thrall" })).toBeInTheDocument();
    expect(screen.getByText("Shaman")).toBeInTheDocument();
  });

  it("lists all characters when no account is given", async () => {
    server.use(
      authorized(),
      http.get("http://localhost:8080/api/v1/azeroth/characters", () =>
        HttpResponse.json({ characters: [thrall], total: 1 }),
      ),
    );
    renderAt("/admin/characters");

    expect(await screen.findByRole("link", { name: "Thrall" })).toBeInTheDocument();
  });
});

describe("CharacterDetailPage", () => {
  beforeEach(() => {
    banBody = undefined;
    mailBody = undefined;
  });

  function detailHandlers() {
    return [
      http.get("http://localhost:8080/api/v1/admin/annotations", () =>
        HttpResponse.json({ annotations: [] }),
      ),
      http.get("http://localhost:8080/api/v1/admin/audit", () =>
        HttpResponse.json({ entries: [] }),
      ),
    ];
  }

  it("renders the character detail", async () => {
    server.use(
      authorized(),
      http.get("http://localhost:8080/api/v1/azeroth/characters/Thrall", () =>
        HttpResponse.json(thrall),
      ),
      ...detailHandlers(),
    );

    renderAt("/admin/characters/Thrall");

    expect(await screen.findByText("80")).toBeInTheDocument();
    expect(screen.getByText("Shaman")).toBeInTheDocument();
  });

  it("bans a character", async () => {
    server.use(
      authorized(["azeroth.character.list", "azeroth.admin.characters.ban"]),
      http.get("http://localhost:8080/api/v1/azeroth/characters/Thrall", () =>
        HttpResponse.json(thrall),
      ),
      ...detailHandlers(),
      http.post(
        "http://localhost:8080/api/v1/azeroth/characters/Thrall/ban",
        async ({ request }) => {
          banBody = await request.json();
          return HttpResponse.json({ result: "banned" });
        },
      ),
    );

    renderAt("/admin/characters/Thrall");
    expect(await screen.findByText("80")).toBeInTheDocument();

    const user = userEvent.setup();
    await user.click(screen.getByRole("button", { name: "Ban" }));
    await user.type(await screen.findByLabelText("Reason"), "cheating");
    await user.click(screen.getByRole("button", { name: "Confirm ban" }));

    await waitFor(() => expect(banBody).toBeDefined());
    expect(banBody).toMatchObject({ duration: "1d", reason: "cheating" });
  });

  it("sends mail to a character as staff", async () => {
    server.use(
      authorized(["azeroth.character.list", "azeroth.admin.mail.send"]),
      http.get("http://localhost:8080/api/v1/azeroth/characters/Thrall", () =>
        HttpResponse.json(thrall),
      ),
      ...detailHandlers(),
      http.post(
        "http://localhost:8080/api/v1/azeroth/admin/characters/Thrall/mail",
        async ({ request }) => {
          mailBody = await request.json();
          return HttpResponse.json({ recipient: "Thrall", results: ["sent"] });
        },
      ),
    );

    renderAt("/admin/characters/Thrall");
    expect(await screen.findByText("80")).toBeInTheDocument();

    const user = userEvent.setup();
    await user.click(screen.getByRole("button", { name: "Send mail" }));
    await user.type(await screen.findByLabelText("Money (copper)"), "100");
    await user.click(screen.getByRole("button", { name: "Send" }));

    await waitFor(() => expect(mailBody).toBeDefined());
    expect(mailBody).toMatchObject({ money: 100 });
  });
});
