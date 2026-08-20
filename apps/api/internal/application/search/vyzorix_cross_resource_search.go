package search

import (
	"context"
	"database/sql"
	"encoding/json"
	"sort"
)

type Result struct {
	Type    string   `json:"type"`
	ID      string   `json:"id"`
	Title   string   `json:"title"`
	Snippet string   `json:"snippet,omitempty"`
	OrgID   string   `json:"org_id,omitempty"`
	Status  string   `json:"status,omitempty"`
	Tags    []string `json:"tags,omitempty"`
}

type Service struct {
	db *sql.DB
}

func NewService(db *sql.DB) *Service {
	return &Service{db: db}
}

//nolint:gocyclo
func (s *Service) Search(ctx context.Context, query, orgID, resourceType string, limit int, tagFilter, sortBy string) ([]Result, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	pattern := "%" + query + "%"
	var results []Result

	if resourceType == "" || resourceType == "all" || resourceType == "device" {
		results = append(results, s.searchDevices(ctx, orgID, pattern, tagFilter, limit)...)
	}
	if resourceType == "" || resourceType == "all" || resourceType == "command" {
		results = append(results, s.searchCommands(ctx, pattern, limit)...)
	}
	if resourceType == "" || resourceType == "all" || resourceType == "event" {
		results = append(results, s.searchEvents(ctx, pattern, limit)...)
	}
	if resourceType == "" || resourceType == "all" || resourceType == "update" {
		results = append(results, s.searchUpdates(ctx, pattern, limit)...)
	}

	if sortBy != "" {
		sortResults(results, sortBy)
	}

	return results, nil
}

func (s *Service) searchDevices(ctx context.Context, orgID, pattern, tagFilter string, limit int) []Result {
	query := `SELECT id, device_name, model, manufacturer, organization_id, online, tags FROM devices WHERE organization_id = ? AND (id LIKE ? OR device_name LIKE ? OR model LIKE ? OR manufacturer LIKE ? OR tags LIKE ?)`
	args := []any{orgID, pattern, pattern, pattern, pattern, pattern}
	if tagFilter != "" {
		query += ` AND tags LIKE ?`
		args = append(args, "%"+tagFilter+"%")
	}
	query += ` LIMIT ?`
	args = append(args, limit)
	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil
	}
	defer func() { _ = rows.Close() }()

	var results []Result
	for rows.Next() {
		var r Result
		var name, model, manufacturer, tagsJSON sql.NullString
		var online bool
		if err := rows.Scan(&r.ID, &name, &model, &manufacturer, &r.OrgID, &online, &tagsJSON); err != nil {
			continue
		}
		r.Type = "device"
		r.Title = name.String
		if model.String != "" {
			r.Snippet = model.String + " " + manufacturer.String
		}
		if online {
			r.Status = "online"
		} else {
			r.Status = "offline"
		}
		if tagsJSON.Valid {
			_ = json.Unmarshal([]byte(tagsJSON.String), &r.Tags)
		}
		results = append(results, r)
	}
	_ = rows.Err()
	return results
}

func (s *Service) searchCommands(ctx context.Context, pattern string, limit int) []Result {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, device_id, command, status FROM commands WHERE id LIKE ? OR command LIKE ? OR device_id LIKE ? LIMIT ?`,
		pattern, pattern, pattern, limit)
	if err != nil {
		return nil
	}
	defer func() { _ = rows.Close() }()

	var results []Result
	for rows.Next() {
		var r Result
		var deviceID, cmd, status sql.NullString
		if err := rows.Scan(&r.ID, &deviceID, &cmd, &status); err != nil {
			continue
		}
		r.Type = "command"
		r.Title = cmd.String
		r.Snippet = "device: " + deviceID.String
		r.Status = status.String
		results = append(results, r)
	}
	_ = rows.Err()
	return results
}

func (s *Service) searchEvents(ctx context.Context, pattern string, limit int) []Result {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, device_id, event_type FROM device_events WHERE device_id LIKE ? OR event_type LIKE ? LIMIT ?`,
		pattern, pattern, limit)
	if err != nil {
		return nil
	}
	defer func() { _ = rows.Close() }()

	var results []Result
	for rows.Next() {
		var r Result
		var deviceID, eventType sql.NullString
		if err := rows.Scan(&r.ID, &deviceID, &eventType); err != nil {
			continue
		}
		r.Type = "event"
		r.Title = eventType.String
		r.Snippet = "device: " + deviceID.String
		results = append(results, r)
	}
	_ = rows.Err()
	return results
}

func (s *Service) searchUpdates(ctx context.Context, pattern string, limit int) []Result {
	rows, err := s.db.QueryContext(ctx,
		`SELECT version, apk_filename, release_type FROM update_versions WHERE version LIKE ? OR apk_filename LIKE ? LIMIT ?`,
		pattern, pattern, limit)
	if err != nil {
		return nil
	}
	defer func() { _ = rows.Close() }()

	var results []Result
	for rows.Next() {
		var r Result
		var version, filename, releaseType sql.NullString
		if err := rows.Scan(&version, &filename, &releaseType); err != nil {
			continue
		}
		r.Type = "update"
		r.ID = version.String
		r.Title = version.String
		r.Snippet = filename.String + " (" + releaseType.String + ")"
		results = append(results, r)
	}
	_ = rows.Err()
	return results
}

func sortResults(results []Result, sortBy string) {
	switch sortBy {
	case "name":
		sort.Slice(results, func(i, j int) bool {
			return results[i].Title < results[j].Title
		})
	case "type":
		sort.Slice(results, func(i, j int) bool {
			return results[i].Type < results[j].Type
		})
	case "status":
		sort.Slice(results, func(i, j int) bool {
			return results[i].Status < results[j].Status
		})
	}
}
