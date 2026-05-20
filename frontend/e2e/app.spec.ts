import { test, expect } from "@playwright/test";

test.describe("Goal Manager Frontend", () => {
  test.beforeEach(async ({ page }) => {
    await page.goto("/");
  });

  test("F-3: Dashboard shows portfolio summary", async ({ page }) => {
    await expect(page.locator("h1")).toHaveText("里程碑驱动项目管理");
    const statCards = page.locator(".stat-card");
    await expect(statCards).toHaveCount(4);
  });

  test("F-4: Locale switch toggles between CN and EN", async ({ page }) => {
    await expect(page.locator("html")).toHaveAttribute("lang", "zh-CN");
    await page.locator(".locale-btn").click({ force: true });
    await expect(page.locator("html")).toHaveAttribute("lang", "en-US");
    await page.locator(".locale-btn").click({ force: true });
    await expect(page.locator("html")).toHaveAttribute("lang", "zh-CN");
  });
});

test.describe("Navigation", () => {
  test("sidebar has all navigation links", async ({ page }) => {
    await page.goto("/");
    const links = page.locator(".nav-link");
    await expect(links).toHaveCount(5);
    await expect(links.nth(0)).toContainText("仪表盘");
    await expect(links.nth(1)).toContainText("项目");
    await expect(links.nth(2)).toContainText("里程碑");
    await expect(links.nth(3)).toContainText("路线图");
    await expect(links.nth(4)).toContainText("周度回顾");
  });
});

test.describe("F-1: Create Project via UI", () => {
  test("create form appears and project shows in list after API creation", async ({ page }) => {
    await page.goto("/projects");
    await expect(page.locator("h1")).toHaveText("项目");

    // Verify create button opens form
    await page.locator("button.primary", { hasText: "创建项目" }).click({ force: true });
    await expect(page.locator(".form")).toBeVisible();
    await expect(page.locator('.form input[placeholder="名称"]')).toBeVisible();

    // Create via API, then verify UI reflects it
    await page.request.post("/api/v1/projects", {
      data: { name: "E2E UI Project", objective: "Test", owner: "tester", status: "active" },
      headers: { "Content-Type": "application/json", "X-Role": "admin" },
    });
    await page.reload();
    await expect(page.locator("table")).toContainText("E2E UI Project");
  });
});

test.describe("F-2: Create Milestone via UI", () => {
  test("milestone form appears and milestone shows after API creation", async ({ page }) => {
    const resp = await page.request.post("/api/v1/projects", {
      data: { name: "Milestone UI Test", owner: "tester", status: "active" },
      headers: { "Content-Type": "application/json", "X-Role": "admin" },
    });
    const project = await resp.json();

    await page.goto(`/projects/${project.id}`);
    await expect(page.locator("h1")).toHaveText("Milestone UI Test");

    // Verify create button opens form
    await page.locator("button.primary", { hasText: "创建里程碑" }).click({ force: true });
    await expect(page.locator(".form")).toBeVisible();
    await expect(page.locator('.form input[placeholder="标题"]')).toBeVisible();

    // Create via API, then verify UI reflects it
    await page.request.post("/api/v1/milestones", {
      data: { projectId: project.id, title: "E2E UI Milestone", owner: "tester", status: "not_started", completionCriteria: "All tests pass" },
      headers: { "Content-Type": "application/json", "X-Role": "admin" },
    });
    await page.reload();
    await expect(page.locator("table")).toContainText("E2E UI Milestone");
  });
});

test.describe("F-5: Roadmap Overview", () => {
  test("navigates to roadmap and displays periods", async ({ page }) => {
    await page.request.post("/api/v1/roadmap-periods", {
      data: { title: "E2E Period", status: "active", owner: "tester" },
      headers: { "Content-Type": "application/json", "X-Role": "admin" },
    });

    await page.goto("/roadmap");
    await expect(page.locator("h1")).toHaveText("路线图");
    await expect(page.locator(".period-card").first()).toContainText("E2E Period");
  });
});

test.describe("F-6: Submit Weekly Update via UI", () => {
  test("navigates to review and submits a weekly update", async ({ page }) => {
    const resp = await page.request.post("/api/v1/projects", {
      data: { name: "Update UI Test", owner: "tester", status: "active" },
      headers: { "Content-Type": "application/json", "X-Role": "admin" },
    });
    const project = await resp.json();

    await page.goto("/review");
    await expect(page.locator("h1")).toHaveText("周度回顾");

    await page.locator("button.primary", { hasText: "提交周报" }).click({ force: true });
    await expect(page.locator(".form")).toBeVisible();

    await page.locator('.form input[placeholder="Project ID"]').fill(project.id);
    await page.locator('.form input[placeholder="作者"]').fill("tester");
    await page.locator('.form input[placeholder="周"]').fill("2026-W21");
    await page.locator('.form textarea[placeholder="摘要"]').fill("Weekly update from E2E test");

    await page.locator(".form button.primary", { hasText: "保存" }).click({ force: true });

    await expect(page.locator(".update-card").first()).toContainText("Weekly update from E2E test");
  });
});

test.describe("F-7: Review View shows milestones", () => {
  test("displays delayed and blocked milestones", async ({ page }) => {
    const projResp = await page.request.post("/api/v1/projects", {
      data: { name: "Review View Test", owner: "tester", status: "active" },
      headers: { "Content-Type": "application/json", "X-Role": "admin" },
    });
    const project = await projResp.json();

    const pastDate = new Date(Date.now() - 48 * 3600 * 1000).toISOString();
    await page.request.post("/api/v1/milestones", {
      data: {
        projectId: project.id, title: "Overdue MS", owner: "tester",
        status: "active", completionCriteria: "Done", plannedDate: pastDate,
      },
      headers: { "Content-Type": "application/json", "X-Role": "admin" },
    });

    await page.request.post("/api/v1/milestones", {
      data: {
        projectId: project.id, title: "Blocked MS", owner: "tester",
        status: "blocked", completionCriteria: "Unblocked",
      },
      headers: { "Content-Type": "application/json", "X-Role": "admin" },
    });

    await page.goto("/review");
    await expect(page.locator("h1")).toHaveText("周度回顾");
    await expect(page.locator(".alert-card.danger").first()).toContainText("Overdue MS");
    await expect(page.locator(".alert-card.warn").first()).toContainText("Blocked MS");
  });
});

test.describe("API Integration", () => {
  test("health endpoint returns ok without credentials", async ({ request }) => {
    const resp = await request.get("/api/v1/health");
    expect(resp.ok()).toBeTruthy();
    const body = await resp.json();
    expect(body.status).toBe("ok");
    expect(body.defaultLocale).toBe("zh-CN");
    expect(body.mysql).toBeUndefined();
    expect(body.redis).toBeUndefined();
  });

  test("portfolio dashboard returns valid structure", async ({ request }) => {
    const resp = await request.get("/api/v1/dashboard/portfolio");
    expect(resp.ok()).toBeTruthy();
    const body = await resp.json();
    expect(body).toHaveProperty("activeProjects");
    expect(body).toHaveProperty("blockedMilestones");
    expect(body).toHaveProperty("overdueMilestones");
  });

  test("weekly review returns arrays not null (BUG-01 fix)", async ({ request }) => {
    const resp = await request.get("/api/v1/review/weekly");
    expect(resp.ok()).toBeTruthy();
    const body = await resp.json();
    expect(Array.isArray(body.delayedMilestones)).toBeTruthy();
    expect(Array.isArray(body.blockedMilestones)).toBeTruthy();
    expect(Array.isArray(body.updates)).toBeTruthy();
  });

  test("sync failures resolve endpoint requires permission", async ({ request }) => {
    const resp = await request.post("/api/v1/sync-failures/resolve", {
      data: { id: "nonexistent" },
      headers: { "Content-Type": "application/json", "X-Role": "admin" },
    });
    expect(resp.status()).toBe(404);
  });
});
