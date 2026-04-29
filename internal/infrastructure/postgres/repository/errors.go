package repository

import (
	"errors"

	"gorm.io/gorm"

	domainrepo "github.com/open-wallet-auth/open-wallet-auth/internal/repository"
)

// mapGormError translates infrastructure errors into repository-level errors.
// mapGormError 将 GORM 基础设施错误转换为仓储层稳定错误。
func mapGormError(err error) error {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return domainrepo.ErrNotFound
	}
	return err
}
