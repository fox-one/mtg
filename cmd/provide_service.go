package cmd

import (
	"fmt"

	"github.com/fox-one/mixin-sdk-go"
	"github.com/fox-one/mtg/core"
	"github.com/fox-one/mtg/pkg/mtg"
	"github.com/fox-one/mtg/service/asset"
	"github.com/fox-one/mtg/service/wallet"
)

func provideMixinClient() *mixin.Client {
	c, err := mixin.NewFromKeystore(&cfg.Dapp.Keystore)
	if err != nil {
		panic(err)
	}

	return c
}

func provideSystem() *core.System {
	members := make([]*core.Member, 0, len(cfg.Group.Members))
	for _, m := range cfg.Group.Members {
		verifyKey, err := mtg.DecodePublicKey(m.VerifyKey)
		if err != nil {
			panic(fmt.Errorf("decode verify key for member %s failed", m.ClientID))
		}

		members = append(members, &core.Member{
			ClientID:  m.ClientID,
			VerifyKey: verifyKey,
		})
	}

	privateKey, err := mtg.DecodePrivateKey(cfg.Group.PrivateKey)
	if err != nil {
		panic(fmt.Errorf("base64 decode group private key failed: %w", err))
	}

	signKey, err := mtg.DecodePrivateKey(cfg.Group.SignKey)
	if err != nil {
		panic(fmt.Errorf("base64 decode group sign key failed: %w", err))
	}

	return &core.System{
		Admins:     cfg.Group.Admins,
		ClientID:   cfg.Dapp.ClientID,
		Members:    members,
		Threshold:  cfg.Group.Threshold,
		VoteAsset:  cfg.Group.Vote.Asset,
		VoteAmount: cfg.Group.Vote.Amount,
		PrivateKey: privateKey,
		SignKey:    signKey,
	}
}

func provideAssetService(client *mixin.Client) core.AssetService {
	return asset.New(client)
}

func provideWalletService(client *mixin.Client) core.WalletService {
	members := make([]string, len(cfg.Group.Members))
	for idx, m := range cfg.Group.Members {
		members[idx] = m.ClientID
	}

	return wallet.New(client, wallet.Config{
		Pin:       cfg.Dapp.Pin,
		Members:   members,
		Threshold: cfg.Group.Threshold,
	})
}
