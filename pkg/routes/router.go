package routes

import (
	"book_service/pkg/api"
	"book_service/pkg/book_service"
	"book_service/pkg/errors"
	"fmt"
	"github.com/gin-gonic/gin"
	redis "gopkg.in/redis.v5"
	"net/http"
	"strconv"
)

func AddRoutes(r *gin.Engine, redisClient *redis.Client, bookService *book_service.BookService) {
	r.Use(func(c *gin.Context) {
		// Store last 3 actions made by each user
		key := c.Query("user")
		if c.FullPath() == "/activity" {
			// avoid logging the activity api
			return
		}

		redisClient.LPush(key, fmt.Sprintf("{ 'Method':'%s', 'RequestUri':'%s'}",
			c.Request.Method, c.Request.URL.RequestURI()))
		redisClient.LTrim(key, 0, 2)
	})

	r.GET("/activity", func(c *gin.Context) {
		user := c.Query("user")
		lastActions, err := redisClient.LRange(user, 0, 2).Result()
		if err != nil {
			handleError(c, err)
			return
		}
		c.JSON(http.StatusOK, lastActions)
	})

	r.GET("/store", func(c *gin.Context) {
		resp, err := bookService.StoreStats()
		if err != nil {
			handleError(c, err)
			return
		}

		c.JSON(http.StatusOK, resp)
	})

	r.GET("/search", func(c *gin.Context) {
		title := c.Query("title")
		author := c.Query("author")
		minPrice, err := strconv.ParseFloat(c.Query("min-price"), 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, "invalid value for min-price query parameter")
			return
		}
		maxPrice, err := strconv.ParseFloat(c.Query("max-price"), 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, "invalid value for max-price query parameter")
			return
		}

		resp, err := bookService.SearchBooks(title, author, float32(minPrice), float32(maxPrice))
		if err != nil {
			handleError(c, err)
			return
		}

		c.JSON(http.StatusOK, resp)
	})

	r.PUT("/book", func(c *gin.Context) {
		book := &api.Book{}
		if err := c.ShouldBindJSON(book); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		id, err := bookService.AddBook(book)
		if err != nil {
			handleError(c, err)
			return
		}
		c.JSON(http.StatusOK, id)
	})

	r.POST("/book/:id", func(c *gin.Context) {
		id := c.Param("id")
		newTitle := c.Query("title")

		if err := bookService.UpdateBookTitle(id, newTitle); err != nil {
			handleError(c, err)
			return
		}
		c.JSON(http.StatusOK, "success")
	})

	r.DELETE("/book/:id", func(c *gin.Context) {
		id := c.Param("id")
		if err := bookService.DeleteBook(id); err != nil {
			handleError(c, err)
			return
		}
		c.JSON(http.StatusOK, "deleted")
	})

	r.GET("/book/:id", func(c *gin.Context) {
		id := c.Param("id")
		book, err := bookService.GetBook(id)
		if err != nil {
			handleError(c, err)
			return
		}

		c.JSON(200, book)
	})
}

func handleError(c *gin.Context, err error) {
	if httpErr, ok := err.(*errors.HttpError); ok && httpErr.Code == http.StatusNotFound {
		c.JSON(http.StatusNotFound, err)
	} else {
		c.JSON(http.StatusInternalServerError, err)
	}
}
