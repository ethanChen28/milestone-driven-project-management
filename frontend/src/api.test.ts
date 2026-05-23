import { afterEach, describe, expect, it, vi } from "vitest";

import { apiFetch, clearAccessToken, dateInputToIso, isoToDateInput, listUsers, setAccessToken, AUTH_MODE_STORAGE_KEY } from "./api";

const originalFetch = globalThis.fetch;

afterEach(() => {
  globalThis.fetch = originalFetch;
  if (typeof window !== "undefined") window.localStorage.clear();
  clearAccessToken();
  vi.unstubAllGlobals();
});

function stubWindowStorage() {
  const values = new Map<string, string>();
  vi.stubGlobal("window", {
    localStorage: {
      getItem: (key: string) => values.get(key) ?? null,
      setItem: (key: string, value: string) => values.set(key, value),
      removeItem: (key: string) => values.delete(key),
      clear: () => values.clear(),
    },
  });
}

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

  it("sends bearer token in token mode", async () => {
    let headers = new Headers();
    stubWindowStorage();
    window.localStorage.setItem(AUTH_MODE_STORAGE_KEY, "token");
    setAccessToken("abc123");
    globalThis.fetch = (async (_input: RequestInfo | URL, init?: RequestInit) => {
      headers = new Headers(init?.headers);
      return new Response(JSON.stringify({ ok: true }), { status: 200 });
    }) as typeof fetch;

    await apiFetch("/health");

    expect(headers.get("Authorization")).toBe("Bearer abc123");
    expect(headers.has("X-Role")).toBe(false);
    expect(headers.has("X-User")).toBe(false);
  });

  it("loads user directory from the API", async () => {
    globalThis.fetch = (async () => new Response(JSON.stringify([{ id: "alice", username: "alice", displayName: "Alice", email: "a@example.com", status: "active", roles: ["project_owner"] }]), { status: 200 })) as typeof fetch;

    const users = await listUsers();

    expect(users[0].id).toBe("alice");
    expect(users[0].roles[0]).toBe("project_owner");
  });
});

describe("date helpers", () => {
  it("converts date inputs to backend-compatible RFC3339 values", () => {
    expect(dateInputToIso("2026-05-28")).toBe("2026-05-28T00:00:00.000Z");
    expect(isoToDateInput("2026-05-28T00:00:00.000Z")).toBe("2026-05-28");
  });
});
