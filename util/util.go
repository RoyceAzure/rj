package util

import (
	"os/exec"
	"path/filepath"
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

// Path 相關
// 取得專案根目錄
// 參數:
// moduleName: 模組名稱
func GetProjectRoot(moduleName string) string {
	// 執行 go list，但是加上額外的過濾條件
	cmd := exec.Command("go", "list", "-m", "-f", "{{.Dir}}", moduleName)
	output, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(output))
}

func WindwosPathToURL(winPath string) string {
	return filepath.ToSlash(winPath)
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

func StringPtrOrNil(s *string) *string {
	if s == nil {
		return nil
	}

	trimmed := strings.TrimSpace(*s)

	if trimmed == "" {
		return nil
	}
	return &trimmed
}
