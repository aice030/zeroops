package receiver

import (
	"context"
	"fmt"

	adb "github.com/qiniu/zeroops/internal/alerting/database"
)

type AlertIssueDAO interface {
	InsertAlertIssue(ctx context.Context, r *AlertIssueRow) error
}

type NoopDAO struct{}

func NewNoopDAO() *NoopDAO { return &NoopDAO{} }

func (d *NoopDAO) InsertAlertIssue(ctx context.Context, r *AlertIssueRow) error { return nil }

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
