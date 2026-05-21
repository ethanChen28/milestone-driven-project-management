import { afterEach, describe, expect, it } from "vitest";

import { apiFetch, dateInputToIso, isoToDateInput } from "./api";

const originalFetch = globalThis.fetch;

afterEach(() => {
  globalThis.fetch = originalFetch;
});

describe("apiFetch", () => {
  it("sends role and user headers", async () => {
    let headers = new Headers();
    globalThis.fetch = (async (_input: RequestInfo | URL, init?: RequestInit) => {
      headers = new Headers(init?.headers);
      return new Response(JSON.stringify({ ok: true }), { status: 200 });
    }) as typeof fetch;

    await apiFetch("/health");

    expect(headers.get("X-Role")).toBe("contributor");
    expect(headers.get("X-User")).toBe("tester");
  });
});

describe("date helpers", () => {
  it("converts date inputs to backend-compatible RFC3339 values", () => {
    expect(dateInputToIso("2026-05-28")).toBe("2026-05-28T00:00:00.000Z");
    expect(isoToDateInput("2026-05-28T00:00:00.000Z")).toBe("2026-05-28");
  });
});
