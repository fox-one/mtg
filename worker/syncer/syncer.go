package syncer

import (
	"context"
	"encoding/base64"
	"errors"
	"time"

	"github.com/fox-one/mtg/core"
	"github.com/fox-one/mtg/internal/mixinet"
	"github.com/fox-one/mtg/pkg/mtg"
	"github.com/fox-one/pkg/logger"
	"github.com/fox-one/pkg/property"
	"github.com/gofrs/uuid"
)

const checkpointKey = "mtg_sync_checkpoint"

func New(
	assets core.AssetStore,
	assetz core.AssetService,
	wallets core.WalletStore,
	walletz core.WalletService,
	property property.Store,
	system *core.System,
) *Syncer {
	return &Syncer{
		assets:   assets,
		assetz:   assetz,
		wallets:  wallets,
		walletz:  walletz,
		property: property,
		assetMap: map[string]bool{},
		system:   system,
	}
}

type Syncer struct {
	assets   core.AssetStore
	assetz   core.AssetService
	wallets  core.WalletStore
	walletz  core.WalletService
	property property.Store
	system   *core.System
	assetMap map[string]bool
}

func (w *Syncer) Run(ctx context.Context) error {
	log := logger.FromContext(ctx).WithField("worker", "syncer")
	ctx = logger.WithContext(ctx, log)

	dur := time.Millisecond

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(dur):
			if err := w.run(ctx); err == nil {
				dur = 100 * time.Millisecond
			} else {
				dur = 500 * time.Millisecond
			}
		}
	}
}

func (w *Syncer) run(ctx context.Context) error {
	log := logger.FromContext(ctx)

	v, err := w.property.Get(ctx, checkpointKey)
	if err != nil {
		log.WithError(err).Errorln("property.Get", checkpointKey)
		return err
	}

	offset := v.Time()

	var (
		outputs   = make([]*core.Output, 0, 8)
		positions = make(map[string]int)
		pos       = 0
	)

	const Limit = 500

	for {
		batch, err := w.walletz.Pull(ctx, offset, Limit)
		if err != nil {
			log.WithError(err).Errorln("walletz.Pull")
			return err
		}

		log.Debugln("pull updated outputs", len(batch), offset)

		for _, u := range batch {
			offset = u.UpdatedAt

			p, ok := positions[u.TraceID]
			if ok {
				outputs[p] = u
				continue
			}

			outputs = append(outputs, u)
			positions[u.TraceID] = pos
			pos += 1
		}

		if len(batch) < Limit {
			break
		}
	}

	if len(outputs) == 0 {
		return errors.New("EOF")
	}

	for _, output := range outputs {
		// save asset
		if _, f := w.assetMap[output.AssetID]; !f {
			if asset, err := w.assetz.Find(ctx, output.AssetID); err == nil {
				if err := w.assets.Save(ctx, asset); err == nil {
					w.assetMap[output.AssetID] = true
				}
			}
		}

		// decrypt memo
		if data := decodeMemo(output.Memo); len(data) > 0 {
			if member, content, err := core.DecodeMemberAction(data, w.system.Members); err == nil {
				output.MemoData.Member, output.MemoData.Data = member.ClientID, content
			} else if content, err := mtg.Decrypt(data, w.system.PrivateKey); err == nil {
				var (
					typ      int
					userID   uuid.UUID
					followID uuid.UUID
				)

				if left, err := mtg.Scan(content, &typ, &userID, &followID); err == nil {
					output.MemoData.Type = typ
					output.MemoData.UserID = userID.String()
					output.MemoData.FollowID = followID.String()
					output.MemoData.Data = left
				} else {
					output.MemoData.Data = content
				}
			}
		}
	}

	mixinet.SortOutputs(outputs)
	if err := w.wallets.Save(ctx, outputs); err != nil {
		log.WithError(err).Errorln("wallets.Save")
		return err
	}

	if err := w.property.Save(ctx, checkpointKey, offset); err != nil {
		log.WithError(err).Errorln("property.Save", checkpointKey)
		return err
	}

	return nil
}

func decodeMemo(memo string) []byte {
	if b, err := base64.StdEncoding.DecodeString(memo); err == nil {
		return b
	}

	if b, err := base64.URLEncoding.DecodeString(memo); err == nil {
		return b
	}

	return []byte(memo)
}
