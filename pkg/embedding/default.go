package embedding

import (
	"context"
	"fmt"
)

type NotImplemented struct {
}

func NewNotImplemented() Model {
	return &NotImplemented{}
}

func (o *NotImplemented) CreateEmbeddings(ctx context.Context, model string, content string) ([]float32, error) {
	return nil, fmt.Errorf("embedding not implemented")
}
