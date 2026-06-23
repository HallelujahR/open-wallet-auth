package auth

import (
	"context"
	"crypto/hmac"
	"crypto/md5"
	"crypto/sha1"
	"fmt"
	"strings"

	"github.com/open-wallet-auth/open-wallet-auth/internal/repository"
)

// verifyPassword accepts the current bcrypt hash or a configured legacy hash.
// verifyPassword 先校验当前 bcrypt 密码；失败后尝试旧系统密码凭证，并在成功后升级哈希。
func (s *Service) verifyPassword(ctx context.Context, userID string, currentHash string, plain string) (bool, error) {
	if currentHash != "" && s.hasher.Compare(currentHash, plain) {
		return true, nil
	}
	if s.legacy == nil {
		return false, nil
	}

	legacyItems, err := s.legacy.FindActiveByUserID(ctx, userID)
	if err != nil {
		return false, err
	}
	for _, item := range legacyItems {
		if !matchLegacyPassword(item, plain) {
			continue
		}

		// 旧密码只用于首次登录。成功后立即升级为当前哈希，减少长期兼容面。
		nextHash, err := s.hasher.Hash(plain)
		if err != nil {
			return false, err
		}
		if err := s.users.UpdatePassword(ctx, userID, nextHash); err != nil {
			return false, err
		}
		if err := s.legacy.MarkMigrated(ctx, item.ID); err != nil {
			return false, err
		}
		return true, nil
	}
	return false, nil
}

// matchLegacyPassword checks supported legacy password hash formats.
// matchLegacyPassword 校验通用旧密码哈希格式；不要在这里写具体业务系统名称。
func matchLegacyPassword(item repository.LegacyCredential, plain string) bool {
	hashType := strings.ToLower(strings.TrimSpace(item.HashType))
	expected := strings.TrimSpace(item.PasswordHash)
	switch hashType {
	case "hmac_sha1", "hmac-sha1":
		return equalHex(expected, hmacSHA1Hex(plain, item.Salt))
	case "sha1":
		return equalHex(expected, sha1Hex(plain+item.Salt))
	case "md5":
		return equalHex(expected, md5Hex(plain+item.Salt))
	default:
		return false
	}
}

// hmacSHA1Hex reproduces common legacy HMAC-SHA1 password storage.
// hmacSHA1Hex 复现旧系统常见的 HMAC-SHA1 密码存储方式。
func hmacSHA1Hex(plain string, key string) string {
	h := hmac.New(sha1.New, []byte(key))
	_, _ = h.Write([]byte(plain))
	return fmt.Sprintf("%x", h.Sum(nil))
}

// sha1Hex returns a lowercase SHA1 hex digest.
// sha1Hex 返回小写 SHA1 十六进制摘要。
func sha1Hex(value string) string {
	h := sha1.New()
	_, _ = h.Write([]byte(value))
	return fmt.Sprintf("%x", h.Sum(nil))
}

// md5Hex returns a lowercase MD5 hex digest for legacy-only verification.
// md5Hex 仅用于旧系统迁移校验，不用于新密码存储。
func md5Hex(value string) string {
	return fmt.Sprintf("%x", md5.Sum([]byte(value)))
}

// equalHex compares normalized hex digests.
// equalHex 比较归一化后的十六进制摘要。
func equalHex(left string, right string) bool {
	return strings.EqualFold(strings.TrimSpace(left), strings.TrimSpace(right))
}
