package cashier

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/fox-one/mixin-sdk-go"
	"github.com/fox-one/mtg/core"
	"github.com/fox-one/pkg/logger"
	"github.com/fox-one/pkg/uuid"
	"github.com/shopspring/decimal"
)

func New(
	wallets core.WalletStore,
	walletz core.WalletService,
	system *core.System,
) *Cashier {
	return &Cashier{
		wallets: wallets,
		walletz: walletz,
		system:  system,
	}
}

type Cashier struct {
	wallets core.WalletStore
	walletz core.WalletService
	system  *core.System
}

func (w *Cashier) Run(ctx context.Context) error {
	log := logger.FromContext(ctx).WithField("worker", "cashier")
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
				dur = 300 * time.Millisecond
			}
		}
	}
}

func (w *Cashier) run(ctx context.Context) error {
	log := logger.FromContext(ctx)

	transfers, err := w.wallets.ListPendingTransfers(ctx)
	if err != nil {
		log.WithError(err).Errorln("list transfers")
		return err
	}

	if len(transfers) == 0 {
		return errors.New("EOF")
	}

	for _, transfer := range transfers {
		_ = w.handleTransfer(ctx, transfer)
	}

	return nil
}

func (w *Cashier) handleTransfer(ctx context.Context, transfer *core.Transfer) error {
	log := logger.FromContext(ctx).WithField("transfer", transfer.TraceID)

	const limit = 32
	outputs, err := w.wallets.ListUnspent(ctx, transfer.AssetID, limit)
	if err != nil {
		log.WithError(err).Errorln("wallets.ListUnspent")
		return err
	}

	var (
		idx    int
		sum    decimal.Decimal
		traces []string
	)

	for _, output := range outputs {
		sum = sum.Add(output.Amount)
		traces = append(traces, output.TraceID)
		idx += 1

		if sum.GreaterThanOrEqual(transfer.Amount) {
			break
		}
	}

	outputs = outputs[:idx]

	if sum.LessThan(transfer.Amount) {
		// merge outputs
		if len(outputs) == limit {
			traceID := uuid.Modify(transfer.TraceID, mixin.HashMembers(traces))
			merge := &core.Transfer{
				TraceID:   traceID,
				AssetID:   transfer.AssetID,
				Amount:    sum,
				Opponents: w.system.MemberIDs(),
				Threshold: w.system.Threshold,
				Memo:      fmt.Sprintf("merge for %s", transfer.TraceID),
			}

			return w.spent(ctx, outputs, merge)
		}

		err := errors.New("insufficient balance")
		log.WithError(err).Errorln("handle transfer", transfer.ID)
		return err
	}

	return w.spent(ctx, outputs, transfer)
}

func (w *Cashier) spent(ctx context.Context, outputs []*core.Output, transfer *core.Transfer) error {
	if tx, err := w.walletz.Spent(ctx, outputs, transfer); err != nil {
		logger.FromContext(ctx).WithError(err).Errorln("walletz.Spent")
		return err
	} else if tx != nil {
		// 签名收集完成，需要提交至主网
		// 此时将该上链 tx 存储至数据库，等待 tx sender worker 完成上链
		if err := w.wallets.CreateRawTransaction(ctx, tx); err != nil {
			logger.FromContext(ctx).WithError(err).Errorln("wallets.CreateRawTransaction")
			return err
		}
	}

	if err := w.wallets.Spent(ctx, outputs, transfer); err != nil {
		logger.FromContext(ctx).WithError(err).Errorln("wallets.Spent")
		return err
	}

	return nil
}
