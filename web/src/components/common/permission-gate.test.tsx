import { QueryClientProvider } from "@tanstack/react-query";
import { render, screen } from "@testing-library/react";
import { HttpResponse, http } from "msw";
import { describe, expect, it } from "vitest";

import { createQueryClient } from "@/app/query-client";
import { server } from "@/test/msw";
import { PermissionGate } from "./permission-gate";

const meUrl = "http://localhost:8080/api/v1/me";

function renderGate(permission: string) {
  return render(
    <QueryClientProvider client={createQueryClient()}>
      <PermissionGate permission={permission} fallback={<span>hidden</span>}>
        <span>secret</span>
      </PermissionGate>
    </QueryClientProvider>,
  );
}

describe("PermissionGate", () => {
  it("renders children when the permission is granted", async () => {
    server.use(
      http.get(meUrl, () =>
        HttpResponse.json({
          user_id: "u1",
          discord_id: "42",
          roles: [],
          permissions: ["store.admin.products"],
        }),
      ),
    );

    renderGate("store.admin.products");
    expect(await screen.findByText("secret")).toBeInTheDocument();
  });

  it("renders the fallback when the permission is missing", async () => {
    server.use(
      http.get(meUrl, () =>
        HttpResponse.json({
          user_id: "u1",
          discord_id: "42",
          roles: [],
          permissions: [],
        }),
      ),
    );

    renderGate("store.admin.products");
    expect(await screen.findByText("hidden")).toBeInTheDocument();
  });
});
