import { expect, test } from "@playwright/test";

function unique(prefix: string) {
  return `${prefix}-${Date.now()}`;
}

function iso(daysFromNow = 0) {
  return new Date(Date.now() + daysFromNow * 86400000).toISOString();
}

test.describe("Task CRUD", () => {
  test("creates, edits, and deletes a task from the UI", async ({ page }) => {
    const projectName = unique("Task Project");
    const milestoneName = unique("Task Milestone");
    const taskTitle = unique("Task Item");
    const editedTitle = `${taskTitle} Updated`;

    const projectResp = await page.request.post("/api/v1/projects", {
      data: {
        name: projectName,
        objective: "UI task CRUD coverage",
        owner: "tester",
        participants: ["alice", "tester"],
        status: "active",
        targetEndDate: iso(9),
      },
      headers: { "Content-Type": "application/json", "X-Role": "admin" },
    });
    const project = await projectResp.json();

    const milestoneResp = await page.request.post("/api/v1/milestones", {
      data: {
        projectId: project.id,
        title: milestoneName,
        owner: "tester",
        status: "not_started",
        completionCriteria: "done",
        plannedDate: iso(8),
      },
      headers: { "Content-Type": "application/json", "X-Role": "admin" },
    });
    const milestone = await milestoneResp.json();

    await page.goto("/tasks/new");
    await page.getByLabel("workspace role").selectOption("contributor");
    await page.getByLabel("current user").selectOption("alice");
    await expect(page.getByLabel("标题")).toBeVisible();
    await page.getByLabel("标题").fill(taskTitle);
    await page.getByLabel("项目").selectOption(project.id);
    await page.getByLabel("里程碑").selectOption(milestone.id);
    await expect(page.getByLabel("工作流")).toHaveCount(0);
    await page.getByTestId("task-owner-select").selectOption("alice");
    await page.getByLabel("预估工作量").fill("2d");
    await page.getByLabel("来源类型").selectOption("internal_task");
    await page.getByLabel("截止日期").fill(iso(8).slice(0, 10));
    await page.getByLabel("标签").fill("ops, ui");
    await page.getByLabel("阻塞").check();
    const postRequestPromise = page.waitForRequest((request) => request.url().includes("/api/v1/work-items") && request.method() === "POST");
    await page.getByRole("button", { name: "保存任务" }).click();
    const postRequest = await postRequestPromise;
    expect(postRequest.headers()["x-user"]).toBe("alice");
    await expect(page).toHaveURL(/\/tasks\?createdId=/);

    const editTaskResp = await page.request.post("/api/v1/work-items", {
      data: {
        sourceType: "internal_task",
        sourceId: unique("ui-edit"),
        sourceUrl: "",
        title: taskTitle,
        projectId: project.id,
        milestoneId: milestone.id,
        workstreamId: "",
        owner: "alice",
        status: "todo",
        estimate: "2d",
        blocked: false,
        dueDate: iso(8),
      },
      headers: { "Content-Type": "application/json", "X-Role": "admin" },
    });
    const editTask = await editTaskResp.json();

    await page.goto(`/tasks/${editTask.id}`);
    await expect(page).toHaveURL(/\/tasks\/.+/);
    await page.getByLabel("标题").fill(editedTitle);
    await page.getByLabel("状态").selectOption("in_progress");
    await page.getByRole("button", { name: "保存任务" }).click();

    await expect(page).toHaveURL(/\/tasks/);
    await page.goto(`/tasks/${editTask.id}`);
    await page.once("dialog", async (dialog) => dialog.accept());
    await page.getByRole("button", { name: "删除任务" }).click();

    await expect(page).toHaveURL(/\/tasks/);
  });

  test("shows milestone markers and project deadline cards", async ({ page }) => {
    const projectName = unique("Deadline Project");
    const milestoneName = unique("Deadline Milestone");
    const taskTitle = unique("Deadline Task");

    const projectResp = await page.request.post("/api/v1/projects", {
      data: {
        name: projectName,
        objective: "Deadline coverage",
        owner: "tester",
        status: "active",
        targetEndDate: iso(20),
      },
      headers: { "Content-Type": "application/json", "X-Role": "admin" },
    });
    const project = await projectResp.json();

    const milestoneResp = await page.request.post("/api/v1/milestones", {
      data: {
        projectId: project.id,
        title: milestoneName,
        owner: "tester",
        status: "active",
        completionCriteria: "done",
        plannedDate: iso(15),
      },
      headers: { "Content-Type": "application/json", "X-Role": "admin" },
    });
    const milestone = await milestoneResp.json();

    await page.request.post("/api/v1/work-items", {
      data: {
        sourceType: "internal_task",
        title: taskTitle,
        projectId: project.id,
        milestoneId: milestone.id,
        owner: "alice",
        status: "todo",
        estimate: "2d",
        dueDate: iso(16),
      },
      headers: { "Content-Type": "application/json", "X-Role": "contributor" },
    });

    await page.goto("/tasks");
    await page.locator(".filters-grid select").first().selectOption(project.id);
    await page.locator(".tab", { hasText: "时间线" }).click();
    await expect(page.locator(".milestone-chip").filter({ hasText: milestoneName }).first()).toBeVisible();
    await expect(page.locator(".summary-card.deadline")).toContainText(projectName);
    await expect(page.locator(".summary-card.deadline")).toContainText(iso(20).slice(0, 10));
  });
});
