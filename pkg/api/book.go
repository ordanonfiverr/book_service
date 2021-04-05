package api

type Book struct {
	Title string `json:"title"`
	AuthorName string `json:"author name"`
	Price float32 `json:"price"`
	EbookAvailable bool `json:"ebook available"`
	PublishDate string `json:"publish date"`
	Username bool `json:"username"`
}
