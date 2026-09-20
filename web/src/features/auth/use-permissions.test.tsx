import { QueryClientProvider } from "@tanstack/react-query";
import { render, screen } from "@testing-library/react";
import { HttpResponse, http } from "msw";
import { describe, expect, it } from "vitest";

import { createQueryClient } from "@/app/query-client";
import { server } from "@/test/msw";
import { Can } from "./use-permissions";

const meUrl = "http://localhost:8080/api/v1/me";

function renderCan(permission: string) {
  const queryClient = createQueryClient();
  return render(
    <QueryClientProvider client={queryClient}>
      <Can permission={permission} fallback={<span>denied</span>}>
        <span>allowed</span>
      </Can>
    </QueryClientProvider>,
  );
}

function meWithPermissions(permissions: string[]) {
  return http.get(meUrl, () =>
    HttpResponse.json({
      user_id: "u1",
      discord_id: "42",
      roles: [],
      permissions,
    }),
  );
}

describe("Can", () => {
  it("renders children when the permission is granted", async () => {
    server.use(meWithPermissions(["gw.store.catalog.read"]));
    renderCan("gw.store.catalog.read");
    expect(await screen.findByText("allowed")).toBeInTheDocument();
  });

  it("renders the fallback when the permission is missing", async () => {
    server.use(meWithPermissions([]));
    renderCan("gw.store.admin.products");
    expect(await screen.findByText("denied")).toBeInTheDocument();
  });
});
