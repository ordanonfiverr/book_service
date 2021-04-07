package api

type Book struct {
	Title          string  `json:"title"`
	AuthorName     string  `json:"author_name"`
	Price          float32 `json:"price"`
	EbookAvailable bool    `json:"ebook_available"`
	PublishDate    string  `json:"publish_date"`
	Username       bool    `json:"username"`
}
