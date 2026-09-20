import { HttpResponse, http } from "msw";
import { setupServer } from "msw/node";

/** Default handlers: the session is anonymous unless a test overrides it. */
export const handlers = [
  http.get("http://localhost:8080/api/v1/me", () =>
    HttpResponse.json(
      { error: { code: "unauthorized", message: "authentication required" } },
      { status: 401 },
    ),
  ),
];

export const server = setupServer(...handlers);
