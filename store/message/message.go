package message

import (
	"context"

	"github.com/fox-one/mtg/core"
	"github.com/fox-one/pkg/store/db"
)

func New(db *db.DB) core.MessageStore {
	return &messageStore{db: db}
}

type messageStore struct {
	db *db.DB
}

func (s *messageStore) Create(ctx context.Context, messages []*core.Message) error {
	return s.db.Tx(func(tx *db.DB) error {
		for _, msg := range messages {
			if err := tx.Update().Create(msg).Error; err != nil {
				return err
			}
		}

		return nil
	})
}
