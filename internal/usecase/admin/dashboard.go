package admin

import (
	"context"
	"time"

	"github.com/open-wallet-auth/open-wallet-auth/internal/domain"
	"github.com/open-wallet-auth/open-wallet-auth/internal/repository"
)

func dashboardWindow(now time.Time, offsetMinutes int) (time.Time, time.Time, []string) {
	local := now.In(time.FixedZone("browser", offsetMinutes*60))
	end := time.Date(local.Year(), local.Month(), local.Day()+1, 0, 0, 0, 0, local.Location())
	start := end.AddDate(0, 0, -7)
	days := make([]string, 7)
	for i := range days {
		days[i] = start.AddDate(0, 0, i).Format("2006-01-02")
	}
	return start, end, days
}

func (s *Service) Dashboard(ctx context.Context, offsetMinutes int) (*repository.DashboardSummary, error) {
	if offsetMinutes < -840 || offsetMinutes > 840 {
		return nil, domain.NewError(ErrInvalidInput, "invalid timezone offset")
	}
	if s.dashboard == nil {
		return nil, domain.NewError(ErrInvalidInput, "dashboard unavailable")
	}
	from, until, days := dashboardWindow(time.Now(), offsetMinutes)
	result, err := s.dashboard.Summary(ctx, from, until, offsetMinutes)
	if err != nil {
		return nil, err
	}
	result.Days = days
	return result, nil
}
