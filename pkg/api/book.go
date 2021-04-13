package api

type Book struct {
	Id             string  `json:"id, omitempty"`
	Title          string  `json:"title"`
	AuthorName     string  `json:"author_name"`
	Price          float32 `json:"price"`
	EbookAvailable bool    `json:"ebook_available"`
	PublishDate    string  `json:"publish_date"`
}
