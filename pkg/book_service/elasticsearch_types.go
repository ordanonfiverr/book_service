package book_service

type ElasticSearchResponse struct {
	Id string `json:"_id"`
	Source interface{} `json:"_source"`
}

type ElasticSearchPostResponse struct {
	Id string `json:"_id"`
}
