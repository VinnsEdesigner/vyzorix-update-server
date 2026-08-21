package storage

import (
	"context"
	"database/sql"
	"encoding/json"
	"strings"
	"time"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/annotation"
)

// AnnotationRepository is the SQL persistence for annotations.
type AnnotationRepository struct {
	db *sql.DB
}

// NewAnnotationRepository creates a new AnnotationRepository.
func NewAnnotationRepository(db *sql.DB) *AnnotationRepository {
	return &AnnotationRepository{db: db}
}

const annotationColumns = `id, org_id, title, text, tags, source, start_time, end_time, created_at, updated_at`

func scanAnnotation(scanner interface{ Scan(...any) error }) (*annotation.Annotation, error) {
	var a annotation.Annotation
	var tags, source string
	var createdAt, updatedAt, startTime int64
	var endTime sql.NullInt64
	err := scanner.Scan(
		&a.ID, &a.OrgID, &a.Title, &a.Text, &tags, &source,
		&startTime, &endTime, &createdAt, &updatedAt,
	)
	if err != nil {
		return nil, err
	}
	a.Source = source
	a.StartTime = time.UnixMilli(startTime)
	a.CreatedAt = time.UnixMilli(createdAt)
	a.UpdatedAt = time.UnixMilli(updatedAt)
	if endTime.Valid {
		t := time.UnixMilli(endTime.Int64)
		a.EndTime = &t
	}
	if tags != "" {
		if err := json.Unmarshal([]byte(tags), &a.Tags); err != nil {
			return nil, err
		}
	}
	if a.Tags == nil {
		a.Tags = []string{}
	}
	return &a, nil
}

// Save upserts an annotation.
func (r *AnnotationRepository) Save(ctx context.Context, a *annotation.Annotation) error {
	tagsJSON, err := json.Marshal(a.Tags)
	if err != nil {
		return err
	}
	var endTime interface{}
	if a.EndTime != nil {
		endTime = a.EndTime.UnixMilli()
	}
	query := `
		INSERT INTO annotations (id, org_id, title, text, tags, source, start_time, end_time, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			title = excluded.title,
			text = excluded.text,
			tags = excluded.tags,
			source = excluded.source,
			start_time = excluded.start_time,
			end_time = excluded.end_time,
			updated_at = excluded.updated_at
	`
	_, err = r.db.ExecContext(ctx, query,
		a.ID, a.OrgID, a.Title, a.Text, string(tagsJSON), a.Source,
		a.StartTime.UnixMilli(), endTime, a.CreatedAt.UnixMilli(), a.UpdatedAt.UnixMilli(),
	)
	return err
}

// GetByID returns an annotation or annotation.ErrNotFound.
func (r *AnnotationRepository) GetByID(ctx context.Context, id string) (*annotation.Annotation, error) {
	a, err := scanAnnotation(r.db.QueryRowContext(ctx,
		`SELECT `+annotationColumns+` FROM annotations WHERE id = ?`, id))
	if err == sql.ErrNoRows {
		return nil, annotation.ErrNotFound
	}
	return a, err
}

// List returns annotations matching the filter, newest first.
func (r *AnnotationRepository) List(ctx context.Context, f *annotation.Filter) ([]*annotation.Annotation, error) {
	limit := 200
	if f.Limit > 0 {
		limit = f.Limit
	}
	query := `SELECT ` + annotationColumns + ` FROM annotations WHERE org_id = ?`
	args := []interface{}{f.OrgID}
	if f.Tag != "" {
		query += ` AND tags LIKE ?`
		args = append(args, `%`+strings.ReplaceAll(f.Tag, `%`, `\%`)+`%`)
	}
	if !f.StartTime.IsZero() {
		query += ` AND start_time >= ?`
		args = append(args, f.StartTime.UnixMilli())
	}
	if !f.EndTime.IsZero() {
		query += ` AND start_time <= ?`
		args = append(args, f.EndTime.UnixMilli())
	}
	query += ` ORDER BY start_time DESC LIMIT ?`
	args = append(args, limit)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var annotations []*annotation.Annotation
	for rows.Next() {
		a, err := scanAnnotation(rows)
		if err != nil {
			return nil, err
		}
		annotations = append(annotations, a)
	}
	return annotations, rows.Err()
}

// Delete removes an annotation.
func (r *AnnotationRepository) Delete(ctx context.Context, id string) (bool, error) {
	res, err := r.db.ExecContext(ctx, `DELETE FROM annotations WHERE id = ?`, id)
	if err != nil {
		return false, err
	}
	n, err := res.RowsAffected()
	return n > 0, err
}
