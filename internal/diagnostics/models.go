package diagnostics

import "time"

// ServerStatus holds overall server health and performance metrics
type ServerStatus struct {
	Version   string `json:"version"`
	Status    string `json:"status"`
	Uptime    string `json:"uptime"`
	CPU       string `json:"cpu"`
	Memory    string `json:"memory"`
	Storage   string `json:"storage"`
	Database  string `json:"database"`
	Users     int    `json:"users"`
	Realtime  string `json:"realtime"`
	Functions int    `json:"functions"`
}

// RealtimeMetrics holds WebSocket metrics
type RealtimeMetrics struct {
	ConnectedUsers int    `json:"connected_users"`
	ActiveRooms    int    `json:"active_rooms"`
	PresenceCount  int    `json:"presence_count"`
	MessagesPerSec string `json:"messages_per_sec"`
}

// StorageAnalytics holds storage metrics and file stats
type StorageAnalytics struct {
	PublicBucketSize  string     `json:"public_bucket_size"`
	PrivateBucketSize string     `json:"private_bucket_size"`
	TotalFiles        int        `json:"total_files"`
	LargestFiles      []FileInfo `json:"largest_files"`
	RecentUploads     []FileInfo `json:"recent_uploads"`
}

type FileInfo struct {
	Name      string    `json:"name"`
	Size      string    `json:"size"`
	Bucket    string    `json:"bucket"`
	UpdatedAt time.Time `json:"updated_at"`
}

// LogEntry represents a single server log record
type LogEntry struct {
	Level     string `json:"level"`    // INFO, WARN, ERROR, DB, WS, FUNC, STORAGE
	Message   string `json:"message"`
	Duration  string `json:"duration,omitempty"`
	Timestamp string `json:"timestamp"`
}
