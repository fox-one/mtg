package asset

import (
	"context"

	"github.com/fox-one/mtg/core"
	"github.com/fox-one/mtg/pkg/number"
	"github.com/fox-one/pkg/lruset"
	"github.com/fox-one/pkg/store/db"
)

func init() {
	db.RegisterMigrate(func(db *db.DB) error {
		tx := db.Update().Model(core.Asset{})

		if err := tx.AutoMigrate(core.Asset{}).Error; err != nil {
			return err
		}

		return nil
	})
}

func New(db *db.DB) core.AssetStore {
	return &assetStore{
		db: db,
	}
}

type assetStore struct {
	db *db.DB
}

func (s *assetStore) Save(ctx context.Context, asset *core.Asset, columns ...string) error {
	return s.db.Tx(func(tx *db.DB) error {
		rows, err := update(tx, asset, columns...)
		if err != nil {
			return err
		}

		if rows == 0 {
			return tx.Update().Create(asset).Error
		}

		return nil
	})
}

func (s *assetStore) Find(ctx context.Context, id string) (*core.Asset, error) {
	var asset core.Asset
	if err := s.db.View().Where("id = ?", id).Take(&asset).Error; err != nil {
		return nil, err
	}

	return &asset, nil
}

func (s *assetStore) ListAll(ctx context.Context) ([]*core.Asset, error) {
	var assets []*core.Asset
	if err := s.db.View().Find(&assets).Error; err != nil {
		return nil, err
	}

	return assets, nil
}

func toUpdateParams(asset *core.Asset) map[string]interface{} {
	return map[string]interface{}{
		"price": asset.Price,
		// "display_symbol": asset.DisplaySymbol,
		// "promote_label":  asset.PromoteLabel,
		// "promote_url":    asset.PromoteURL,
		// "promote_color":  asset.PromoteColor,
	}
}

func update(tx *db.DB, asset *core.Asset, columns ...string) (int64, error) {
	updates := toUpdateParams(asset)
	if set := setWithColumns(columns); set != nil {
		for k := range updates {
			if !set.Contains(k) {
				delete(updates, k)
			}
		}
	}

	u := tx.Update().Model(asset).Updates(updates)
	return u.RowsAffected, u.Error
}

func (s *assetStore) Update(ctx context.Context, asset *core.Asset) error {
	_, err := update(s.db, asset)
	return err
}

func (s *assetStore) ListPrices(ctx context.Context, ids ...string) (number.Values, error) {
	var assets []*core.Asset
	if err := s.db.View().Select("id, price").Where("id IN (?)", ids).Find(&assets).Error; err != nil {
		return nil, err
	}

	prices := make(number.Values, len(assets))
	for _, asset := range assets {
		prices[asset.ID] = asset.Price
	}

	return prices, nil
}

func setWithColumns(columns []string) *lruset.Set {
	if len(columns) == 0 {
		return nil
	}

	set := lruset.New(len(columns))
	for _, column := range columns {
		set.Add(column)
	}

	return set
}
