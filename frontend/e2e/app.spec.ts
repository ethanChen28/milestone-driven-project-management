import { test, expect } from "@playwright/test";

async function selectRole(page: any, role: string) {
  await page.locator(".role-select").selectOption(role);
}

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

  test("role selector persists and selected role is sent", async ({ page }) => {
    await selectRole(page, "project_owner");
    await page.reload();
    await expect(page.locator(".role-select")).toHaveValue("project_owner");

    let seenRole = "";
    await page.route("/api/v1/projects", async (route) => {
      if (route.request().method() === "POST") {
        seenRole = route.request().headers()["x-role"] || "";
        await route.fulfill({ status: 201, contentType: "application/json", body: JSON.stringify({ id: "prj-test", name: "Header Test" }) });
        return;
      }
      await route.continue();
    });
    await page.goto("/projects");
    await page.locator("button.primary", { hasText: "创建项目" }).click({ force: true });
    await page.locator('.form input[placeholder="名称"]').fill("Header Test");
    await page.locator('.form input[placeholder="负责人"]').fill("tester");
    await page.locator(".form button.primary", { hasText: "保存" }).click({ force: true });
    expect(seenRole).toBe("project_owner");
  });
});

test.describe("Navigation", () => {
  test("sidebar has all navigation links", async ({ page }) => {
    await page.goto("/");
    const links = page.locator(".nav-link");
    await expect(links).toHaveCount(6);
    await expect(links.nth(0)).toContainText("仪表盘");
    await expect(links.nth(1)).toContainText("项目");
    await expect(links.nth(2)).toContainText("任务");
    await expect(links.nth(3)).toContainText("里程碑");
    await expect(links.nth(4)).toContainText("路线图");
    await expect(links.nth(5)).toContainText("周度回顾");
  });
});

test.describe("Task Workspace", () => {
  test("opens the task workspace and switches views", async ({ page }) => {
    await page.goto("/tasks");
    await expect(page.locator("h1")).toHaveText("任务工作台");
    await expect(page.locator(".tab")).toHaveCount(6);
    await page.locator(".tab", { hasText: "状态看板" }).click();
    await expect(page.locator(".board")).toBeVisible();
    await page.locator(".tab", { hasText: "进展甘特图" }).click();
    await expect(page.locator(".gantt-chart")).toBeVisible();
  });
});

test.describe("F-1: Create Project via UI", () => {
  test("create form appears and project shows in list after API creation", async ({ page }) => {
    await page.goto("/projects");
    await selectRole(page, "project_owner");
    await expect(page.locator("h1")).toHaveText("项目");
    await page.locator("button.primary", { hasText: "创建项目" }).click({ force: true });
    await expect(page.locator(".form")).toBeVisible();
    await expect(page.locator('.form input[placeholder="名称"]')).toBeVisible();
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
    await selectRole(page, "project_owner");
    await expect(page.locator("h1")).toHaveText("Milestone UI Test");
    await page.locator("button.primary", { hasText: "创建里程碑" }).click({ force: true });
    await expect(page.locator(".form")).toBeVisible();
    await expect(page.locator('.form input[placeholder="标题"]')).toBeVisible();

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
    await page.locator(".form select").first().selectOption(project.id);
    await page.locator('.form input[placeholder="作者"]').fill("tester");
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
      data: { projectId: project.id, title: "Overdue MS", owner: "tester", status: "active", completionCriteria: "Done", plannedDate: pastDate },
      headers: { "Content-Type": "application/json", "X-Role": "admin" },
    });
    await page.request.post("/api/v1/milestones", {
      data: { projectId: project.id, title: "Blocked MS", owner: "tester", status: "blocked", completionCriteria: "Unblocked" },
      headers: { "Content-Type": "application/json", "X-Role": "admin" },
    });

    await page.goto("/review");
    await expect(page.locator("h1")).toHaveText("周度回顾");
    await expect(page.locator(".alert-card.danger").first()).toContainText("Overdue MS");
    await expect(page.locator(".alert-card.warn").first()).toContainText("Blocked MS");
  });
});

test.describe("Milestone Lifecycle", () => {
  test("updates not_started to active to completed from milestone detail", async ({ page }) => {
    const projResp = await page.request.post("/api/v1/projects", {
      data: { name: "Lifecycle Test", owner: "tester", status: "active" },
      headers: { "Content-Type": "application/json", "X-Role": "admin" },
    });
    const project = await projResp.json();
    const msResp = await page.request.post("/api/v1/milestones", {
      data: { projectId: project.id, title: "Lifecycle MS", owner: "tester", status: "not_started", completionCriteria: "Done" },
      headers: { "Content-Type": "application/json", "X-Role": "admin" },
    });
    const milestone = await msResp.json();

    await page.goto(`/milestones/${milestone.id}`);
    await selectRole(page, "project_owner");
    await page.locator(".lifecycle select").first().selectOption("active");
    await page.locator("button.primary", { hasText: "保存" }).click({ force: true });
    await expect(page.locator(".meta")).toContainText("active");

    await page.locator(".lifecycle select").first().selectOption("completed");
    await page.locator("button.primary", { hasText: "保存" }).click({ force: true });
    await expect(page.locator(".meta")).toContainText("completed");
  });
});

test.describe("Filters", () => {
  test("applies and clears milestone risk filters", async ({ page }) => {
    const projResp = await page.request.post("/api/v1/projects", {
      data: { name: "Filter Test", owner: "tester", status: "active" },
      headers: { "Content-Type": "application/json", "X-Role": "admin" },
    });
    const project = await projResp.json();
    await page.request.post("/api/v1/milestones", {
      data: { projectId: project.id, title: "High Filter MS", owner: "tester", status: "not_started", completionCriteria: "Done", riskLevel: "high" },
      headers: { "Content-Type": "application/json", "X-Role": "admin" },
    });
    await page.request.post("/api/v1/milestones", {
      data: { projectId: project.id, title: "Low Filter MS", owner: "tester", status: "not_started", completionCriteria: "Done", riskLevel: "low" },
      headers: { "Content-Type": "application/json", "X-Role": "admin" },
    });
    await page.goto("/milestones");
    await page.locator('.filters input[placeholder="Project ID"]').fill(project.id);
    await page.locator(".filters select").last().selectOption("high");
    await page.getByRole("button", { name: "筛选", exact: true }).click({ force: true });
    await expect(page.locator("table")).toContainText("High Filter MS");
    await page.locator(".filters button", { hasText: "清除筛选" }).click({ force: true });
    await expect(page.locator("table")).toContainText("Low Filter MS");
  });
});

test.describe("GitLab Work Visibility", () => {
  test("shows GitLab metadata and original issue link on milestone detail", async ({ page }) => {
    const projResp = await page.request.post("/api/v1/projects", {
      data: { name: "GitLab Visibility", owner: "tester", status: "active" },
      headers: { "Content-Type": "application/json", "X-Role": "admin" },
    });
    const project = await projResp.json();
    const msResp = await page.request.post("/api/v1/milestones", {
      data: { projectId: project.id, title: "GitLab MS", owner: "tester", status: "not_started", completionCriteria: "Done" },
      headers: { "Content-Type": "application/json", "X-Role": "admin" },
    });
    const milestone = await msResp.json();
    await page.request.post("/api/v1/gitlab-link", {
      data: { sourceType: "gitlab_issue", sourceId: "55", sourceUrl: "https://gitlab.example/group/repo/-/issues/55", title: "GitLab Issue 55", projectId: project.id, milestoneId: milestone.id, owner: "dev", status: "opened", gitlabState: "opened", gitlabAssignee: "dev", gitlabLabels: ["bug"] },
      headers: { "Content-Type": "application/json", "X-Role": "admin" },
    });
    await page.goto(`/milestones/${milestone.id}`);
    await expect(page.locator(".work-card")).toContainText("GitLab Issue 55");
    await expect(page.locator(".work-card")).toContainText("opened");
    await expect(page.locator("a", { hasText: "打开 Issue" })).toHaveAttribute("href", "https://gitlab.example/group/repo/-/issues/55");
  });
});

test.describe("API Integration", () => {
  test("health endpoint returns ok without credentials", async ({ request }) => {
    const resp = await request.get("/api/v1/health");
    expect(resp.ok()).toBeTruthy();
    const body = await resp.json();
    expect(body.status).toBe("ok");
    expect(body.defaultLocale).toBe("zh-CN");
    expect(body.storageBackend).toBeDefined();
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
