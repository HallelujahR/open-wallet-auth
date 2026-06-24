package repository

import (
	"context"

	"github.com/open-wallet-auth/open-wallet-auth/internal/domain/client"
)

// ClientRepository defines persistence operations for application clients.
type ClientRepository interface {
	FindByClientID(ctx context.Context, clientID string) (*client.Client, error)
	Create(ctx context.Context, client *client.Client) error
	List(ctx context.Context) ([]client.Client, error)
}

// ClientConfigRepository edits application-client configuration.
// ClientConfigRepository 编辑接入应用配置，通常只供管理后台使用。
type ClientConfigRepository interface {
	Update(ctx context.Context, client *client.Client) (*client.Client, error)
}

// ClientAccessReader checks whether a user can access one application client.
// ClientAccessReader 校验用户是否拥有某个业务系统的准入授权。
type ClientAccessReader interface {
	FindActiveMember(ctx context.Context, clientID string, userID string) (*client.Member, error)
}

// ClientMemberRepository manages application allow-list members.
// ClientMemberRepository 管理应用成员白名单。
type ClientMemberRepository interface {
	ClientAccessReader
	ListMembers(ctx context.Context, clientID string) ([]client.Member, error)
	UpsertMember(ctx context.Context, member *client.Member) error
	UpdateMember(ctx context.Context, member *client.Member) error
	DeleteMember(ctx context.Context, clientID string, memberID string) error
}
