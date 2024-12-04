package rag

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pgvector/pgvector-go"
	pgxvector "github.com/pgvector/pgvector-go/pgx"
	"os"
)

type VectorStore struct {
	pool *pgxpool.Pool
}

func NewPgVectorStore() (*VectorStore, error) {
	poolConfig, err := pgxpool.ParseConfig(os.Getenv("POSTGRES_URL"))
	if err != nil {
		return nil, err
	}
	poolConfig.ConnConfig.DefaultQueryExecMode = pgx.QueryExecModeCacheDescribe
	poolConfig.AfterConnect = func(ctx context.Context, conn *pgx.Conn) error {
		return pgxvector.RegisterTypes(ctx, conn)
	}

	pool, err := pgxpool.NewWithConfig(context.Background(), poolConfig)
	if err != nil {
		return nil, err
	}

	return &VectorStore{
		pool: pool,
	}, nil
}

func (s *VectorStore) Init(ctx context.Context) error {
	query := `
        CREATE TABLE IF NOT EXISTS knowledge_base (
			id SERIAL PRIMARY KEY,
			document_id TEXT NOT NULL,
			part_index INT NOT NULL,
			content TEXT,
			embedding vector(1536),
			UNIQUE (document_id, part_index)
		)
  	`
	if _, err := s.pool.Exec(ctx, query); err != nil {
		return err
	}

	return nil
}

func (s *VectorStore) GetMostRelevantDocuments(ctx context.Context, embedding []float32, threshold float32, limit int) ([]*DocumentPart, error) {
	query := fmt.Sprintf("SELECT document_id, part_index, content FROM knowledge_base WHERE embedding <-> $1 < $2 ORDER BY embedding <-> $1 LIMIT %v", limit)
	rows, err := s.pool.Query(ctx, query, pgvector.NewVector(embedding), threshold)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var documents []*DocumentPart
	for rows.Next() {
		document := &DocumentPart{}
		if err := rows.Scan(&document.DocumentID, &document.PartIndex, &document.Content); err != nil {
			return nil, err
		}
		documents = append(documents, document)
	}

	if rows.Err() != nil {
		return nil, rows.Err()
	}

	return documents, nil
}

func (s *VectorStore) SaveDocumentEmbedding(ctx context.Context, documentID string, partIndex int, content string, embedding []float32) error {
	_, err := s.pool.Exec(ctx, `
		INSERT INTO knowledge_base (document_id, part_index, content, embedding)
			VALUES ($1, $2, $3, $4)
		ON CONFLICT (document_id, part_index)
		DO UPDATE SET
			content = EXCLUDED.content,
			embedding = EXCLUDED.embedding;
	`, documentID, partIndex, content, pgvector.NewVector(embedding))

	if err != nil {
		return err
	}

	return nil
}

func (s *VectorStore) CheckDocumentExist(ctx context.Context, documentID string) (bool, error) {
	var exists bool
	err := s.pool.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1 FROM knowledge_base WHERE document_id = $1
		)
	`, documentID).Scan(&exists)

	if err != nil {
		return false, nil
	}

	return exists, nil
}

func (s *VectorStore) DeleteDocuments(ctx context.Context, documentID string) error {
	_, err := s.pool.Exec(ctx, `
		DELETE FROM knowledge_base WHERE document_id = $1
	`, documentID)

	if err != nil {
		return err
	}

	return nil
}

func (s *VectorStore) Release() {
	s.pool.Close()
}
