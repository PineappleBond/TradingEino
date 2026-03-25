package private

import (
	"github.com/PineappleBond/TradingEino/backend/pkg/okex/events"
	"github.com/PineappleBond/TradingEino/backend/pkg/okex/models/account"
	"github.com/PineappleBond/TradingEino/backend/pkg/okex/models/trade"
)

type (
	Account struct {
		Arg      *events.Argument   `json:"arg"`
		Balances []*account.Balance `json:"data"`
	}
	Position struct {
		Arg       *events.Argument    `json:"arg"`
		Positions []*account.Position `json:"data"`
	}
	BalanceAndPosition struct {
		Arg                 *events.Argument              `json:"arg"`
		BalanceAndPositions []*account.BalanceAndPosition `json:"data"`
	}
	Order struct {
		Arg    *events.Argument `json:"arg"`
		Orders []*trade.Order   `json:"data"`
	}
)
