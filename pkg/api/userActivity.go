package api

type UserActivity struct {
	Method     string `json:"method"`
	RequestUri string `json:"request_uri"`
}
