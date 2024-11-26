package storage

import (
	"context"
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3iface"
	"github.com/stretchr/testify/assert"
)

type MockS3Client struct {
	s3iface.S3API
	putObjectErr   error
	putObjectInput *s3.PutObjectInput
}

func (m *MockS3Client) PutObjectWithContext(ctx context.Context, input *s3.PutObjectInput, opts ...request.Option) (*s3.PutObjectOutput, error) {
	m.putObjectInput = input
	return &s3.PutObjectOutput{}, m.putObjectErr
}

func TestS3FileStorage_Upload(t *testing.T) {
	tests := []struct {
		name    string
		key     string
		data    []byte
		bucket  string
		mockErr error
		want    string
		wantErr bool
	}{
		{
			name:    "successful upload",
			key:     "test-key",
			data:    []byte("test data"),
			bucket:  "test-bucket",
			mockErr: nil,
			want:    "test-key",
			wantErr: false,
		},
		{
			name:    "upload fails",
			key:     "test-key",
			data:    []byte("test data"),
			bucket:  "test-bucket",
			mockErr: errors.New("s3 error"),
			want:    "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockS3 := &MockS3Client{
				putObjectErr: tt.mockErr,
			}

			storage := &S3FileStorage{
				client: mockS3,
				bucket: tt.bucket,
			}

			got, err := storage.Upload(context.Background(), tt.key, tt.data)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Equal(t, tt.want, got)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}

			assert.Equal(t, aws.StringValue(mockS3.putObjectInput.Bucket), tt.bucket)
			assert.Equal(t, aws.StringValue(mockS3.putObjectInput.Key), tt.key)
		})
	}
}

func TestNewS3FileStorage(t *testing.T) {
	tests := []struct {
		name           string
		region         string
		bucket         string
		endpoint       string
		disableSSL     bool
		forcePathStyle bool
	}{
		{
			name:           "create with standard config",
			region:         "us-west-2",
			bucket:         "test-bucket",
			endpoint:       "",
			disableSSL:     false,
			forcePathStyle: false,
		},
		{
			name:           "create with custom endpoint",
			region:         "us-west-2",
			bucket:         "test-bucket",
			endpoint:       "http://localhost:4566",
			disableSSL:     true,
			forcePathStyle: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage := NewS3FileStorage(
				tt.region,
				tt.bucket,
				tt.endpoint,
				tt.disableSSL,
				tt.forcePathStyle,
			)

			assert.NotNil(t, storage)
			assert.NotNil(t, storage.client)
			assert.Equal(t, tt.bucket, storage.bucket)
		})
	}
}
