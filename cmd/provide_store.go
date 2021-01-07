package cmd

import (
	"github.com/fox-one/mtg/core"
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

func provideWalletStore(db *db.DB) core.WalletStore {
	return wallet.New(db)
}
