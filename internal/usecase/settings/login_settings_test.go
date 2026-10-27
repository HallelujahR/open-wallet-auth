package settings

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/open-wallet-auth/open-wallet-auth/internal/repository"
)

type memorySettings struct{ value json.RawMessage }

func (m *memorySettings) Get(context.Context, string) (json.RawMessage, error) {
	if m.value == nil {
		return nil, repository.ErrNotFound
	}
	return m.value, nil
}
func (m *memorySettings) Upsert(_ context.Context, _ string, value json.RawMessage) error {
	m.value = append(json.RawMessage(nil), value...)
	return nil
}

func TestLoginTitlePersistsAndUpdatesWithoutRestart(t *testing.T) {
	ctx := context.Background()
	store := &memorySettings{value: json.RawMessage(`{"login":{"brand_name":"Existing Brand","enable_wallet":true}}`)}
	service := NewService(store, Snapshot{}, ReadonlySettings{})
	current, err := service.Get(ctx)
	if err != nil || current.Login.HeroTitle != "一个账号，\n连接你的应用。" || current.Login.BrandName != "Existing Brand" || !current.Login.EnableWallet {
		t.Fatalf("legacy settings lost: %+v, %v", current.Login, err)
	}
	for _, title := range []string{"连接团队，\n开启协作。", "新的登录标题", ""} {
		current.Login.HeroTitle = title
		if _, err := service.Update(ctx, current); err != nil {
			t.Fatal(err)
		}
		want := title
		if want == "" {
			want = "一个账号，\n连接你的应用。"
		}
		for _, reader := range []*Service{service, NewService(store, Snapshot{}, ReadonlySettings{})} {
			got, err := reader.LoginSettings(ctx)
			if err != nil || got.HeroTitle != want || got.BrandName != "Existing Brand" || !got.EnableWallet {
				t.Fatalf("updated settings not retained: %+v, %v", got, err)
			}
		}
	}
}
