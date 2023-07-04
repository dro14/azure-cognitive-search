package types

type PaymeRequest struct {
	Method string `json:"method"`
	Params Params `json:"params"`
	ID     int    `json:"id"`
}

type Params struct {
	ID      string  `json:"id"`
	Time    int64   `json:"time"`
	Amount  int     `json:"amount"`
	Account Account `json:"account"`
	Reason  int     `json:"reason"`
	From    int64   `json:"from"`
	To      int64   `json:"to"`
}

type Account struct {
	OrderID string `json:"order_id"`
}
