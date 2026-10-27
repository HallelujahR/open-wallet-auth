package repository

import (
	"context"
	"time"
)

type LoginBucket struct {
	Day     string `json:"day"`
	Method  string `json:"method"`
	Success bool   `json:"success"`
	Count   int64  `json:"count"`
}

type DashboardSummary struct {
	ActiveUsers    int64         `json:"active_users"`
	Clients        int64         `json:"clients"`
	ActiveSessions int64         `json:"active_sessions"`
	LoginBuckets   []LoginBucket `json:"login_buckets"`
	Days           []string      `json:"days"`
}

// DashboardRepository aggregates complete datasets instead of counting one page.
type DashboardRepository interface {
	Summary(context.Context, time.Time, time.Time, int) (*DashboardSummary, error)
}
