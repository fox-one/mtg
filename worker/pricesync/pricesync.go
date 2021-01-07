package pricesync

import (
	"context"
	"time"

	"github.com/fox-one/mtg/core"
	"github.com/fox-one/pkg/logger"
)

func New(
	assets core.AssetStore,
	assetz core.AssetService,
) *Sync {
	return &Sync{
		assets: assets,
		assetz: assetz,
	}
}

type Sync struct {
	assets core.AssetStore
	assetz core.AssetService
}

func (w *Sync) Run(ctx context.Context) error {
	log := logger.FromContext(ctx).WithField("worker", "price syncer")
	ctx = logger.WithContext(ctx, log)

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(time.Second):
			_ = w.run(ctx)
		}
	}
}

func (w *Sync) run(ctx context.Context) error {
	log := logger.FromContext(ctx)

	assets, err := w.assets.ListAll(ctx)
	if err != nil {
		log.WithError(err).Error("assets.ListAll")
		return err
	}

	for _, asset := range assets {
		assetz, err := w.assetz.Find(ctx, asset.ID)
		if err != nil {
			log.WithError(err).Errorf("assetz.Find(%s)", asset.ID)
			continue
		}

		if asset.Price.Equal(assetz.Price) {
			continue
		}

		if err := w.assets.Save(ctx, assetz, "price"); err != nil {
			log.WithError(err).Errorf("assets.Update(%s)", asset.ID)
			continue
		}
	}

	return nil
}
