package auth

import (
	"crypto/md5"
	"fmt"
)

// ComputeTokenHash hashes a token for SSO session comparison.
//
// Both the auth middleware (request-time check) and the RBAC service (login-time
// storage) MUST use this exact function — the stored session hash and the
// request-time hash are compared directly at runtime. Keeping a single shared
// implementation prevents the divergence that previously broke every
// authenticated request.
func ComputeTokenHash(token, userID string) string {
	value := md5.Sum([]byte(token + userID))
	return fmt.Sprintf("%x", value)
}
