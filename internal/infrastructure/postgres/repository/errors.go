package repository

import (
	"errors"

	"gorm.io/gorm"

	domainrepo "github.com/open-wallet-auth/open-wallet-auth/internal/repository"
)

func mapGormError(err error) error {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return domainrepo.ErrNotFound
	}
	return err
}
