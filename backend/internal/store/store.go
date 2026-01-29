package store

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrNotFound = errors.New("not found")
var ErrForbidden = errors.New("forbidden")

type Store struct {
	pool *pgxpool.Pool
}

func New(pool *pgxpool.Pool) *Store {
	return &Store{pool: pool}
}

func (s *Store) CreateCollection(ctx context.Context, ownerID, name string) (Collection, error) {
	row := s.pool.QueryRow(ctx, `
		INSERT INTO collections (owner_user_id, name)
		VALUES ($1, $2)
		RETURNING id, owner_user_id, name, created_at
	`, ownerID, name)

	var c Collection
	if err := row.Scan(&c.ID, &c.OwnerUserID, &c.Name, &c.CreatedAt); err != nil {
		return Collection{}, err
	}
	return c, nil
}

func (s *Store) ListCollections(ctx context.Context, ownerID string) ([]Collection, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, owner_user_id, name, created_at
		FROM collections
		WHERE owner_user_id = $1
		ORDER BY created_at DESC
	`, ownerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	collections := []Collection{}
	for rows.Next() {
		var c Collection
		if err := rows.Scan(&c.ID, &c.OwnerUserID, &c.Name, &c.CreatedAt); err != nil {
			return nil, err
		}
		collections = append(collections, c)
	}
	return collections, rows.Err()
}

func (s *Store) CreateDocumentAndJob(ctx context.Context, ownerID, collectionID, title string) (Document, IngestionJob, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return Document{}, IngestionJob{}, err
	}
	defer func() {
		_, _ = tx.Exec(ctx, "ROLLBACK")
	}()

	var exists bool
	if err := tx.QueryRow(ctx, `
		SELECT EXISTS(
			SELECT 1 FROM collections WHERE id = $1 AND owner_user_id = $2
		)
	`, collectionID, ownerID).Scan(&exists); err != nil {
		return Document{}, IngestionJob{}, err
	}
	if !exists {
		return Document{}, IngestionJob{}, ErrForbidden
	}

	var doc Document
	if err := tx.QueryRow(ctx, `
		INSERT INTO documents (collection_id, owner_user_id, title, status)
		VALUES ($1, $2, $3, 'uploaded')
		RETURNING id, collection_id, owner_user_id, title, status, created_at
	`, collectionID, ownerID, title).Scan(
		&doc.ID, &doc.CollectionID, &doc.OwnerUserID, &doc.Title, &doc.Status, &doc.CreatedAt,
	); err != nil {
		return Document{}, IngestionJob{}, err
	}

	var job IngestionJob
	if err := tx.QueryRow(ctx, `
		INSERT INTO ingestion_jobs (document_id, status, progress)
		VALUES ($1, 'pending', 0)
		RETURNING id, document_id, status, progress, error, created_at, updated_at
	`, doc.ID).Scan(
		&job.ID, &job.DocumentID, &job.Status, &job.Progress, &job.Error, &job.CreatedAt, &job.UpdatedAt,
	); err != nil {
		return Document{}, IngestionJob{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return Document{}, IngestionJob{}, err
	}

	return doc, job, nil
}

func (s *Store) GetDocument(ctx context.Context, ownerID, documentID string) (Document, error) {
	row := s.pool.QueryRow(ctx, `
		SELECT id, collection_id, owner_user_id, title, status, created_at
		FROM documents
		WHERE id = $1 AND owner_user_id = $2
	`, documentID, ownerID)

	var doc Document
	if err := row.Scan(&doc.ID, &doc.CollectionID, &doc.OwnerUserID, &doc.Title, &doc.Status, &doc.CreatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Document{}, ErrNotFound
		}
		return Document{}, err
	}
	return doc, nil
}

func (s *Store) GetIngestionJob(ctx context.Context, ownerID, jobID string) (IngestionJob, error) {
	row := s.pool.QueryRow(ctx, `
		SELECT j.id, j.document_id, j.status, j.progress, j.error, j.created_at, j.updated_at
		FROM ingestion_jobs j
		JOIN documents d ON d.id = j.document_id
		WHERE j.id = $1 AND d.owner_user_id = $2
	`, jobID, ownerID)

	var job IngestionJob
	if err := row.Scan(&job.ID, &job.DocumentID, &job.Status, &job.Progress, &job.Error, &job.CreatedAt, &job.UpdatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return IngestionJob{}, ErrNotFound
		}
		return IngestionJob{}, err
	}
	return job, nil
}

func (s *Store) Ping(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	return s.pool.Ping(ctx)
}
