package rag

import (
	"context"
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

func (c *Client) BuildKnowledgeBase(ctx context.Context, knowledgeBasePath string, model string) error {
	if knowledgeBasePath == "" {
		return nil
	}
	if err := c.store.Init(ctx); err != nil {
		return err
	}
	return filepath.WalkDir(knowledgeBasePath, func(path string, d fs.DirEntry, e error) error {
		if d != nil && d.IsDir() {
			return nil
		}
		logger := slog.With(slog.String("file", path))
		ext := filepath.Ext(path)
		if ext != ".txt" {
			return nil
		}
		logger.Info("build knowledge base")

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
