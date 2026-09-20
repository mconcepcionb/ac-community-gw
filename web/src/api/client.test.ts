import { HttpResponse, http } from "msw";
import { describe, expect, it } from "vitest";

import { server } from "@/test/msw";
import { ApiError, isApiError, isUnauthorized } from "./errors";
import { authMe, azerothAccountsCreate } from "./generated";

const meUrl = "http://localhost:8080/api/v1/me";

describe("api client", () => {
  it("normalises the backend error envelope into ApiError", async () => {
    server.use(
      http.get(meUrl, () =>
        HttpResponse.json(
          {
            error: { code: "unauthorized", message: "authentication required" },
            request_id: "req-1",
          },
          { status: 401 },
        ),
      ),
    );

    let caught: unknown;
    try {
      await authMe({ throwOnError: true });
    } catch (error) {
      caught = error;
    }

    expect(isApiError(caught)).toBe(true);
    expect(caught).toBeInstanceOf(ApiError);
    expect(isUnauthorized(caught)).toBe(true);
    if (caught instanceof ApiError) {
      expect(caught.status).toBe(401);
      expect(caught.code).toBe("unauthorized");
      expect(caught.requestId).toBe("req-1");
    }
  });

  it("keeps validation details from the envelope", async () => {
    server.use(
      http.post("http://localhost:8080/api/v1/azeroth/accounts", () =>
        HttpResponse.json(
          {
            error: {
              code: "unprocessable_entity",
              message: "validation failed",
              details: { field: "username" },
            },
          },
          { status: 422 },
        ),
      ),
    );

    let caught: unknown;
    try {
      await azerothAccountsCreate({
        body: { username: "", password: "", email: "" },
        throwOnError: true,
      });
    } catch (error) {
      caught = error;
    }

    expect(isApiError(caught)).toBe(true);
    if (caught instanceof ApiError) {
      expect(caught.status).toBe(422);
      expect(caught.details).toEqual({ field: "username" });
    }
  });

  it("returns typed data on success", async () => {
    server.use(
      http.get(meUrl, () =>
        HttpResponse.json({
          user_id: "u1",
          discord_id: "42",
          roles: ["member"],
          permissions: ["store.catalog.read"],
        }),
      ),
    );

    const { data } = await authMe();
    expect(data?.discord_id).toBe("42");
    expect(data?.permissions).toEqual(["store.catalog.read"]);
  });
});
