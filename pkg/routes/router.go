package routes

import (
	"book_service/pkg/handlers"
	"github.com/gin-gonic/gin"
)

func AddRoutes(r *gin.Engine, handler *handlers.Handler) {
	r.Use(handler.CollectUserActivity)

	r.GET("/activity", handler.Activity)

	r.GET("/store", handler.Store)

	r.GET("/search", handler.Search)

	r.PUT("/book", handler.CreateBook)

	r.POST("/book/:id", handler.UpdateBookTitle)

	r.DELETE("/book/:id", handler.DeleteBook)

	r.GET("/book/:id", handler.GetBook)
}
