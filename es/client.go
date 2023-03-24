package es

import (
	"context"
	"encoding/json"
)

type Client interface {
	Dump(ctx context.Context, receive chan<- json.RawMessage) error
}
