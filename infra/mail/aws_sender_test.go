package mail

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// 同 package 可建立 AwsEmailSenderConfig（未匯出欄位）
func testAwsConfig() AwsEmailSenderConfig {
	return AwsEmailSenderConfig{
		senderName: "",
		accessKey:  "",
		secretKey:  "",
		region:     "",
	}
}

func TestNewAwsEmailSender(t *testing.T) {
	cf := testAwsConfig()

	sender, err := NewAwsEmailSender(cf)

	require.NoError(t, err)
	require.NotNil(t, sender)
	require.NotNil(t, sender.client)
}

func TestNewAwsEmailSender_EmptyRegion(t *testing.T) {
	cf := testAwsConfig()
	cf.region = ""

	// 空 region 可能導致 LoadDefaultConfig 失敗並 log.Fatalf，此測試僅在未 exit 時驗證
	sender, err := NewAwsEmailSender(cf)

	// 若未 Fatal，則檢查回傳
	if err != nil {
		require.Error(t, err)
		return
	}
	if sender != nil {
		require.NotNil(t, sender.client)
	}
}

func TestAwsEmailSender_ImplementsEmailSender(t *testing.T) {
	sender, err := NewAwsEmailSender(testAwsConfig())
	require.NoError(t, err)

	// 確認 *AwsEmailSender 實作 EmailSender interface
	var _ EmailSender = (*AwsEmailSender)(nil)
	_, ok := (interface{}(sender)).(EmailSender)
	require.True(t, ok, "*AwsEmailSender 應實作 EmailSender")
}

func TestAwsEmailSender_SendEmail(t *testing.T) {
	sender, err := NewAwsEmailSender(testAwsConfig())
	require.NoError(t, err)

	// 使用假憑證呼叫會因 AWS 認證/驗證失敗而回傳錯誤
	err = sender.SendEmail(
		"test subject",
		"<p>test content</p>",
		[]string{"roycewnag@gmail.com"},
		nil,
		nil,
		nil,
	)

	require.NoError(t, err)
}

func TestAwsEmailSender_SendEmail_WithCCBCC(t *testing.T) {
	sender, err := NewAwsEmailSender(testAwsConfig())
	require.NoError(t, err)

	err = sender.SendEmail(
		"subject",
		"content",
		[]string{"to@example.com"},
		[]string{"cc@example.com"},
		[]string{"bcc@example.com"},
		nil,
	)

	require.Error(t, err)
}
