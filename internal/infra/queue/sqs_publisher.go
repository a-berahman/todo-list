// internal/infra/queue/sqs_publisher.go
package queue

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/aws/aws-sdk-go/service/sqs/sqsiface"
)

type SQSPublisher struct {
	client   sqsiface.SQSAPI
	queueURL string
}

func NewSQSPublisher(region, queueURL, endpoint string, disableSSL bool) *SQSPublisher {
	sess := session.Must(session.NewSession(
		&aws.Config{
			Region:     aws.String(region),
			Endpoint:   aws.String(endpoint),
			DisableSSL: aws.Bool(disableSSL),
		}))
	return &SQSPublisher{
		client:   sqs.New(sess),
		queueURL: queueURL,
	}
}

func (p *SQSPublisher) Publish(ctx context.Context, message string) error {
	_, err := p.client.SendMessageWithContext(ctx, &sqs.SendMessageInput{
		QueueUrl:    aws.String(p.queueURL),
		MessageBody: aws.String(message),
	})
	return err
}
