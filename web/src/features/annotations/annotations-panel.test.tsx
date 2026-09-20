import { QueryClientProvider } from "@tanstack/react-query";
import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { HttpResponse, http } from "msw";
import { beforeEach, describe, expect, it } from "vitest";

import { createQueryClient } from "@/app/query-client";
import { server } from "@/test/msw";
import { AnnotationsPanel } from "./annotations-panel";

const meUrl = "http://localhost:8080/api/v1/me";
const notesUrl = "http://localhost:8080/api/v1/admin/annotations";

let createBody: unknown;

function renderPanel() {
  const queryClient = createQueryClient();
  render(
    <QueryClientProvider client={queryClient}>
      <AnnotationsPanel targetType="account" targetId="ADMIN" />
    </QueryClientProvider>,
  );
}

describe("AnnotationsPanel", () => {
  beforeEach(() => {
    createBody = undefined;
  });

  it("lists annotations and adds one", async () => {
    server.use(
      http.get(meUrl, () =>
        HttpResponse.json({
          user_id: "u1",
          discord_id: "1",
          roles: [],
          permissions: ["gw.notes.read", "gw.notes.write"],
        }),
      ),
      http.get(notesUrl, () =>
        HttpResponse.json({
          annotations: [
            {
              id: "a1",
              target_type: "account",
              target_id: "ADMIN",
              author_id: "u1",
              body: "Chargeback risk",
              created_at: "2024-01-02T03:04:05Z",
              updated_at: "2024-01-02T03:04:05Z",
            },
          ],
        }),
      ),
      http.post(notesUrl, async ({ request }) => {
        createBody = await request.json();
        return HttpResponse.json(
          {
            id: "a2",
            target_type: "account",
            target_id: "ADMIN",
            author_id: "u1",
            body: "second",
            created_at: "2024-01-02T03:04:05Z",
            updated_at: "2024-01-02T03:04:05Z",
          },
          { status: 201 },
        );
      }),
    );

    renderPanel();
    expect(await screen.findByText("Chargeback risk")).toBeInTheDocument();

    const user = userEvent.setup();
    await user.type(screen.getByLabelText("New annotation"), "second");
    await user.click(screen.getByRole("button", { name: "Add note" }));

    await waitFor(() => expect(createBody).toBeDefined());
    expect(createBody).toMatchObject({
      target_type: "account",
      target_id: "ADMIN",
      body: "second",
    });
  });

  it("hides the write controls without gw.notes.write", async () => {
    server.use(
      http.get(meUrl, () =>
        HttpResponse.json({
          user_id: "u1",
          discord_id: "1",
          roles: [],
          permissions: ["gw.notes.read"],
        }),
      ),
      http.get(notesUrl, () => HttpResponse.json({ annotations: [] })),
    );

    renderPanel();
    expect(await screen.findByText("No annotations")).toBeInTheDocument();
    expect(screen.queryByRole("button", { name: "Add note" })).not.toBeInTheDocument();
  });
});
