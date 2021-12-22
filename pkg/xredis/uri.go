package xredis

import (
	"fmt"
	"strings"

	"github.com/google/uuid"
)

func NewUri(kinds ...string) string {
	return fmt.Sprintf("gid://%s/%s", strings.Join(kinds, "/"), uuid.New())
}

func IsValidUri(i string) bool {
	return strings.HasPrefix(i, "gid://")
}
