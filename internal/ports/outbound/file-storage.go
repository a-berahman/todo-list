package outbound

import "context"

type FileStorage interface {
	Upload(ctx context.Context, key string, file []byte) (string, error)
}
