package db

type AccountInfo struct {
	Account       string `json:"account"`
	Confirmed     int    `json:"confirmed"`
	Unconfirmed   int    `json:"unconfirmed"`
	Locked        int    `json:"locked"`
	TotalSent     int    `json:"total_sent"`
	TotalReceived int    `json:"total_received"`
}
