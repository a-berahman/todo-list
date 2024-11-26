package queue

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/aws/aws-sdk-go/service/sqs/sqsiface"
	"github.com/stretchr/testify/assert"
)

type MockSQSClient struct {
	sqsiface.SQSAPI
	sendMessageFunc func(*sqs.SendMessageInput) (*sqs.SendMessageOutput, error)
}

func (m *MockSQSClient) SendMessageWithContext(ctx context.Context, input *sqs.SendMessageInput, opts ...request.Option) (*sqs.SendMessageOutput, error) {
	return m.sendMessageFunc(input)
}

type publishTestCase struct {
	name          string
	message       string
	queueURL      string
	mockResponse  *sqs.SendMessageOutput
	mockError     error
	expectedError error
}

func TestSQSPublisher_Publish(t *testing.T) {
	testCases := []publishTestCase{
		{
			name:     "successful message publish",
			message:  "test message",
			queueURL: "https://sqs.test.amazonaws.com/123456789012/test-queue",
			mockResponse: &sqs.SendMessageOutput{
				MessageId: aws.String("test-message-id"),
			},
			mockError:     nil,
			expectedError: nil,
		},
		{
			name:          "empty message",
			message:       "",
			queueURL:      "https://sqs.test.amazonaws.com/123456789012/test-queue",
			mockResponse:  nil,
			mockError:     aws.ErrMissingEndpoint,
			expectedError: aws.ErrMissingEndpoint,
		},
		{
			name:          "invalid queue URL",
			message:       "test message",
			queueURL:      "",
			mockResponse:  nil,
			mockError:     aws.ErrMissingEndpoint,
			expectedError: aws.ErrMissingEndpoint,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {

			mockSQS := &MockSQSClient{
				sendMessageFunc: func(input *sqs.SendMessageInput) (*sqs.SendMessageOutput, error) {
					// Verify input parameters
					assert.Equal(t, tc.queueURL, *input.QueueUrl)
					assert.Equal(t, tc.message, *input.MessageBody)
					return tc.mockResponse, tc.mockError
				},
			}

			publisher := &SQSPublisher{
				client:   mockSQS,
				queueURL: tc.queueURL,
			}

			err := publisher.Publish(context.Background(), tc.message)

			if tc.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tc.expectedError, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestNewSQSPublisher(t *testing.T) {
	testCases := []struct {
		name       string
		region     string
		queueURL   string
		endpoint   string
		disableSSL bool
	}{
		{
			name:       "create publisher with default config",
			region:     "us-east-1",
			queueURL:   "https://sqs.test.amazonaws.com/123456789012/test-queue",
			endpoint:   "http://localhost:4566",
			disableSSL: true,
		},
		{
			name:       "create publisher with SSL enabled",
			region:     "eu-west-1",
			queueURL:   "https://sqs.test.amazonaws.com/123456789012/test-queue",
			endpoint:   "https://sqs.eu-west-1.amazonaws.com",
			disableSSL: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			publisher := NewSQSPublisher(tc.region, tc.queueURL, tc.endpoint, tc.disableSSL)

			assert.NotNil(t, publisher)
			assert.NotNil(t, publisher.client)
			assert.Equal(t, tc.queueURL, publisher.queueURL)
		})
	}
}
