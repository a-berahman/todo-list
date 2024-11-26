// internal/infra/storage/s3_storage.go
package storage

import (
	"bytes"
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3iface"
)

type S3FileStorage struct {
	client s3iface.S3API
	bucket string
}

func NewS3FileStorage(region, bucket, endpoint string, disableSSL, forcePathStyle bool) *S3FileStorage {
	sess := session.Must(session.NewSession(
		&aws.Config{
			Region:           aws.String(region),
			Endpoint:         aws.String(endpoint),
			DisableSSL:       aws.Bool(disableSSL),
			S3ForcePathStyle: aws.Bool(forcePathStyle),
		}))
	return &S3FileStorage{
		client: s3.New(sess),
		bucket: bucket,
	}
}

func (s *S3FileStorage) Upload(ctx context.Context, key string, data []byte) (string, error) {
	_, err := s.client.PutObjectWithContext(ctx, &s3.PutObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
		Body:   bytes.NewReader(data),
	})
	if err != nil {
		return "", err
	}
	return key, nil
}
