package api

import (
	"context"
	"fmt"

	"github.com/PineappleBond/TradingEino/backend/pkg/okex"
	tradeRequests "github.com/PineappleBond/TradingEino/backend/pkg/okex/requests/rest/trade"
)

// OrderResult represents the result of placing an order.
type OrderResult struct {
	OrderID string `json:"order_id"`
	State   string `json:"state"`
	SMsg    string `json:"s_msg"` // Error message if failed
}

// CancelOrderResult represents the result of canceling an order.
type CancelOrderResult struct {
	OrderID string `json:"order_id"`
	State   string `json:"state"`
	SMsg    string `json:"s_msg"` // Error message if failed
}

// OrderDetails represents detailed order information.
type OrderDetails struct {
	OrderID  string  `json:"order_id"`
	InstID   string  `json:"inst_id"`
	Side     string  `json:"side"`
	PosSide  string  `json:"pos_side"`
	OrdType  string  `json:"ord_type"`
	Size     float64 `json:"size"`
	Price    float64 `json:"price"`
	AvgPrice float64 `json:"avg_price"`
	State    string  `json:"state"`
	FillSize float64 `json:"fill_size"`
}

// PlaceOrder places a new order via OKX REST API.
// Parameters:
//   - instID: Instrument ID (e.g., "ETH-USDT-SWAP")
//   - side: Order side ("buy" or "sell")
//   - posSide: Position side ("long", "short", or "net")
//   - ordType: Order type ("market", "limit", etc.)
//   - size: Order size
//   - price: Order price (empty for market orders)
//
// Returns the order ID and state if successful.
func (c *Client) PlaceOrder(
	ctx context.Context,
	instID, side, posSide, ordType, size, price string,
) (*OrderResult, error) {
	req := []tradeRequests.PlaceOrder{
		{
			InstID:  instID,
			Side:    okex.OrderSide(side),
			PosSide: okex.PositionSide(posSide),
			OrdType: okex.OrderType(ordType),
			Sz:      size,
			Px:      price,
			TdMode:  okex.TradeCrossMode, // Use cross margin mode by default
		},
	}

	resp, err := c.Rest.Trade.PlaceOrder(req)
	if err != nil {
		return nil, fmt.Errorf("place order: %w", err)
	}

	// Check response code
	if resp.Code != 0 {
		return nil, fmt.Errorf("place order failed: code=%d, msg=%s", resp.Code, resp.Msg)
	}

	if len(resp.PlaceOrders) == 0 {
		return nil, fmt.Errorf("place order failed: empty response")
	}

	result := resp.PlaceOrders[0]

	// Check for order-level errors
	if result.SCode != 0 {
		return nil, fmt.Errorf("order placement error: code=%d, msg=%s", result.SCode, result.SMsg)
	}

	return &OrderResult{
		OrderID: result.OrdID,
		State:   "pending",
		SMsg:    result.SMsg,
	}, nil
}

// CancelOrder cancels an existing order via OKX REST API.
// Parameters:
//   - instID: Instrument ID (e.g., "ETH-USDT-SWAP")
//   - orderID: OKX order ID to cancel
//
// Returns the cancellation result.
func (c *Client) CancelOrder(ctx context.Context, instID, orderID string) (*CancelOrderResult, error) {
	req := []tradeRequests.CancelOrder{
		{
			InstID: instID,
			OrdID:  orderID,
		},
	}

	resp, err := c.Rest.Trade.CandleOrder(req)
	if err != nil {
		return nil, fmt.Errorf("cancel order: %w", err)
	}

	// Check response code
	if resp.Code != 0 {
		return nil, fmt.Errorf("cancel order failed: code=%d, msg=%s", resp.Code, resp.Msg)
	}

	if len(resp.CancelOrders) == 0 {
		return nil, fmt.Errorf("cancel order failed: empty response")
	}

	result := resp.CancelOrders[0]

	// Check for order-level errors
	if result.SCode != 0 {
		return nil, fmt.Errorf("order cancellation error: code=%v, msg=%s", result.SCode, result.SMsg)
	}

	return &CancelOrderResult{
		OrderID: result.OrdID,
		State:   "cancelled",
		SMsg:    result.SMsg,
	}, nil
}

// GetOrderDetails retrieves detailed information about an order via OKX REST API.
// Parameters:
//   - instID: Instrument ID (e.g., "ETH-USDT-SWAP")
//   - orderID: OKX order ID to query
//
// Returns the order details.
func (c *Client) GetOrderDetails(ctx context.Context, instID, orderID string) (*OrderDetails, error) {
	req := tradeRequests.OrderDetails{
		InstID: instID,
		OrdID:  orderID,
	}

	resp, err := c.Rest.Trade.GetOrderDetail(req)
	if err != nil {
		return nil, fmt.Errorf("get order details: %w", err)
	}

	// Check response code
	if resp.Code != 0 {
		return nil, fmt.Errorf("get order details failed: code=%d, msg=%s", resp.Code, resp.Msg)
	}

	if len(resp.Orders) == 0 {
		return nil, fmt.Errorf("order not found: %s", orderID)
	}

	order := resp.Orders[0]

	return &OrderDetails{
		OrderID:  order.OrdID,
		InstID:   order.InstID,
		Side:     string(order.Side),
		PosSide:  string(order.PosSide),
		OrdType:  string(order.OrdType),
		Size:     float64(order.Sz),
		Price:    float64(order.Px),
		AvgPrice: float64(order.AvgPx),
		State:    string(order.State),
		FillSize: float64(order.AccFillSz),
	}, nil
}
