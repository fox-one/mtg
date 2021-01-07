package cmd

import (
	"time"

	"github.com/fox-one/mtg/core"
	"github.com/fox-one/mtg/store/asset"
	"github.com/fox-one/mtg/store/message"
	"github.com/fox-one/mtg/store/wallet"
	"github.com/fox-one/pkg/property"
	"github.com/fox-one/pkg/store/db"
	propertystore "github.com/fox-one/pkg/store/property"
)

func provideDatabase() *db.DB {
	return db.MustOpen(cfg.DB)
}

func providePropertyStore(db *db.DB) property.Store {
	return propertystore.New(db)
}

func provideAssetStore(db *db.DB, exp time.Duration) core.AssetStore {
	assets := asset.New(db)
	if exp > 0 {
		assets = asset.Cache(assets, exp)
	}

	return assets
}

func provideMessageStore(db *db.DB) core.MessageStore {
	return message.New(db)
}

func provideWalletStore(db *db.DB) core.WalletStore {
	return wallet.New(db)
}
