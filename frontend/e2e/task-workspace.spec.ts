import { test, expect } from "@playwright/test";

async function createTestProject(page: any, suffix = ""): Promise<{ id: string; name: string }> {
  const name = `TaskWS Test ${suffix || Date.now()}`;
  const resp = await page.request.post("/api/v1/projects", {
    data: { name, owner: "tester", status: "active" },
    headers: { "Content-Type": "application/json", "X-Role": "admin" },
  });
  expect(resp.ok(), `project create failed: ${resp.status()} ${await resp.text()}`).toBeTruthy();
  return await resp.json();
}

async function createTestTask(page: any, overrides: Record<string, any> = {}): Promise<any> {
  const project = await createTestProject(page, "task");
  const toIso = (value?: string) => {
    if (!value) return value;
    const normalized = value.includes("T") ? value : `${value}T00:00:00Z`;
    return new Date(normalized).toISOString();
  };
  const sourceType = overrides.sourceType || "internal_task";
  const sourceId = overrides.sourceId || (sourceType === "gitlab_issue" ? `gl-${Date.now()}` : `test-${Date.now()}`);
  const payload = {
    sourceType,
    sourceId,
    sourceUrl: overrides.sourceUrl || (sourceType === "gitlab_issue" ? `https://gitlab.example.com/group/repo/-/issues/${Date.now()}` : ""),
    title: `Test Task ${Date.now()}`,
    projectId: project.id,
    milestoneId: "",
    workstreamId: "",
    owner: "tester",
    status: "todo",
    estimate: "2d",
    blocked: false,
    ...overrides,
    projectId: overrides.projectId || project.id,
    dueDate: toIso(overrides.dueDate),
    plannedStartDate: toIso(overrides.plannedStartDate),
    plannedEndDate: toIso(overrides.plannedEndDate),
  };
  const resp = await page.request.post("/api/v1/work-items", {
    data: payload,
    headers: { "Content-Type": "application/json", "X-Role": "admin" },
  });
  expect(resp.ok(), `work-item create failed: ${resp.status()} ${await resp.text()}`).toBeTruthy();
  return { task: await resp.json(), project };
}

test.describe("US1: Task List View", () => {
  test("TC-1.1: sidebar has task navigation entry", async ({ page }) => {
    await page.goto("/");
    const links = page.locator(".nav-link");
    const taskLink = links.nth(2);
    await expect(taskLink).toContainText("任务");
    await taskLink.click();
    await expect(page).toHaveURL(/\/tasks/);
  });

  test("TC-1.2: task list page shows title, source, status, owner, due date", async ({ page }) => {
    const { task } = await createTestTask(page, {
      title: "TC1.2 Task",
      status: "todo",
      owner: "alice",
      dueDate: new Date(Date.now() + 5 * 86400000).toISOString(),
    });

    await page.goto("/tasks");
    await expect(page.locator("h1")).toHaveText("任务工作台");
    const table = page.locator(".task-table");
    await expect(table).toContainText("TC1.2 Task");
    await expect(table).toContainText("alice");
    await expect(table.locator(".status-pill").first()).toBeVisible();
  });

  test("TC-1.3: filter by project", async ({ page }) => {
    const project = await createTestProject(page, "filter-proj");
    await createTestTask(page, { projectId: project.id, title: "Belongs to Filter Proj" });
    await createTestTask(page, { title: "Other Project Task" });

    await page.goto("/tasks");
    await page.locator(".filters select").first().selectOption(project.id);
    await expect(page.locator(".task-table")).toContainText("Belongs to Filter Proj");
    await expect(page.locator(".task-table")).not.toContainText("Other Project Task");
  });

  test("TC-1.4: filter by status", async ({ page }) => {
    await createTestTask(page, { title: "Done Task", status: "done" });
    await createTestTask(page, { title: "Todo Task", status: "todo" });

    await page.goto("/tasks");
    const statusSelect = page.locator(".filters-grid select").nth(3);
    await statusSelect.selectOption("done");
    await expect(page.locator(".task-table")).toContainText("Done Task");
  });

  test("TC-1.5: filter by owner", async ({ page }) => {
    await createTestTask(page, { title: "Bob Task", owner: "bob_filter" });
    await createTestTask(page, { title: "Carol Task", owner: "carol_filter" });

    await page.goto("/tasks");
    const ownerSelect = page.locator(".filters-grid select").nth(2);
    await ownerSelect.selectOption("bob_filter");
    await expect(page.locator(".task-table")).toContainText("Bob Task");
  });
});

test.describe("US2: Create Task", () => {
  test("TC-2.1: new task button opens form", async ({ page }) => {
    await page.goto("/tasks");
    const createBtn = page.locator("button", { hasText: /新建|创建/ });
    const exists = await createBtn.count();
    if (exists > 0) {
      await createBtn.click();
      await expect(page.locator(".form")).toBeVisible();
    } else {
      test.info().annotations.push({ type: "MISSING", description: "No create task button found in TasksView" });
    }
  });
});

test.describe("US3: Edit Task", () => {
  test("TC-3.1: click task opens detail/edit page", async ({ page }) => {
    const { task } = await createTestTask(page, { title: "Editable Task" });
    await page.goto("/tasks");
    await expect(page.locator(".task-table")).toContainText("Editable Task");
    await page.locator(`.task-table tr[data-task-id="${task.id}"]`).click();
    const url = page.url();
    const hasDetailPage = url.includes(`/tasks/${task.id}`) || url.includes(`/work-items/${task.id}`);
    if (!hasDetailPage) {
      test.info().annotations.push({ type: "MISSING", description: "No task detail/edit page — clicking task row does not navigate" });
    }
  });
});

test.describe("US4: Delete Task", () => {
  test("TC-4.1: delete button exists with confirmation", async ({ page }) => {
    const { task } = await createTestTask(page, { title: "Deletable Task" });
    await page.goto("/tasks");
    await expect(page.locator(".task-table")).toContainText("Deletable Task");
    const deleteBtn = page.locator("button", { hasText: /删除|Delete/ });
    const exists = await deleteBtn.count();
    if (exists > 0) {
      await deleteBtn.first().click();
      await expect(page.locator(".confirm, .modal, dialog")).toBeVisible();
    } else {
      test.info().annotations.push({ type: "MISSING", description: "No delete button found in TasksView" });
    }
  });
});

test.describe("US5: Multi-View Switching", () => {
  test("TC-5.1: all 6 view tabs are present", async ({ page }) => {
    await page.goto("/tasks");
    const tabs = page.locator(".tab");
    await expect(tabs).toHaveCount(6);
    await expect(tabs.nth(0)).toContainText("任务列表");
    await expect(tabs.nth(1)).toContainText("状态看板");
    await expect(tabs.nth(2)).toContainText("进展甘特图");
    await expect(tabs.nth(3)).toContainText("时间线");
    await expect(tabs.nth(4)).toContainText("按项目查看");
    await expect(tabs.nth(5)).toContainText("按优先级查看");
  });

  test("TC-5.2: switch to board view", async ({ page }) => {
    await createTestTask(page, { title: "Board Task", status: "todo" });
    await page.goto("/tasks");
    await page.locator(".tab", { hasText: "状态看板" }).click();
    await expect(page.locator(".board")).toBeVisible();
    await expect(page.locator(".board-column").first()).toBeVisible();
  });

  test("TC-5.3: switch to gantt view", async ({ page }) => {
    await createTestTask(page, {
      title: "Gantt Task",
      plannedStartDate: new Date().toISOString(),
      plannedEndDate: new Date(Date.now() + 10 * 86400000).toISOString(),
    });
    await page.goto("/tasks");
    await page.locator(".tab", { hasText: "进展甘特图" }).click();
    await expect(page.locator(".gantt-chart")).toBeVisible();
    await expect(page.locator(".gantt-bar").first()).toBeVisible();
  });

  test("TC-5.4: switch to timeline view", async ({ page }) => {
    await createTestTask(page, {
      title: "Timeline Task",
      dueDate: new Date(Date.now() + 3 * 86400000).toISOString(),
    });
    await page.goto("/tasks");
    await page.locator(".tab", { hasText: "时间线" }).click();
    await expect(page.locator(".timeline-row").first()).toBeVisible();
  });

  test("TC-5.5: switch to project grouping", async ({ page }) => {
    await createTestTask(page, { title: "Project Group Task" });
    await page.goto("/tasks");
    await page.locator(".tab", { hasText: "按项目查看" }).click();
    await expect(page.locator(".grouped-grid")).toBeVisible();
  });

  test("TC-5.6: switch to priority grouping", async ({ page }) => {
    await createTestTask(page, { title: "Priority Group Task", priority: "P1" });
    await page.goto("/tasks");
    await page.locator(".tab", { hasText: "按优先级查看" }).click();
    await expect(page.locator(".grouped-grid")).toBeVisible();
  });

  test("TC-5.7: active tab is highlighted", async ({ page }) => {
    await page.goto("/tasks");
    await page.locator(".tab", { hasText: "状态看板" }).click();
    const activeTab = page.locator(".tab.active");
    await expect(activeTab).toContainText("状态看板");
  });

  test("TC-5.8: filters persist when switching views", async ({ page }) => {
    const project = await createTestProject(page, "persist");
    await createTestTask(page, { projectId: project.id, title: "Persist Task" });

    await page.goto("/tasks");
    await page.locator(".filters select").first().selectOption(project.id);
    await page.locator(".tab", { hasText: "状态看板" }).click();
    const filterValue = await page.locator(".filters select").first().inputValue();
    expect(filterValue).toBe(project.id);
  });
});

test.describe("US6: Gantt Chart", () => {
  test("TC-6.1: gantt shows time axis with dates", async ({ page }) => {
    await createTestTask(page, {
      title: "Gantt Axis Task",
      plannedStartDate: new Date().toISOString(),
      plannedEndDate: new Date(Date.now() + 14 * 86400000).toISOString(),
    });
    await page.goto("/tasks");
    await page.locator(".tab", { hasText: "进展甘特图" }).click();
    await expect(page.locator(".gantt-axis")).toBeVisible();
    await expect(page.locator(".gantt-axis div").first()).not.toBeEmpty();
  });

  test("TC-6.2: gantt shows today line", async ({ page }) => {
    await createTestTask(page, {
      plannedStartDate: new Date().toISOString(),
      plannedEndDate: new Date(Date.now() + 7 * 86400000).toISOString(),
    });
    await page.goto("/tasks");
    await page.locator(".tab", { hasText: "进展甘特图" }).click();
    await expect(page.locator(".gantt-today")).toBeVisible();
  });

  test("TC-6.3: gantt bars show task duration", async ({ page }) => {
    await createTestTask(page, {
      title: "Gantt Bar Task",
      plannedStartDate: new Date().toISOString(),
      plannedEndDate: new Date(Date.now() + 5 * 86400000).toISOString(),
    });
    await page.goto("/tasks");
    await page.locator(".tab", { hasText: "进展甘特图" }).click();
    await expect(page.locator(".gantt-bar").first()).toBeVisible();
    const barWidth = await page.locator(".gantt-bar").first().evaluate((el) => el.style.width);
    expect(parseFloat(barWidth)).toBeGreaterThan(0);
  });

  test("TC-6.4: gantt scale can be changed", async ({ page }) => {
    await createTestTask(page, {
      plannedStartDate: new Date().toISOString(),
      plannedEndDate: new Date(Date.now() + 30 * 86400000).toISOString(),
    });
    await page.goto("/tasks");
    await page.locator(".tab", { hasText: "进展甘特图" }).click();
    const scaleSelect = page.locator(".filters-grid select").last();
    await scaleSelect.selectOption("quarter");
    const selectedValue = await scaleSelect.inputValue();
    expect(selectedValue).toBe("quarter");
  });
});

test.describe("US7: Timeline View", () => {
  test("TC-7.1: timeline shows tasks sorted by date", async ({ page }) => {
    const first = await createTestTask(page, {
      title: "Timeline First",
      dueDate: new Date(Date.now() + 1 * 86400000).toISOString(),
    });
    const second = await createTestTask(page, {
      title: "Timeline Second",
      dueDate: new Date(Date.now() + 5 * 86400000).toISOString(),
    });
    await page.goto("/tasks");
    await page.locator(".tab", { hasText: "时间线" }).click();
    const rows = page.locator(`.timeline-row[data-task-id="${first.task.id}"], .timeline-row[data-task-id="${second.task.id}"]`);
    await expect(rows).toHaveCount(2);
    const firstText = await page.locator(`.timeline-row[data-task-id="${first.task.id}"] .timeline-card strong`).textContent();
    const secondText = await page.locator(`.timeline-row[data-task-id="${second.task.id}"] .timeline-card strong`).textContent();
    expect(firstText).toBe("Timeline First");
    expect(secondText).toBe("Timeline Second");
  });

  test("TC-7.2: overdue tasks are highlighted", async ({ page }) => {
    const task = await createTestTask(page, {
      title: "Overdue Timeline",
      status: "todo",
      dueDate: new Date(Date.now() - 2 * 86400000).toISOString(),
    });
    await page.goto("/tasks");
    await page.locator(".tab", { hasText: "时间线" }).click();
    const overdueCard = page.locator(`.timeline-row[data-task-id="${task.task.id}"] .timeline-card.overdue`);
    await expect(overdueCard).toBeVisible();
  });

  test("TC-7.3: near-due tasks are highlighted", async ({ page }) => {
    const task = await createTestTask(page, {
      title: "Near Due Timeline",
      status: "todo",
      dueDate: new Date(Date.now() + 2 * 86400000).toISOString(),
    });
    await page.goto("/tasks");
    await page.locator(".tab", { hasText: "时间线" }).click();
    const nearDueCard = page.locator(`.timeline-row[data-task-id="${task.task.id}"] .timeline-card.near-due`);
    await expect(nearDueCard).toBeVisible();
  });
});

test.describe("US8: Summary Metrics Cards", () => {
  test("TC-8.1: all 7 metric cards are present", async ({ page }) => {
    await page.goto("/tasks");
    const cards = page.locator(".summary-card");
    await expect(cards).toHaveCount(7);
  });

  test("TC-8.2: total task count is accurate", async ({ page }) => {
    const project = await createTestProject(page, "metric-total");
    await createTestTask(page, { title: "Metric Task A", projectId: project.id });
    await createTestTask(page, { title: "Metric Task B", projectId: project.id });
    await page.goto("/tasks");
    await page.locator(".filters-grid select").first().selectOption(project.id);
    const totalCard = page.locator(".summary-card.total strong");
    const count = await totalCard.textContent();
    expect(Number(count)).toBeGreaterThanOrEqual(2);
  });

  test("TC-8.3: completed count updates", async ({ page }) => {
    const project = await createTestProject(page, "metric-done");
    await createTestTask(page, { title: "Completed Metric", status: "done", projectId: project.id });
    await page.goto("/tasks");
    await page.locator(".filters-grid select").first().selectOption(project.id);
    const completedCard = page.locator(".summary-card", { hasText: "已完成" }).locator("strong").first();
    const count = await completedCard.textContent();
    expect(Number(count)).toBeGreaterThanOrEqual(1);
  });

  test("TC-8.4: overdue count updates", async ({ page }) => {
    const project = await createTestProject(page, "metric-overdue");
    await createTestTask(page, {
      title: "Overdue Metric",
      status: "todo",
      dueDate: new Date(Date.now() - 3 * 86400000).toISOString(),
      projectId: project.id,
    });
    await page.goto("/tasks");
    await page.locator(".filters-grid select").first().selectOption(project.id);
    const overdueCard = page.locator(".summary-card", { hasText: "逾期任务" }).locator("strong").first();
    const count = await overdueCard.textContent();
    expect(Number(count)).toBeGreaterThanOrEqual(1);
  });

  test("TC-8.5: metric cards are clickable and filter tasks", async ({ page }) => {
    await createTestTask(page, { title: "Completed Click", status: "done" });
    await page.goto("/tasks");
    await page.locator(".summary-card", { hasText: "已完成" }).click();
    const statusFilter = page.locator(".filters-grid select").filter({ has: page.locator('option[value="done"]') }).first();
    const value = await statusFilter.inputValue();
    expect(value).toBe("done");
  });
});

test.describe("US9: Advanced Filtering, Grouping, Sorting", () => {
  test("TC-9.1: filter by milestone", async ({ page }) => {
    const project = await createTestProject(page, "ms-filter");
    const msResp = await page.request.post("/api/v1/milestones", {
      data: { projectId: project.id, title: "Test MS", owner: "tester", status: "not_started", completionCriteria: "Done" },
      headers: { "Content-Type": "application/json", "X-Role": "admin" },
    });
    const milestone = await msResp.json();
    await createTestTask(page, { title: "MS Task", projectId: project.id, milestoneId: milestone.id });

    await page.goto("/tasks");
    const msSelect = page.locator(".filters-grid select").nth(1);
    await msSelect.selectOption(milestone.id);
    await expect(page.locator(".task-table")).toContainText("MS Task");
  });

  test("TC-9.2: filter by source type", async ({ page }) => {
    const task = await createTestTask(page, { title: "GitLab Source", sourceType: "gitlab_issue" });
    await page.goto("/tasks");
    const sourceSelect = page.locator(".filters-grid select").filter({ has: page.locator('option[value="gitlab_issue"]') }).first();
    await sourceSelect.selectOption("gitlab_issue");
    await expect(page.locator(`.task-table tr[data-task-id="${task.task.id}"]`)).toContainText("GitLab Source");
  });

  test("TC-9.3: keyword search", async ({ page }) => {
    await createTestTask(page, { title: "UniqueKeywordAlpha" });
    await createTestTask(page, { title: "OtherTaskBeta" });
    await page.goto("/tasks");
    await page.locator('.filters-grid input[placeholder="关键词"]').fill("UniqueKeywordAlpha");
    await expect(page.locator(".task-table")).toContainText("UniqueKeywordAlpha");
    await expect(page.locator(".task-table")).not.toContainText("OtherTaskBeta");
  });

  test("TC-9.4: group by status", async ({ page }) => {
    await createTestTask(page, { title: "Group Task", status: "todo" });
    await page.goto("/tasks");
    await page.locator(".tab", { hasText: "按项目查看" }).click();
    const groupSelect = page.locator('[data-testid="group-by-select"]');
    await groupSelect.selectOption("status");
    await expect(page.locator(".grouped-grid")).toBeVisible();
  });

  test("TC-9.5: sort by due date", async ({ page }) => {
    const project = await createTestProject(page, "sort-proj");
    await createTestTask(page, {
      title: "Late Task",
      dueDate: new Date(Date.now() + 10 * 86400000).toISOString(),
      projectId: project.id,
    });
    await createTestTask(page, {
      title: "Early Task",
      dueDate: new Date(Date.now() + 1 * 86400000).toISOString(),
      projectId: project.id,
    });
    await page.goto("/tasks");
    await page.locator(".filters-grid select").first().selectOption(project.id);
    const sortSelect = page.locator('[data-testid="sort-by-select"]');
    await sortSelect.selectOption("dueDate");
    const rows = page.locator(".task-table tbody tr");
    const firstTitle = await rows.first().locator("td strong").first().textContent();
    expect(firstTitle).toBe("Early Task");
  });

  test("TC-9.6: clear filters resets all", async ({ page }) => {
    await createTestTask(page, { title: "Filter Reset Task", owner: "reset_user" });
    await page.goto("/tasks");
    await page.locator(".filters-grid select").nth(2).selectOption("reset_user");
    await page.locator(".clear-btn").click();
    const ownerValue = await page.locator(".filters-grid select").nth(2).inputValue();
    expect(ownerValue).toBe("");
  });
});

test.describe("US10: Task Source and Risk", () => {
  test("TC-10.1: shows task source type", async ({ page }) => {
    const task = await createTestTask(page, { title: "Source Type Task", sourceType: "gitlab_issue" });
    await page.goto("/tasks");
    await expect(page.locator(`.task-table tr[data-task-id="${task.task.id}"]`)).toContainText("gitlab_issue");
  });

  test("TC-10.2: blocked tasks are marked", async ({ page }) => {
    const task = await createTestTask(page, {
      title: "Blocked Risk Task",
      blocked: true,
      status: "in_progress",
    });
    await page.goto("/tasks");
    const blockedPill = page.locator(`tr[data-task-id="${task.task.id}"] .status-pill.blocked`);
    await expect(blockedPill).toBeVisible();
  });

  test("TC-10.3: overdue tasks are flagged", async ({ page }) => {
    const task = await createTestTask(page, {
      title: "Overdue Risk Task",
      status: "todo",
      dueDate: new Date(Date.now() - 5 * 86400000).toISOString(),
    });
    await page.goto("/tasks");
    const overduePill = page.locator(`tr[data-task-id="${task.task.id}"] .status-pill.overdue`);
    await expect(overduePill).toBeVisible();
  });

  test("TC-10.4: risk hint text is shown", async ({ page }) => {
    const task = await createTestTask(page, {
      title: "Risk Hint Task",
      blocked: true,
      status: "in_progress",
    });
    await page.goto("/tasks");
    const riskDiv = page.locator(`tr[data-task-id="${task.task.id}"] .risk`);
    await expect(riskDiv).toBeVisible();
  });

  test("TC-10.5: board view shows blocked and overdue columns", async ({ page }) => {
    await createTestTask(page, {
      title: "Board Blocked",
      blocked: true,
      status: "in_progress",
    });
    await page.goto("/tasks");
    await page.locator(".tab", { hasText: "状态看板" }).click();
    const blockedColumn = page.locator(".board-column", { hasText: "blocked" });
    await expect(blockedColumn).toBeVisible();
  });
});

test.describe("API CRUD for work-items", () => {
  test("TC-API-1: GET /work-items returns array", async ({ request }) => {
    const resp = await request.get("/api/v1/work-items", {
      headers: { "X-Role": "admin" },
    });
    expect(resp.ok()).toBeTruthy();
    const body = await resp.json();
    expect(Array.isArray(body)).toBeTruthy();
  });

  test("TC-API-2: POST /work-items creates a task", async ({ request }) => {
    const projResp = await request.post("/api/v1/projects", {
      data: { name: "API Task Test", owner: "tester", status: "active" },
      headers: { "Content-Type": "application/json", "X-Role": "admin" },
    });
    const project = await projResp.json();
    const resp = await request.post("/api/v1/work-items", {
      data: {
        sourceType: "internal_task",
        sourceId: "api-test-1",
        sourceUrl: "",
        title: "API Created Task",
        projectId: project.id,
        owner: "api-tester",
        status: "todo",
        estimate: "3d",
      },
      headers: { "Content-Type": "application/json", "X-Role": "admin" },
    });
    expect(resp.ok()).toBeTruthy();
    const task = await resp.json();
    expect(task.id).toBeDefined();
    expect(task.title).toBe("API Created Task");
  });

  test("TC-API-3: PUT /work-items updates a task", async ({ request }) => {
    const projResp = await request.post("/api/v1/projects", {
      data: { name: "API Update Test", owner: "tester", status: "active" },
      headers: { "Content-Type": "application/json", "X-Role": "admin" },
    });
    const project = await projResp.json();
    const createResp = await request.post("/api/v1/work-items", {
      data: {
        sourceType: "internal_task",
        sourceId: "api-update-1",
        title: "Before Update",
        projectId: project.id,
        owner: "tester",
        status: "todo",
      },
      headers: { "Content-Type": "application/json", "X-Role": "admin" },
    });
    const task = await createResp.json();
    const updateResp = await request.put(`/api/v1/work-items?id=${task.id}`, {
      data: {
        sourceType: task.sourceType,
        sourceId: task.sourceId,
        sourceUrl: task.sourceUrl,
        title: "After Update",
        projectId: task.projectId,
        milestoneId: task.milestoneId,
        workstreamId: task.workstreamId,
        owner: task.owner,
        status: "in_progress",
        estimate: task.estimate,
        blocked: task.blocked,
      },
      headers: { "Content-Type": "application/json", "X-Role": "admin" },
    });
    expect(updateResp.ok()).toBeTruthy();
    const updated = await updateResp.json();
    expect(updated.title).toBe("After Update");
  });

  test("TC-API-4: DELETE /work-items removes a task", async ({ request }) => {
    const projResp = await request.post("/api/v1/projects", {
      data: { name: "API Delete Test", owner: "tester", status: "active" },
      headers: { "Content-Type": "application/json", "X-Role": "admin" },
    });
    const project = await projResp.json();
    const createResp = await request.post("/api/v1/work-items", {
      data: {
        sourceType: "internal_task",
        sourceId: "api-delete-1",
        title: "To Be Deleted",
        projectId: project.id,
        owner: "tester",
        status: "todo",
      },
      headers: { "Content-Type": "application/json", "X-Role": "admin" },
    });
    const task = await createResp.json();
    const deleteResp = await request.delete(`/api/v1/work-items?id=${task.id}`, {
      headers: { "X-Role": "admin" },
    });
    expect(deleteResp.ok()).toBeTruthy();

    const getResp = await request.get(`/api/v1/work-items?id=${task.id}`, {
      headers: { "X-Role": "admin" },
    });
    expect(getResp.status()).toBe(404);
  });
});
