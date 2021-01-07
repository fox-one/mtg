package wallet

import (
	"context"
	"fmt"

	"github.com/fox-one/mixin-sdk-go"
	"github.com/fox-one/mtg/core"
)

func MustState(walletz core.WalletService, state string) core.WalletService {
	return &mustState{
		WalletService: walletz,
		state:         state,
	}
}

func MustSpent(walletz core.WalletService) core.WalletService {
	return MustState(walletz, mixin.UTXOStateSpent)
}

type mustState struct {
	core.WalletService
	state string
}

func (s *mustState) Spend(ctx context.Context, outputs []*core.Output, transfer *core.Transfer) (*core.RawTransaction, error) {
	for _, output := range outputs {
		if output.State != s.state {
			return nil, fmt.Errorf("state %q not allowed, must %q", output.State, s.state)
		}
	}

	return s.WalletService.Spend(ctx, outputs, transfer)
}
