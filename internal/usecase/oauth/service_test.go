package oauth

import (
	"context"
	"testing"

	"github.com/open-wallet-auth/open-wallet-auth/internal/domain/user"
	"github.com/open-wallet-auth/open-wallet-auth/internal/repository"
)

// TestResolveOAuthTargetUserAutoLinksVerifiedEmail keeps Google/GitHub verified emails low-friction.
// TestResolveOAuthTargetUserAutoLinksVerifiedEmail 验证已验证 OAuth 邮箱命中已有账号时会自动归属到该账号。
func TestResolveOAuthTargetUserAutoLinksVerifiedEmail(t *testing.T) {
	existing := &user.User{ID: "usr_existing", Email: "river@example.com", Status: user.StatusActive}
	service := &Service{users: &memoryOAuthUsers{byEmail: map[string]*user.User{existing.Email: existing}}}

	got, err := service.resolveOAuthTargetUser(context.Background(), "google", &ProviderUser{
		Subject:       "google-subject",
		Email:         existing.Email,
		EmailVerified: true,
		Username:      "River",
	}, "")
	if err != nil {
		t.Fatalf("resolve target user: %v", err)
	}
	if got.ID != existing.ID {
		t.Fatalf("expected existing user %q, got %q", existing.ID, got.ID)
	}
}

// TestResolveOAuthTargetUserRejectsUnverifiedEmailConflict preserves explicit binding safety.
// TestResolveOAuthTargetUserRejectsUnverifiedEmailConflict 验证未验证邮箱仍然不会自动合并账号。
func TestResolveOAuthTargetUserRejectsUnverifiedEmailConflict(t *testing.T) {
	existing := &user.User{ID: "usr_existing", Email: "river@example.com", Status: user.StatusActive}
	service := &Service{users: &memoryOAuthUsers{byEmail: map[string]*user.User{existing.Email: existing}}}

	_, err := service.resolveOAuthTargetUser(context.Background(), "github", &ProviderUser{
		Subject:       "github-subject",
		Email:         existing.Email,
		EmailVerified: false,
		Username:      "river",
	}, "")
	if err == nil {
		t.Fatal("expected unverified email conflict to be rejected")
	}
}

type memoryOAuthUsers struct {
	byEmail map[string]*user.User
}

func (m *memoryOAuthUsers) FindByID(ctx context.Context, id string) (*user.User, error) {
	for _, item := range m.byEmail {
		if item.ID == id {
			return item, nil
		}
	}
	return nil, repository.ErrNotFound
}

func (m *memoryOAuthUsers) FindByEmail(ctx context.Context, email string) (*user.User, error) {
	if item, ok := m.byEmail[email]; ok {
		return item, nil
	}
	return nil, repository.ErrNotFound
}

func (m *memoryOAuthUsers) FindByPhone(ctx context.Context, phone string) (*user.User, error) {
	return nil, repository.ErrNotFound
}

func (m *memoryOAuthUsers) Create(ctx context.Context, u *user.User) error {
	if m.byEmail == nil {
		m.byEmail = map[string]*user.User{}
	}
	m.byEmail[u.Email] = u
	return nil
}

func (m *memoryOAuthUsers) UpdateLoginInfo(ctx context.Context, userID string) error {
	return nil
}

func (m *memoryOAuthUsers) UpdateEmail(ctx context.Context, userID string, email string) error {
	return nil
}

func (m *memoryOAuthUsers) UpdatePhone(ctx context.Context, userID string, phone string) error {
	return nil
}

func (m *memoryOAuthUsers) UpdatePassword(ctx context.Context, userID string, passwordHash string) error {
	return nil
}

func (m *memoryOAuthUsers) UpdateProfile(ctx context.Context, userID string, username string, avatar string) error {
	return nil
}
