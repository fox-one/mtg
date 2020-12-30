package notifier

import (
	"context"

	"github.com/fox-one/mtg/core"
)

func Mute() core.Notifier {
	return &dumb{}
}

type dumb struct{}

func (d *dumb) Snapshot(ctx context.Context, transfer *core.Transfer, signedTx string) error {
	return nil
}
