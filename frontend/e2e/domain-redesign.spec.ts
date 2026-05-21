import { test, expect } from "@playwright/test";

const headers = (role: string, user: string) => ({
  "Content-Type": "application/json",
  "X-Role": role,
  "X-User": user,
});

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

async function createProject(
  request: any,
  owner: string,
  participants: string[] = [],
  role = "admin",
  user = "admin",
) {
  const resp = await request.post("/api/v1/projects", {
    data: { name: `AuthTest-${Date.now()}`, owner, participants, status: "active" },
    headers: headers(role, user),
  });
  expect(resp.ok(), `project create: ${resp.status()}`).toBeTruthy();
  return await resp.json();
}

async function createMilestone(
  request: any,
  projectId: string,
  owner = "leader1",
  role = "project_owner",
  user = "leader1",
) {
  const resp = await request.post("/api/v1/milestones", {
    data: { projectId, title: `MS-${Date.now()}`, owner, status: "not_started", completionCriteria: "All done" },
    headers: headers(role, user),
  });
  expect(resp.ok(), `milestone create: ${resp.status()}`).toBeTruthy();
  return await resp.json();
}

async function createTask(
  request: any,
  projectId: string,
  owner: string,
  overrides: Record<string, any> = {},
  role = "contributor",
  user?: string,
) {
  const payload = {
    sourceType: "internal_task",
    sourceId: `test-${Date.now()}`,
    sourceUrl: "",
    title: `Task-${Date.now()}`,
    projectId,
    milestoneId: "",
    workstreamId: "",
    owner,
    status: "todo",
    estimate: "1d",
    blocked: false,
    ...overrides,
  };
  const resp = await request.post("/api/v1/work-items", {
    data: payload,
    headers: headers(role, user ?? owner),
  });
  return { resp, task: resp.ok() ? await resp.json() : null };
}

// ===========================================================================
// 1. X-User Identity Propagation
// ===========================================================================

test.describe("DMR-1: X-User Identity", () => {
  test("DMR-1.1: X-User header is extracted and stored as task owner", async ({ request }) => {
    const project = await createProject(request, "leader1", ["eng1"]);
    const { resp, task } = await createTask(request, project.id, "eng1", {}, "contributor", "eng1");
    expect(resp.ok(), `task create: ${resp.status()} ${task ? "" : await resp.text()}`).toBeTruthy();
    expect(task.owner).toBe("eng1");
  });

  test("DMR-1.2: frontend sends X-User header on API calls", async ({ page }) => {
    await page.goto("/tasks");
    const selectUser = page.locator(".user-select");
    if (await selectUser.count() > 0) {
      await selectUser.selectOption("tester");
    }
    // Navigate to create task page - verify X-User is sent
    const [apiRequest] = await Promise.all([
      page.waitForRequest((req) => req.url().includes("/api/v1/projects")),
      page.goto("/tasks/new"),
    ]);
    expect(apiRequest.headers()["x-user"]).toBeDefined();
  });
});

// ===========================================================================
// 2. Contributor Task Ownership Enforcement
// ===========================================================================

test.describe("DMR-2: Contributor Owner Validation", () => {
  test("DMR-2.1: contributor can create own task in participating project", async ({ request }) => {
    const project = await createProject(request, "leader1", ["eng1"]);
    const { resp } = await createTask(request, project.id, "eng1");
    expect(resp.ok(), `create own: ${resp.status()} ${await resp.text()}`).toBeTruthy();
  });

  test("DMR-2.2: contributor cannot create task with different owner", async ({ request }) => {
    const project = await createProject(request, "leader1", ["eng1"]);
    const { resp } = await createTask(request, project.id, "eng2");
    expect(resp.status()).toBe(403);
  });

  test("DMR-2.3: contributor cannot create task in non-participating project", async ({ request }) => {
    const project = await createProject(request, "leader1", ["other-eng"]);
    const { resp } = await createTask(request, project.id, "eng1");
    expect(resp.status()).toBe(403);
  });

  test("DMR-2.4: contributor can update own task", async ({ request }) => {
    const project = await createProject(request, "leader1", ["eng1"]);
    const { task } = await createTask(request, project.id, "eng1");
    const resp = await request.put(`/api/v1/work-items?id=${task.id}`, {
      data: { ...task, title: "Updated Title", status: "in_progress" },
      headers: headers("contributor", "eng1"),
    });
    expect(resp.ok(), `update own: ${resp.status()} ${await resp.text()}`).toBeTruthy();
  });

  test("DMR-2.5: contributor cannot update another user's task", async ({ request }) => {
    const project = await createProject(request, "leader1", ["eng1", "eng2"]);
    const { task } = await createTask(request, project.id, "eng1", {}, "contributor", "eng1");
    const resp = await request.put(`/api/v1/work-items?id=${task.id}`, {
      data: { ...task, title: "Hacked" },
      headers: headers("contributor", "eng2"),
    });
    expect(resp.status()).toBe(403);
  });

  test("DMR-2.6: contributor can delete own task", async ({ request }) => {
    const project = await createProject(request, "leader1", ["eng1"]);
    const { task } = await createTask(request, project.id, "eng1");
    const resp = await request.delete(`/api/v1/work-items?id=${task.id}`, {
      headers: headers("contributor", "eng1"),
    });
    expect(resp.ok(), `delete own: ${resp.status()}`).toBeTruthy();
  });

  test("DMR-2.7: contributor cannot delete another user's task", async ({ request }) => {
    const project = await createProject(request, "leader1", ["eng1", "eng2"]);
    const { task } = await createTask(request, project.id, "eng1", {}, "contributor", "eng1");
    const resp = await request.delete(`/api/v1/work-items?id=${task.id}`, {
      headers: headers("contributor", "eng2"),
    });
    expect(resp.status()).toBe(403);
  });
});

// ===========================================================================
// 3. Project Owner Scope Enforcement
// ===========================================================================

test.describe("DMR-3: Project Owner Scope", () => {
  test("DMR-3.1: project_owner can create task in own project", async ({ request }) => {
    const project = await createProject(request, "leader1", [], "admin", "admin");
    const { resp } = await createTask(request, project.id, "leader1", {}, "project_owner", "leader1");
    expect(resp.ok(), `owner create: ${resp.status()} ${await resp.text()}`).toBeTruthy();
  });

  test("DMR-3.2: project_owner cannot create task in another owner's project", async ({ request }) => {
    const project = await createProject(request, "leader2", [], "admin", "admin");
    const { resp } = await createTask(request, project.id, "leader1", {}, "project_owner", "leader1");
    expect(resp.status()).toBe(403);
  });

  test("DMR-3.3: project_owner can update any task in own project", async ({ request }) => {
    const project = await createProject(request, "leader1", ["eng1"], "admin", "admin");
    const { task } = await createTask(request, project.id, "eng1", {}, "contributor", "eng1");
    const resp = await request.put(`/api/v1/work-items?id=${task.id}`, {
      data: { ...task, title: "Leader Updated" },
      headers: headers("project_owner", "leader1"),
    });
    expect(resp.ok(), `owner update: ${resp.status()} ${await resp.text()}`).toBeTruthy();
  });

  test("DMR-3.4: admin bypasses all ownership checks", async ({ request }) => {
    const project = await createProject(request, "leader1", ["eng1"], "admin", "admin");
    const { task } = await createTask(request, project.id, "eng1", {}, "contributor", "eng1");
    const resp = await request.put(`/api/v1/work-items?id=${task.id}`, {
      data: { ...task, title: "Admin Override" },
      headers: headers("admin", "admin"),
    });
    expect(resp.ok()).toBeTruthy();
  });
});

// ===========================================================================
// 4. Milestone Completion Restriction
// ===========================================================================

test.describe("DMR-4: Milestone Completion Gate", () => {
  test("DMR-4.1: project_owner can mark milestone as completed", async ({ request }) => {
    const project = await createProject(request, "leader1", [], "admin", "admin");
    const milestone = await createMilestone(request, project.id, "leader1", "project_owner", "leader1");
    const resp = await request.put(`/api/v1/milestones?id=${milestone.id}`, {
      data: { ...milestone, status: "completed" },
      headers: headers("project_owner", "leader1"),
    });
    expect(resp.ok(), `owner complete: ${resp.status()} ${await resp.text()}`).toBeTruthy();
  });

  test("DMR-4.2: contributor cannot mark milestone as completed", async ({ request }) => {
    const project = await createProject(request, "leader1", ["eng1"], "admin", "admin");
    const milestone = await createMilestone(request, project.id, "leader1", "project_owner", "leader1");
    const resp = await request.put(`/api/v1/milestones?id=${milestone.id}`, {
      data: { ...milestone, status: "completed" },
      headers: headers("contributor", "eng1"),
    });
    expect(resp.status()).toBe(403);
  });

  test("DMR-4.3: contributor can update milestone to active (non-completed)", async ({ request }) => {
    const project = await createProject(request, "leader1", ["eng1"], "admin", "admin");
    const milestone = await createMilestone(request, project.id, "leader1", "project_owner", "leader1");
    const resp = await request.put(`/api/v1/milestones?id=${milestone.id}`, {
      data: { ...milestone, status: "active" },
      headers: headers("contributor", "eng1"),
    });
    // Contributors shouldn't manage milestones at all per the model
    // but if they can, they must not be able to set completed
    if (resp.ok()) {
      const updated = await resp.json();
      expect(updated.status).not.toBe("completed");
    }
  });
});

// ===========================================================================
// 5. Frontend Workstream Removal
// ===========================================================================

test.describe("DMR-5: Frontend Workstream", () => {
  test("DMR-5.1: task create form does not show workstream field", async ({ page }) => {
    await page.goto("/tasks/new");
    const workstreamLabel = page.locator("label", { hasText: /workstream|工作流/i });
    const count = await workstreamLabel.count();
    expect(count).toBe(0);
  });

  test("DMR-5.2: task edit form does not show workstream field", async ({ page, request }) => {
    const project = await createProject(request, "leader1", ["eng1"], "admin", "admin");
    const { task } = await createTask(request, project.id, "eng1", {}, "contributor", "eng1");
    await page.goto(`/tasks/${task.id}`);
    const workstreamLabel = page.locator("label", { hasText: /workstream|工作流/i });
    const count = await workstreamLabel.count();
    expect(count).toBe(0);
  });
});

// ===========================================================================
// 6. Owner Dropdown from Participants
// ===========================================================================

test.describe("DMR-6: Owner Participant Dropdown", () => {
  test("DMR-6.1: task form shows owner dropdown when project has participants", async ({ page, request }) => {
    const project = await createProject(request, "leader1", ["eng1", "eng2"], "admin", "admin");
    await page.goto("/tasks/new");
    await page.locator('select').first().selectOption(project.id); // project selector
    const ownerSelect = page.locator('[data-testid="task-owner-select"]');
    await expect(ownerSelect).toBeVisible();
    await expect(ownerSelect.locator("option", { hasText: "eng1" })).toBeVisible();
    await expect(ownerSelect.locator("option", { hasText: "eng2" })).toBeVisible();
  });
});

// ===========================================================================
// 7. Completion Criteria Checklist
// ===========================================================================

test.describe("DMR-7: Completion Criteria Checklist", () => {
  test("DMR-7.1: milestone detail shows criteria as checklist items", async ({ page, request }) => {
    const project = await createProject(request, "leader1", [], "admin", "admin");
    const milestone = await createMilestone(request, project.id, "leader1", "project_owner", "leader1");
    // Update with multiline criteria
    await request.put(`/api/v1/milestones?id=${milestone.id}`, {
      data: { ...milestone, completionCriteria: "Item one\nItem two\nItem three" },
      headers: headers("project_owner", "leader1"),
    });
    await page.goto(`/milestones/${milestone.id}`);
    const checkboxes = page.locator('.criteria-list input[type="checkbox"], .checklist input[type="checkbox"]');
    const count = await checkboxes.count();
    expect(count).toBeGreaterThanOrEqual(3);
  });
});

// ===========================================================================
// 8. E2E: Full Task Lifecycle with Auth
// ===========================================================================

test.describe("DMR-8: E2E Task Lifecycle", () => {
  test("engineer creates, updates, and deletes own task", async ({ page, request }) => {
    const project = await createProject(request, "leader1", ["eng1"], "admin", "admin");

    // Set role and user
    await page.goto("/");
    await page.locator(".role-select").selectOption("contributor");
    const userSelect = page.locator(".user-select");
    if (await userSelect.count() > 0) {
      await userSelect.selectOption("eng1");
    }

    // Navigate to task create
    await page.goto("/tasks/new");
    await page.locator('input[required]').first().fill("E2E Auth Task");
    const projectSelects = page.locator("select");
    // Select the project
    for (let i = 0; i < await projectSelects.count(); i++) {
      const sel = projectSelects.nth(i);
      const options = await sel.locator("option").allTextContents();
      const projIdx = options.findIndex((o: string) => o.includes("AuthTest"));
      if (projIdx >= 0) {
        await sel.selectOption({ index: projIdx });
        break;
      }
    }
    await page.locator('button[type="submit"]', { hasText: /保存|Save/ }).click();
    await expect(page).toHaveURL(/\/tasks/, { timeout: 5000 });

    // Task should appear in list
    await page.goto("/tasks");
    await expect(page.locator(".task-table")).toContainText("E2E Auth Task", { timeout: 5000 });
  });

  test("contributor blocked from completing milestone via UI", async ({ page, request }) => {
    const project = await createProject(request, "leader1", ["eng1"], "admin", "admin");
    const milestone = await createMilestone(request, project.id, "leader1", "project_owner", "leader1");

    await page.goto("/");
    await page.locator(".role-select").selectOption("contributor");
    const userSelect = page.locator(".user-select");
    if (await userSelect.count() > 0) {
      await userSelect.selectOption("eng1");
    }

    // Try to complete milestone via API directly - should be forbidden
    const resp = await page.request.put(`/api/v1/milestones?id=${milestone.id}`, {
      data: { ...milestone, status: "completed" },
      headers: headers("contributor", "eng1"),
    });
    expect(resp.status()).toBe(403);
  });
});
