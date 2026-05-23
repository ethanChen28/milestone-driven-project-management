import { expect, test } from "@playwright/test";

test("user service auth contract exposes directory, login token, and token-mode mutation path", async ({ request }) => {
  const users = await request.get("/api/v1/users?workspaceId=default");
  expect(users.ok()).toBeTruthy();
  const directory = await users.json();
  expect(directory.some((user: { id: string }) => user.id === "alice")).toBeTruthy();

  const login = await request.post("/api/v1/auth/login", {
    data: { username: "alice", password: "password", workspaceId: "default" },
  });
  expect(login.ok()).toBeTruthy();
  const session = await login.json();
  expect(session.accessToken).toBeTruthy();
  expect(session.user.id).toBe("alice");

  const health = await request.get("/api/v1/health");
  const healthBody = await health.json();
  test.skip(healthBody.authMode !== "token", "backend is not running in token auth mode");

  const created = await request.post("/api/v1/projects", {
    headers: { Authorization: `Bearer ${session.accessToken}` },
    data: { name: `Token Project ${Date.now()}`, owner: "alice", participants: ["alice"], status: "active", healthStatus: "on_track" },
  });
  expect(created.status()).toBe(201);
});
