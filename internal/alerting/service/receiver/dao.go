package receiver

import (
	"context"
	"fmt"
	"time"

	adb "github.com/qiniu/zeroops/internal/alerting/database"
)

type AlertIssueDAO interface {
	InsertAlertIssue(ctx context.Context, r *AlertIssueRow) error
}

// ServiceStateWriter optionally allows writing to service_states table.
type ServiceStateWriter interface {
	UpsertServiceState(ctx context.Context, service, version string, reportAt time.Time, healthState string) error
}

type NoopDAO struct{}

func NewNoopDAO() *NoopDAO { return &NoopDAO{} }

func (d *NoopDAO) InsertAlertIssue(ctx context.Context, r *AlertIssueRow) error { return nil }

func (d *NoopDAO) UpsertServiceState(ctx context.Context, service, version string, reportAt time.Time, healthState string) error {
	return nil
}

type PgDAO struct{ DB *adb.Database }

func NewPgDAO(db *adb.Database) *PgDAO { return &PgDAO{DB: db} }

func (d *PgDAO) InsertAlertIssue(ctx context.Context, r *AlertIssueRow) error {
	const q = `
	INSERT INTO alert_issues
		(id, state, level, alert_state, title, labels, alert_since)
	VALUES
		($1, $2, $3, $4, $5, $6, $7)
	`
	if _, err := d.DB.ExecContext(ctx, q, r.ID, r.State, r.Level, r.AlertState, r.Title, r.LabelJSON, r.AlertSince); err != nil {
		return fmt.Errorf("insert alert_issue: %w", err)
	}
	return nil
}

// UpsertServiceState inserts or updates service_states with health_state and earliest report_at.
// detail, resolved_at, correlation_id remain empty/unchanged.
func (d *PgDAO) UpsertServiceState(ctx context.Context, service, version string, reportAt time.Time, healthState string) error {
	const q = `
	INSERT INTO service_states (service, version, report_at, health_state)
	VALUES ($1, $2, $3, $4)
	ON CONFLICT (service, version) DO UPDATE
	SET health_state = EXCLUDED.health_state,
	    report_at = LEAST(service_states.report_at, EXCLUDED.report_at)
	`
	if _, err := d.DB.ExecContext(ctx, q, service, version, reportAt, healthState); err != nil {
		return fmt.Errorf("upsert service_state: %w", err)
	}
	return nil
}
