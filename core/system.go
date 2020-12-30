package core

import (
	"crypto/ed25519"

	"github.com/shopspring/decimal"
)

// System stores system information.
type System struct {
	Admins     []string
	ClientID   string
	Members    []*Member
	Threshold  uint8
	VoteAsset  string
	VoteAmount decimal.Decimal
	PrivateKey ed25519.PrivateKey
	SignKey    ed25519.PrivateKey
	Version    string
}

func (s *System) MemberIDs() []string {
	ids := make([]string, len(s.Members))
	for idx, m := range s.Members {
		ids[idx] = m.ClientID
	}

	return ids
}
