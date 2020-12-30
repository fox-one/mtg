package asset

import (
	"context"
	"time"

	"github.com/fox-one/mtg/core"
	"github.com/patrickmn/go-cache"
	"golang.org/x/sync/singleflight"
)

func Cache(store core.AssetStore, exp time.Duration) core.AssetStore {
	return &cacheAssetStore{
		AssetStore: store,
		cache:      cache.New(exp, cache.NoExpiration),
		sf:         &singleflight.Group{},
	}
}

type cacheAssetStore struct {
	core.AssetStore
	cache *cache.Cache
	sf    *singleflight.Group
}

func (s *cacheAssetStore) Save(ctx context.Context, asset *core.Asset, columns ...string) error {
	if err := s.AssetStore.Save(ctx, asset, columns...); err != nil {
		return err
	}

	s.cache.Delete(asset.ID)
	return nil
}

func (s *cacheAssetStore) Find(ctx context.Context, id string) (*core.Asset, error) {
	v, err, _ := s.sf.Do(id, func() (interface{}, error) {
		if v, ok := s.cache.Get(id); ok {
			return v, nil
		}

		asset, err := s.AssetStore.Find(ctx, id)
		if err != nil {
			return nil, err
		}

		s.cache.SetDefault(id, asset)
		return asset, nil
	})

	if err != nil {
		return nil, err
	}

	return v.(*core.Asset), nil
}
