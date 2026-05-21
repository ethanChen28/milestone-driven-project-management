import { test, expect } from "@playwright/test";

const headers = { "Content-Type": "application/json", "X-Role": "admin", "X-User": "admin" };

async function createProjectSpaceFixture(request: any) {
  const projectResp = await request.post("/api/v1/projects", {
    data: { name: `Project Space ${Date.now()}`, owner: "tester", participants: ["tester"], status: "active", healthStatus: "on_track", priority: "P1" },
    headers,
  });
  expect(projectResp.ok(), await projectResp.text()).toBeTruthy();
  const project = await projectResp.json();

  const milestoneResp = await request.post("/api/v1/milestones", {
    data: { projectId: project.id, title: "Space Milestone", owner: "tester", status: "blocked", healthStatus: "at_risk", riskLevel: "high", dependencySummary: "Waiting on vendor", completionCriteria: "Done" },
    headers,
  });
  expect(milestoneResp.ok(), await milestoneResp.text()).toBeTruthy();
  const milestone = await milestoneResp.json();

  const workResp = await request.post("/api/v1/work-items", {
    data: { sourceType: "external_dependency", sourceId: "vendor-1", sourceUrl: "https://vendor.example/ticket/1", title: "Vendor API", projectId: project.id, milestoneId: milestone.id, owner: "tester", status: "blocked", priority: "P0", blocked: true },
    headers,
  });
  expect(workResp.ok(), await workResp.text()).toBeTruthy();

  const updateResp = await request.post("/api/v1/weekly-updates", {
    data: { projectId: project.id, milestoneId: milestone.id, author: "tester", week: "2026-W21", summary: "Risk update", risk: "Vendor delay", blockers: "Vendor API", decisionsNeeded: "Approve fallback" },
    headers,
  });
  expect(updateResp.ok(), await updateResp.text()).toBeTruthy();

  return { project, milestone };
}

test.describe("Project Space", () => {
  test("opens project space tabs and milestone quick filter", async ({ page, request }) => {
    const { project, milestone } = await createProjectSpaceFixture(request);

    await page.goto(`/projects/${project.id}`);
    await expect(page.locator("h1")).toHaveText(project.name);
    await expect(page.locator(".space-tabs")).toContainText("概览");
    await expect(page.locator(".milestone-card")).toContainText("Space Milestone");

    await page.locator(".milestone-card .btn", { hasText: "查看工作项" }).click({ force: true });
    await expect(page).toHaveURL(new RegExp(`tab=work-items.*milestoneId=${milestone.id}|milestoneId=${milestone.id}.*tab=work-items`));
    await expect(page.locator(".work-card")).toContainText("Vendor API");

    await page.locator(".tab", { hasText: "风险" }).click({ force: true });
    await expect(page.locator(".risk-card").first()).toContainText("Space Milestone");

    await page.locator(".tab", { hasText: "依赖" }).click({ force: true });
    await expect(page.locator(".risk-card").first()).toContainText("Waiting on vendor");
  });

  test("project-space API returns rollups and derived relationship signals", async ({ request }) => {
    const { project } = await createProjectSpaceFixture(request);
    const resp = await request.get(`/api/v1/project-space?id=${project.id}`);
    expect(resp.ok(), await resp.text()).toBeTruthy();
    const body = await resp.json();
    expect(body.rollups.blockedMilestones).toBeGreaterThanOrEqual(1);
    expect(body.rollups.blockedWorkItems).toBeGreaterThanOrEqual(1);
    expect(body.risks.length).toBeGreaterThanOrEqual(3);
    expect(body.dependencies.length).toBeGreaterThanOrEqual(2);
  });
});
