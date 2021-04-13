package api

type StoreStats struct {
	Count  int64 `json:"count"`
	Dcount int   `json:"dcount"`
}
