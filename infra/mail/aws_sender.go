package mail

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/ses"
	"github.com/aws/aws-sdk-go-v2/service/ses/types"
)

type AwsEmailSenderConfig struct {
	senderName string
	accessKey  string
	secretKey  string
	region     string
}

type AwsEmailSender struct {
	config AwsEmailSenderConfig
	client *ses.Client
}

func NewAwsEmailSender(cf AwsEmailSenderConfig) (*AwsEmailSender, error) {
	creds := credentials.NewStaticCredentialsProvider(cf.accessKey, cf.secretKey, "")
	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithCredentialsProvider(creds),
		config.WithRegion(cf.region),
	)
	if err != nil {
		log.Fatalf("無法載入 AWS 設定: %v", err)
	}
	client := ses.NewFromConfig(cfg)
	return &AwsEmailSender{
		config: cf,
		client: client,
	}, nil
}

/*
附件格式  中文編碼??
*/
func (a *AwsEmailSender) SendEmail(
	subject string,
	content string,
	to []string,
	cc []string,
	bcc []string,
	attachFiles []string,
) error {
	charSet := "UTF-8"
	input := &ses.SendEmailInput{
		Source: aws.String(a.config.senderName),
		Destination: &types.Destination{
			ToAddresses: to,
		},
		Message: &types.Message{
			Subject: &types.Content{
				Data:    aws.String(subject),
				Charset: aws.String(charSet),
			},
			Body: &types.Body{
				// 設定 HTML 內容
				Html: &types.Content{
					Data:    aws.String(content),
					Charset: aws.String(charSet),
				},
			},
		},
	}

	_, err := a.client.SendEmail(context.TODO(), input)
	return err
}
