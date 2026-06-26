package dashboard

// DashboardStatsResponse represents the response for GET /v1/dashboard/stats.
type DashboardStatsResponse struct {
	Devices   DevicesStats   `json:"devices"`
	Commands  CommandsStats  `json:"commands"`
	Activity  ActivityStats  `json:"activity"`
}

// DevicesStats represents device statistics.
type DevicesStats struct {
	Total  int `json:"total"`
	Online int `json:"online"`
	Offline int `json:"offline"`
}

// CommandsStats represents command statistics.
type CommandsStats struct {
	TotalToday int `json:"totalToday"`
	Pending    int `json:"pending"`
	Failed     int `json:"failed"`
}

// ActivityStats represents activity in the last 24 hours.
type ActivityStats struct {
	Last24h ActivityDetail `json:"last24h"`
}

// ActivityDetail represents detailed activity.
type ActivityDetail struct {
	Commands        int `json:"commands"`
	Registrations  int `json:"registrations"`
	Deregistrations int `json:"deregistrations"`
}
