package mail

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewGmailSender(t *testing.T) {
	name := "Test Sender"
	from := "test@example.com"
	password := "secret"

	sender := NewGmailSender(name, from, password)
	require.NotNil(t, sender)

	// 確認可轉型為 GmailSender 且欄位正確
	gs, ok := sender.(*GmailSender)
	require.True(t, ok)
	require.Equal(t, name, gs.name)
	require.Equal(t, from, gs.fromEmailAddress)
	require.Equal(t, password, gs.fromEmailPassword)
}

func TestGmailSender_SendEmail_AttachmentFileNotFound(t *testing.T) {
	sender := NewGmailSender("Test", "test@example.com", "password").(*GmailSender)

	// 使用不存在的檔案路徑，應在附加階段就回傳錯誤（不會真的連到 SMTP）
	err := sender.SendEmail(
		"subject",
		"content",
		[]string{"to@example.com"},
		nil,
		nil,
		[]string{"/non/existent/path/file.txt"},
	)

	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to attach file")
}

func TestGmailSender_SendEmail_EmptyAttachmentList(t *testing.T) {
	// 無附件時會嘗試連線到 Gmail SMTP，沒有正確憑證會失敗
	// 此測試僅驗證在無附件情況下不會在 attach 階段出錯（實際 Send 會因連線/認證失敗）
	sender := NewGmailSender("Test", "test@example.com", "wrong-password").(*GmailSender)

	err := sender.SendEmail(
		"subject",
		"<p>html content</p>",
		[]string{"to@example.com"},
		nil,
		nil,
		nil, // 無附件
	)

	// 預期會因 SMTP 連線或認證失敗而回傳錯誤，而不是 attach 錯誤
	require.Error(t, err)
	require.NotContains(t, err.Error(), "failed to attach file")
}

func TestGmailSender_SendEmail_ValidAttachment(t *testing.T) {
	// 建立暫存檔作為附件
	dir := t.TempDir()
	attachPath := filepath.Join(dir, "attach.txt")
	err := os.WriteFile(attachPath, []byte("test content"), 0644)
	require.NoError(t, err)

	sender := NewGmailSender("Test", "test@example.com", "wrong-password").(*GmailSender)

	// 有有效附件，但 SMTP 會失敗，所以預期得到的是 Send 的錯誤而非 attach 錯誤
	err = sender.SendEmail(
		"subject",
		"content",
		[]string{"to@example.com"},
		nil,
		nil,
		[]string{attachPath},
	)

	require.Error(t, err)
	require.NotContains(t, err.Error(), "failed to attach file")
}
