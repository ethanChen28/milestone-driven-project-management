import { createRouter, createWebHistory } from "vue-router";

const router = createRouter({
  history: createWebHistory(),
  routes: [
    { path: "/", name: "dashboard", component: () => import("./views/DashboardView.vue") },
    { path: "/projects", name: "projects", component: () => import("./views/ProjectsView.vue") },
    { path: "/projects/:id", name: "project-detail", component: () => import("./views/ProjectDetailView.vue") },
    { path: "/tasks", name: "tasks", component: () => import("./views/TasksView.vue") },
    { path: "/tasks/new", name: "task-create", component: () => import("./views/TaskDetailView.vue") },
    { path: "/tasks/:id", name: "task-detail", component: () => import("./views/TaskDetailView.vue") },
    { path: "/milestones", name: "milestones", component: () => import("./views/MilestonesView.vue") },
    { path: "/milestones/:id", name: "milestone-detail", component: () => import("./views/MilestoneDetailView.vue") },
    { path: "/roadmap", name: "roadmap", component: () => import("./views/RoadmapView.vue") },
    { path: "/review", name: "review", component: () => import("./views/ReviewView.vue") },
  ],
});

export default router;
