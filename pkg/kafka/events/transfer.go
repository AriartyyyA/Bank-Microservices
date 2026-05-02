package events

type TransferEvent struct {
	TransactionID string `json:"transaction_id"`
	FromWalletID  string `json:"from_wallet_id"`
	ToWalletID    string `json:"to_wallet_id"`
	Amount        int64  `json:"amount"`
}
