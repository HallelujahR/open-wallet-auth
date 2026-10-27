package repository

import (
	"context"
	"time"

	"github.com/open-wallet-auth/open-wallet-auth/internal/infrastructure/postgres/model"
	domainrepo "github.com/open-wallet-auth/open-wallet-auth/internal/repository"
)

func (r *ActivityRepository) Summary(ctx context.Context, from, until time.Time, offsetMinutes int) (*domainrepo.DashboardSummary, error) {
	result := &domainrepo.DashboardSummary{LoginBuckets: []domainrepo.LoginBucket{}}
	db := r.db.WithContext(ctx)
	if err := db.Model(&model.User{}).Where("status = ?", "active").Count(&result.ActiveUsers).Error; err != nil {
		return nil, err
	}
	if err := db.Model(&model.Client{}).Count(&result.Clients).Error; err != nil {
		return nil, err
	}
	if err := db.Model(&model.RefreshToken{}).Where("revoked_at IS NULL AND expires_at > ?", time.Now().UTC()).Count(&result.ActiveSessions).Error; err != nil {
		return nil, err
	}
	err := db.Model(&model.LoginLog{}).
		Select("to_char(created_at AT TIME ZONE 'UTC' + (? * interval '1 minute'), 'YYYY-MM-DD') AS day, login_method AS method, success, COUNT(*) AS count", offsetMinutes).
		Where("created_at >= ? AND created_at < ?", from, until).
		Group("day, login_method, success").Order("day").Scan(&result.LoginBuckets).Error
	if err != nil {
		return nil, err
	}
	return result, nil
}
