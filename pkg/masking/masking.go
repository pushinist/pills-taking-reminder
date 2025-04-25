package masking

import (
	"fmt"
	"strings"
)

func MaskData(key string, value any) string {
	switch key {
	case "user_id":
		return maskUserID(fmt.Sprint(value))
	default:
		return "****"
	}
}

func maskUserID(id string) string {
	if len(id) <= 4 {
		return "****"
	}
	return id[:2] + strings.Repeat("*", len(id)-4) + id[len(id)-2:]
}
