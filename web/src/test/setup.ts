import "@testing-library/jest-dom/vitest";
import { configure } from "@testing-library/react";
import { afterAll, afterEach, beforeAll } from "vitest";

import { client } from "@/api/client";
import { server } from "./msw";

configure({ asyncUtilTimeout: 3000 });

Object.defineProperty(window, "scrollTo", { value: () => {}, writable: true });
Object.defineProperty(window.HTMLElement.prototype, "scrollIntoView", {
  value: () => {},
  writable: true,
});

// Radix components (Switch, Select) rely on ResizeObserver, absent in jsdom.
class ResizeObserverStub {
  observe() {}
  unobserve() {}
  disconnect() {}
}
globalThis.ResizeObserver = ResizeObserverStub as unknown as typeof ResizeObserver;

// React Query passes a jsdom AbortSignal, which undici rejects when building a
// Request ("not an instance of AbortSignal"). Tests never rely on cancellation,
// so the signal is dropped for every Request created in the test environment.
const NativeRequest = globalThis.Request;
class SanitizedRequest extends NativeRequest {
  constructor(input: RequestInfo | URL, init?: RequestInit) {
    if (init?.signal) {
      const sanitized: RequestInit = { ...init, signal: null };
      super(input, sanitized);
    } else {
      super(input, init);
    }
  }
}
globalThis.Request = SanitizedRequest as typeof Request;

beforeAll(() => {
  client.setConfig({ baseUrl: "http://localhost:8080" });
  server.listen({ onUnhandledRequest: "error" });
});

afterEach(() => {
  server.resetHandlers();
});

afterAll(() => {
  server.close();
});
