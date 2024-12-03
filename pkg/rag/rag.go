package rag

import (
	"context"
	"fmt"
	"github.com/tonnie17/wxagent/pkg/embedding"
	"io/fs"
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
	return filepath.WalkDir(dir, func(path string, d fs.DirEntry, e error) error {
		if d != nil && d.IsDir() {
			return nil
		}
		ext := filepath.Ext(path)
		switch ext {
		case ".txt":
			documentID := filepath.Base(path)
			documentExist, err := c.store.CheckDocumentExist(ctx, documentID)
			if err != nil {
				return err
			}
			if documentExist {
				return nil
			}

			out, err := processTextFile(path)
			if err != nil {
				return err
			}

			var partIndex int
			for content := range out {
				embeddingData, err := c.embeddingModel.CreateEmbeddings(ctx, model, content)
				if err != nil {
					return err
				}
				partIndex++
				if err := c.store.SaveDocumentEmbedding(ctx, documentID, partIndex, content, embeddingData); err != nil {
					return err
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
