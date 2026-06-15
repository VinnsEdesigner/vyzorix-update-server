// Package storage provides SQLite database operations.
package storage

import (
	"context"
	"database/sql"
	"errors"
	"time"
)

// CommandStatus represents the status of a command dispatch.
type CommandStatus struct {
	DispatchID  string     `json:"dispatchId"`
	DeviceID    string     `json:"deviceId"`
	Command     string     `json:"command"`
	Args        string     `json:"args,omitempty"`
	Status      string     `json:"status"`
	Delivery    string     `json:"delivery"`
	CreatedAt   time.Time  `json:"createdAt"`
	DeliveredAt *time.Time `json:"deliveredAt,omitempty"`
	CompletedAt *time.Time `json:"completedAt,omitempty"`
	Result      string     `json:"result,omitempty"`
	WakeError   string     `json:"wakeError,omitempty"`
}

// SaveCommand saves a new command dispatch.
func (s *Store) SaveCommand(ctx context.Context, dispatchID, deviceID, command string, args []byte, delivery string) error {
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO commands(dispatch_id,device_id,command,args,delivery,created_at) VALUES(?,?,?,?,?,?)`,
		dispatchID, deviceID, command, string(args), delivery, time.Now().UnixMilli(),
	)
	return err
}

// MarkWake marks whether a wake command was sent successfully.
func (s *Store) MarkWake(ctx context.Context, dispatchID string, errText string) error {
	wakeSent := 1
	if errText != "" {
		wakeSent = 0
	}
	_, err := s.db.ExecContext(ctx,
		`UPDATE commands SET wake_sent=?, wake_error=? WHERE dispatch_id=?`,
		wakeSent, errText, dispatchID,
	)
	return err
}

// MarkDelivered marks a command as delivered.
func (s *Store) MarkDelivered(ctx context.Context, dispatchID string) error {
	_, err := s.db.ExecContext(ctx,
		`UPDATE commands SET delivery='sent', delivered_at=?, status='sent' WHERE dispatch_id=?`,
		time.Now().UnixMilli(), dispatchID,
	)
	return err
}

// GetCommandStatus retrieves the status of a command dispatch.
func (s *Store) GetCommandStatus(ctx context.Context, dispatchID string) (*CommandStatus, error) {
	var cs CommandStatus
	var deliveredAt, completedAt sql.NullInt64

	err := s.db.QueryRowContext(ctx, `
		SELECT dispatch_id, device_id, command, args, status, delivery,
		       created_at, delivered_at, completed_at, result, wake_error
		FROM commands WHERE dispatch_id = ?
	`, dispatchID).Scan(
		&cs.DispatchID, &cs.DeviceID, &cs.Command, &cs.Args,
		&cs.Status, &cs.Delivery, &cs.CreatedAt,
		&deliveredAt, &completedAt, &cs.Result, &cs.WakeError,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	if deliveredAt.Valid {
		t := time.UnixMilli(deliveredAt.Int64).UTC()
		cs.DeliveredAt = &t
	}
	if completedAt.Valid {
		t := time.UnixMilli(completedAt.Int64).UTC()
		cs.CompletedAt = &t
	}

	return &cs, nil
}

// UpdateCommandStatus updates the status of a command.
func (s *Store) UpdateCommandStatus(ctx context.Context, dispatchID, status string) error {
	_, err := s.db.ExecContext(ctx,
		`UPDATE commands SET status=? WHERE dispatch_id=?`,
		status, dispatchID,
	)
	return err
}

// MarkCommandCompleted marks a command as completed with result.
func (s *Store) MarkCommandCompleted(ctx context.Context, dispatchID, result string) error {
	_, err := s.db.ExecContext(ctx,
		`UPDATE commands SET status='completed', completed_at=?, result=? WHERE dispatch_id=?`,
		time.Now().UnixMilli(), result, dispatchID,
	)
	return err
}

// MarkCommandFailed marks a command as failed with error.
func (s *Store) MarkCommandFailed(ctx context.Context, dispatchID, errMsg string) error {
	_, err := s.db.ExecContext(ctx,
		`UPDATE commands SET status='failed', completed_at=?, result=? WHERE dispatch_id=?`,
		time.Now().UnixMilli(), errMsg, dispatchID,
	)
	return err
}

// ListCommandsByDevice retrieves commands for a device.
func (s *Store) ListCommandsByDevice(ctx context.Context, deviceID string, limit int) ([]CommandStatus, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT dispatch_id, device_id, command, args, status, delivery,
		        created_at, delivered_at, completed_at, result, COALESCE(wake_error, '')
		 FROM commands WHERE device_id = ? ORDER BY created_at DESC LIMIT ?`,
		deviceID, limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close() //nolint:errcheck

	var commands []CommandStatus
	for rows.Next() {
		var cs CommandStatus
		var deliveredAt, completedAt sql.NullInt64

		if err := rows.Scan(
			&cs.DispatchID, &cs.DeviceID, &cs.Command, &cs.Args,
			&cs.Status, &cs.Delivery, &cs.CreatedAt,
			&deliveredAt, &completedAt, &cs.Result, &cs.WakeError,
		); err != nil {
			return nil, err
		}

		if deliveredAt.Valid {
			t := time.UnixMilli(deliveredAt.Int64).UTC()
			cs.DeliveredAt = &t
		}
		if completedAt.Valid {
			t := time.UnixMilli(completedAt.Int64).UTC()
			cs.CompletedAt = &t
		}

		commands = append(commands, cs)
	}
	return commands, rows.Err()
}

// DeleteOldCommands removes commands older than the given timestamp.
func (s *Store) DeleteOldCommands(ctx context.Context, olderThan int64) (int64, error) {
	result, err := s.db.ExecContext(ctx,
		`DELETE FROM commands WHERE created_at < ?`,
		olderThan,
	)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}