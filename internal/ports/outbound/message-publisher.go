package outbound

import "context"

type MessagePublisher interface {
	Publish(ctx context.Context, message string) error
}
