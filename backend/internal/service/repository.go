package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"goal-manager/backend/internal/domain"

	_ "github.com/go-sql-driver/mysql"
)

const stateKey = "domain"

// Repository persists the service-owned domain state. The service remains the
// owner of validation, RBAC, and rollup rules; repositories only load/save data.
type Repository interface {
	Load(context.Context) (State, error)
	Save(context.Context, State) error
	Name() string
	Durable() bool
}

type State struct {
	RoadmapPeriods map[string]domain.RoadmapPeriod     `json:"roadmapPeriods"`
	RoadmapItems   map[string]domain.RoadmapItem       `json:"roadmapItems"`
	Projects       map[string]domain.Project           `json:"projects"`
	Milestones     map[string]domain.Milestone         `json:"milestones"`
	Workstreams    map[string]domain.Workstream        `json:"workstreams"`
	WorkItems      map[string]domain.LinkedWorkItem    `json:"workItems"`
	Updates        map[string]domain.WeeklyUpdate      `json:"updates"`
	GitLabConfigs  map[string]domain.GitLabConfig      `json:"gitlabConfigs"`
	SyncRules      map[string]domain.SyncRule          `json:"syncRules"`
	SyncJobs       map[string]domain.SyncJob           `json:"syncJobs"`
	SyncFailures   map[string]domain.SyncFailure       `json:"syncFailures"`
	Notifications  map[string]domain.NotificationEvent `json:"notifications"`
	Alerts         map[string]domain.Alert             `json:"alerts"`
	Sequence       int                                 `json:"sequence"`
}

func emptyState() State {
	return State{
		RoadmapPeriods: map[string]domain.RoadmapPeriod{},
		RoadmapItems:   map[string]domain.RoadmapItem{},
		Projects:       map[string]domain.Project{},
		Milestones:     map[string]domain.Milestone{},
		Workstreams:    map[string]domain.Workstream{},
		WorkItems:      map[string]domain.LinkedWorkItem{},
		Updates:        map[string]domain.WeeklyUpdate{},
		GitLabConfigs:  map[string]domain.GitLabConfig{},
		SyncRules:      map[string]domain.SyncRule{},
		SyncJobs:       map[string]domain.SyncJob{},
		SyncFailures:   map[string]domain.SyncFailure{},
		Notifications:  map[string]domain.NotificationEvent{},
		Alerts:         map[string]domain.Alert{},
	}
}

func normalizeState(state State) State {
	if state.RoadmapPeriods == nil {
		state.RoadmapPeriods = map[string]domain.RoadmapPeriod{}
	}
	if state.RoadmapItems == nil {
		state.RoadmapItems = map[string]domain.RoadmapItem{}
	}
	if state.Projects == nil {
		state.Projects = map[string]domain.Project{}
	}
	if state.Milestones == nil {
		state.Milestones = map[string]domain.Milestone{}
	}
	if state.Workstreams == nil {
		state.Workstreams = map[string]domain.Workstream{}
	}
	if state.WorkItems == nil {
		state.WorkItems = map[string]domain.LinkedWorkItem{}
	}
	if state.Updates == nil {
		state.Updates = map[string]domain.WeeklyUpdate{}
	}
	if state.GitLabConfigs == nil {
		state.GitLabConfigs = map[string]domain.GitLabConfig{}
	}
	if state.SyncRules == nil {
		state.SyncRules = map[string]domain.SyncRule{}
	}
	if state.SyncJobs == nil {
		state.SyncJobs = map[string]domain.SyncJob{}
	}
	if state.SyncFailures == nil {
		state.SyncFailures = map[string]domain.SyncFailure{}
	}
	if state.Notifications == nil {
		state.Notifications = map[string]domain.NotificationEvent{}
	}
	if state.Alerts == nil {
		state.Alerts = map[string]domain.Alert{}
	}
	return state
}

type MemoryRepository struct{}

func NewMemoryRepository() *MemoryRepository                    { return &MemoryRepository{} }
func (r *MemoryRepository) Load(context.Context) (State, error) { return emptyState(), nil }
func (r *MemoryRepository) Save(context.Context, State) error   { return nil }
func (r *MemoryRepository) Name() string                        { return "memory" }
func (r *MemoryRepository) Durable() bool                       { return false }

type MySQLRepository struct {
	db *sql.DB
}

func NewMySQLRepository(ctx context.Context, dsn string) (*MySQLRepository, error) {
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}
	if err := waitForMySQL(ctx, db, 60*time.Second); err != nil {
		db.Close()
		return nil, err
	}
	repo := &MySQLRepository{db: db}
	if err := repo.ensureSchema(ctx); err != nil {
		db.Close()
		return nil, err
	}
	return repo, nil
}

func waitForMySQL(ctx context.Context, db *sql.DB, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	var lastErr error
	for {
		if err := db.PingContext(ctx); err == nil {
			return nil
		} else {
			lastErr = err
		}
		if time.Now().After(deadline) {
			return fmt.Errorf("mysql not ready after %s: %w", timeout, lastErr)
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(time.Second):
		}
	}
}

func (r *MySQLRepository) Load(ctx context.Context) (State, error) {
	var payload []byte
	err := r.db.QueryRowContext(ctx, `SELECT payload FROM app_state WHERE state_key = ?`, stateKey).Scan(&payload)
	if err == sql.ErrNoRows {
		return emptyState(), nil
	}
	if err != nil {
		return State{}, err
	}
	state := emptyState()
	if len(payload) > 0 {
		if err := json.Unmarshal(payload, &state); err != nil {
			return State{}, err
		}
	}
	return normalizeState(state), nil
}

func (r *MySQLRepository) Save(ctx context.Context, state State) error {
	payload, err := json.Marshal(normalizeState(state))
	if err != nil {
		return err
	}
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	_, err = tx.ExecContext(ctx, `INSERT INTO app_state (state_key, payload, updated_at) VALUES (?, ?, ?) ON DUPLICATE KEY UPDATE payload = VALUES(payload), updated_at = VALUES(updated_at)`, stateKey, payload, time.Now().UTC())
	if err != nil {
		tx.Rollback()
		return err
	}
	if err := tx.Commit(); err != nil {
		return err
	}
	return nil
}

func (r *MySQLRepository) Name() string  { return "mysql" }
func (r *MySQLRepository) Durable() bool { return true }

func (r *MySQLRepository) ensureSchema(ctx context.Context) error {
	statements := []string{
		`CREATE TABLE IF NOT EXISTS app_state (state_key VARCHAR(64) PRIMARY KEY, payload JSON NOT NULL, updated_at DATETIME NOT NULL)`,
		`CREATE TABLE IF NOT EXISTS id_sequences (prefix VARCHAR(32) PRIMARY KEY, last_val BIGINT NOT NULL DEFAULT 0)`,
	}
	for _, stmt := range statements {
		if _, err := r.db.ExecContext(ctx, stmt); err != nil {
			return fmt.Errorf("ensure mysql repository schema: %w", err)
		}
	}
	return nil
}
