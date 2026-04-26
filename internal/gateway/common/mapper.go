package common

import userv1 "github.com/squ1ky/flyte/gen/proto/user"

func RoleFromProto(role userv1.Role) string {
	switch role {
	case userv1.Role_ROLE_ADMIN:
		return RoleAdmin
	case userv1.Role_ROLE_USER:
		return RoleUser
	default:
		return ""
	}
}
