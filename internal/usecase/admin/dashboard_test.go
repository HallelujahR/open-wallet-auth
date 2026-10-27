package admin

import (
	"testing"
	"time"
)

func TestDashboardWindowUsesBrowserDayAcrossUTCBoundary(t *testing.T) {
	now := time.Date(2026, 10, 26, 17, 15, 0, 0, time.UTC)
	start, end, days := dashboardWindow(now, 480)
	if days[6] != "2026-10-27" || days[0] != "2026-10-21" {
		t.Fatalf("wrong local days: %v", days)
	}
	if start.UTC().Format(time.RFC3339) != "2026-10-20T16:00:00Z" || end.Sub(start) != 7*24*time.Hour {
		t.Fatalf("wrong query window %v - %v", start, end)
	}
	_, _, western := dashboardWindow(now, -600)
	if western[6] != "2026-10-26" {
		t.Fatalf("wrong western day: %v", western)
	}
}
