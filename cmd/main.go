package main

import (
	"book_service/pkg/book_service"
	"book_service/pkg/routes"
	"github.com/gin-gonic/gin"
	redis "gopkg.in/redis.v5"
)

func main() {
	bookService := book_service.NewBookService("http://10.200.10.1:9200")

	client := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	r := gin.Default()
	routes.AddRoutes(r, client, bookService)

	r.Run("0.0.0.0:8081") // listen and serve on 0.0.0.0:8080 (for windows "localhost:8080")
}
