package util

import (
	"os/exec"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

func StringToNumeric(value string) pgtype.Numeric {
	var result pgtype.Numeric
	value = strings.Replace(value, ",", "", -1)
	_ = result.Scan(value)
	return result
}

func GetProjectRoot() string {
	// 從 go.mod 讀取模塊路徑
	cmd := exec.Command("go", "list", "-m", "-f", "{{.Dir}}")
	output, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(output))
}

// UUIDToBytes converts a UUID to a byte slice
func UUIDToBytes(id uuid.UUID) []byte {
	bytes := make([]byte, 16)
	copy(bytes, id[:])
	return bytes
}

func RFC3339ToTimestamp(timeStr string) (int64, error) {
	t, err := time.Parse(time.RFC3339, timeStr)
	if err != nil {
		return 0, err
	}
	return t.Unix(), nil
}
