USE goal_manager;

-- GitLab sync columns are defined in 001_schema.sql for clean installs.

CREATE TABLE IF NOT EXISTS gitlab_configs (
  id VARCHAR(64) PRIMARY KEY,
  name VARCHAR(255) NOT NULL,
  base_url VARCHAR(512) NOT NULL,
  access_token VARCHAR(512) NOT NULL DEFAULT '',
  `group` VARCHAR(255) NOT NULL DEFAULT '',
  repository VARCHAR(255) NOT NULL DEFAULT '',
  created_at DATETIME NOT NULL,
  updated_at DATETIME NOT NULL
);

CREATE TABLE IF NOT EXISTS sync_rules (
  id VARCHAR(64) PRIMARY KEY,
  gitlab_config_id VARCHAR(64) NOT NULL,
  project_id VARCHAR(64) NOT NULL,
  milestone_id VARCHAR(64) NULL,
  label VARCHAR(255) NULL,
  assignee VARCHAR(128) NULL,
  gitlab_milestone VARCHAR(255) NULL,
  query_filter TEXT NULL,
  enabled TINYINT(1) NOT NULL DEFAULT 1,
  created_at DATETIME NOT NULL,
  updated_at DATETIME NOT NULL,
  CONSTRAINT fk_sync_rules_gitlab_config FOREIGN KEY (gitlab_config_id) REFERENCES gitlab_configs(id),
  CONSTRAINT fk_sync_rules_project FOREIGN KEY (project_id) REFERENCES projects(id)
);

CREATE TABLE IF NOT EXISTS sync_jobs (
  id VARCHAR(64) PRIMARY KEY,
  rule_id VARCHAR(64) NOT NULL,
  status VARCHAR(32) NOT NULL,
  started_at DATETIME NULL,
  completed_at DATETIME NULL,
  items_synced INT NOT NULL DEFAULT 0,
  items_failed INT NOT NULL DEFAULT 0,
  error_message TEXT NULL,
  created_at DATETIME NOT NULL,
  updated_at DATETIME NOT NULL,
  CONSTRAINT fk_sync_jobs_rule FOREIGN KEY (rule_id) REFERENCES sync_rules(id)
);

CREATE TABLE IF NOT EXISTS sync_failures (
  id VARCHAR(64) PRIMARY KEY,
  work_item_id VARCHAR(64) NOT NULL,
  source_id VARCHAR(128) NOT NULL,
  error TEXT NOT NULL,
  retry_count INT NOT NULL DEFAULT 0,
  last_attempt DATETIME NOT NULL,
  resolved TINYINT(1) NOT NULL DEFAULT 0,
  created_at DATETIME NOT NULL,
  updated_at DATETIME NOT NULL
);

CREATE TABLE IF NOT EXISTS notification_events (
  id VARCHAR(64) PRIMARY KEY,
  event_type VARCHAR(64) NOT NULL,
  target VARCHAR(128) NOT NULL,
  channel VARCHAR(32) NOT NULL,
  title VARCHAR(255) NOT NULL,
  message TEXT NOT NULL,
  delivered TINYINT(1) NOT NULL DEFAULT 0,
  created_at DATETIME NOT NULL,
  updated_at DATETIME NOT NULL
);

CREATE TABLE IF NOT EXISTS alerts (
  id VARCHAR(64) PRIMARY KEY,
  alert_type VARCHAR(64) NOT NULL,
  target_id VARCHAR(64) NOT NULL,
  target_type VARCHAR(32) NOT NULL,
  message TEXT NOT NULL,
  dismissed TINYINT(1) NOT NULL DEFAULT 0,
  created_at DATETIME NOT NULL,
  updated_at DATETIME NOT NULL
);
