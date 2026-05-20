CREATE DATABASE IF NOT EXISTS goal_manager CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
USE goal_manager;

CREATE TABLE IF NOT EXISTS app_state (
  state_key VARCHAR(64) PRIMARY KEY,
  payload JSON NOT NULL,
  updated_at DATETIME NOT NULL
);

CREATE TABLE IF NOT EXISTS id_sequences (
  prefix VARCHAR(32) PRIMARY KEY,
  last_val BIGINT NOT NULL DEFAULT 0
);

CREATE TABLE IF NOT EXISTS roadmap_periods (
  id VARCHAR(64) PRIMARY KEY,
  title VARCHAR(255) NOT NULL,
  description TEXT NOT NULL,
  owner VARCHAR(128) NOT NULL,
  status VARCHAR(32) NOT NULL,
  priority VARCHAR(32) NOT NULL,
  period_start DATETIME NOT NULL,
  period_end DATETIME NOT NULL,
  created_at DATETIME NOT NULL,
  updated_at DATETIME NOT NULL
);

CREATE TABLE IF NOT EXISTS roadmap_items (
  id VARCHAR(64) PRIMARY KEY,
  period_id VARCHAR(64) NOT NULL,
  title VARCHAR(255) NOT NULL,
  description TEXT NOT NULL,
  owner VARCHAR(128) NOT NULL,
  priority VARCHAR(32) NOT NULL,
  status VARCHAR(32) NOT NULL,
  created_at DATETIME NOT NULL,
  updated_at DATETIME NOT NULL,
  CONSTRAINT fk_roadmap_items_period FOREIGN KEY (period_id) REFERENCES roadmap_periods(id)
);

CREATE TABLE IF NOT EXISTS projects (
  id VARCHAR(64) PRIMARY KEY,
  roadmap_item_id VARCHAR(64) NOT NULL,
  name VARCHAR(255) NOT NULL,
  summary TEXT NOT NULL,
  objective TEXT NOT NULL,
  owner VARCHAR(128) NOT NULL,
  participants TEXT NOT NULL,
  project_type VARCHAR(64) NOT NULL,
  status VARCHAR(32) NOT NULL,
  health_status VARCHAR(32) NOT NULL,
  target_start_date DATETIME NULL,
  target_end_date DATETIME NULL,
  actual_end_date DATETIME NULL,
  priority VARCHAR(32) NOT NULL,
  tags TEXT NOT NULL,
  created_at DATETIME NOT NULL,
  updated_at DATETIME NOT NULL,
  CONSTRAINT fk_projects_roadmap_item FOREIGN KEY (roadmap_item_id) REFERENCES roadmap_items(id)
);

CREATE TABLE IF NOT EXISTS milestones (
  id VARCHAR(64) PRIMARY KEY,
  project_id VARCHAR(64) NOT NULL,
  title VARCHAR(255) NOT NULL,
  milestone_type VARCHAR(64) NOT NULL,
  description TEXT NOT NULL,
  completion_criteria TEXT NOT NULL,
  owner VARCHAR(128) NOT NULL,
  planned_date DATETIME NULL,
  forecast_date DATETIME NULL,
  completed_date DATETIME NULL,
  status VARCHAR(32) NOT NULL,
  health_status VARCHAR(32) NOT NULL,
  progress_percent INT NOT NULL,
  risk_level VARCHAR(32) NOT NULL,
  dependency_summary TEXT NOT NULL,
  created_at DATETIME NOT NULL,
  updated_at DATETIME NOT NULL,
  CONSTRAINT fk_milestones_project FOREIGN KEY (project_id) REFERENCES projects(id)
);

CREATE TABLE IF NOT EXISTS workstreams (
  id VARCHAR(64) PRIMARY KEY,
  project_id VARCHAR(64) NOT NULL,
  milestone_id VARCHAR(64) NOT NULL,
  name VARCHAR(255) NOT NULL,
  owner VARCHAR(128) NOT NULL,
  status VARCHAR(32) NOT NULL,
  description TEXT NOT NULL,
  created_at DATETIME NOT NULL,
  updated_at DATETIME NOT NULL,
  CONSTRAINT fk_workstreams_project FOREIGN KEY (project_id) REFERENCES projects(id),
  CONSTRAINT fk_workstreams_milestone FOREIGN KEY (milestone_id) REFERENCES milestones(id)
);

CREATE TABLE IF NOT EXISTS linked_work_items (
  id VARCHAR(64) PRIMARY KEY,
  source_type VARCHAR(32) NOT NULL,
  source_id VARCHAR(128) NULL,
  source_url VARCHAR(255) NULL,
  title VARCHAR(255) NOT NULL,
  project_id VARCHAR(64) NULL,
  milestone_id VARCHAR(64) NULL,
  workstream_id VARCHAR(64) NULL,
  owner VARCHAR(128) NOT NULL,
  status VARCHAR(32) NOT NULL,
  estimate VARCHAR(64) NOT NULL,
  due_date DATETIME NULL,
  blocked TINYINT(1) NOT NULL DEFAULT 0,
  created_at DATETIME NOT NULL,
  updated_at DATETIME NOT NULL,
  CONSTRAINT fk_work_items_project FOREIGN KEY (project_id) REFERENCES projects(id)
);

CREATE TABLE IF NOT EXISTS weekly_updates (
  id VARCHAR(64) PRIMARY KEY,
  project_id VARCHAR(64) NOT NULL,
  milestone_id VARCHAR(64) NULL,
  author VARCHAR(128) NOT NULL,
  week VARCHAR(32) NOT NULL,
  summary TEXT NOT NULL,
  progress TEXT NOT NULL,
  risk TEXT NOT NULL,
  blockers TEXT NOT NULL,
  decisions_needed TEXT NOT NULL,
  next_steps TEXT NOT NULL,
  created_at DATETIME NOT NULL,
  updated_at DATETIME NOT NULL,
  CONSTRAINT fk_weekly_updates_project FOREIGN KEY (project_id) REFERENCES projects(id)
);
