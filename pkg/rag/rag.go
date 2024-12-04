package rag

import (
	"context"
	"fmt"
	"github.com/tonnie17/wxagent/pkg/embedding"
	"io/fs"
	"log/slog"
	"path/filepath"
)

type Client struct {
	embeddingModel embedding.Model
	store          *VectorStore
}

func NewClient(embeddingModel embedding.Model, store *VectorStore) *Client {
	return &Client{
		embeddingModel: embeddingModel,
		store:          store,
	}
}

func (c *Client) Query(ctx context.Context, model string, content string, limit int) ([]*DocumentPart, error) {
	embeddingData, err := c.embeddingModel.CreateEmbeddings(ctx, model, content)
	if err != nil {
		return nil, err
	}

	return c.store.GetMostRelevantDocuments(ctx, embeddingData, 1, limit)
}

func (c *Client) LoadData(ctx context.Context, dir string, model string) error {
	if dir == "" {
		return nil
	}
	return filepath.WalkDir(dir, func(path string, d fs.DirEntry, e error) error {
		if d != nil && d.IsDir() {
			return nil
		}
		logger := slog.With(slog.String("path", path))
		ext := filepath.Ext(path)
		switch ext {
		case ".txt":
			documentID := filepath.Base(path)
			documentExist, err := c.store.CheckDocumentExist(ctx, documentID)
			if err != nil {
				logger.Error("check document exist failed", slog.Any("err", err))
				return nil
			}
			if documentExist {
				return nil
			}

			out, err := processTextFile(path)
			if err != nil {
				logger.Error("process text file failed", slog.Any("err", err))
				return nil
			}

			var partIndex int
			for content := range out {
				embeddingData, err := c.embeddingModel.CreateEmbeddings(ctx, model, content)
				if err != nil {
					logger.Error("create embeddings failed", slog.Any("err", err))
					return nil
				}
				partIndex++
				if err := c.store.SaveDocumentEmbedding(ctx, documentID, partIndex, content, embeddingData); err != nil {
					logger.Error("save embedding failed", slog.Any("err", err))
					return nil
				}
			}
		}

		return nil
	})
}

func Prompt(context string, question string) string {
	return fmt.Sprintf(`You are an assistant. Answer the question based on the given context.

Context:
%v

Question:
%v

Answer:`, context, question)
}
