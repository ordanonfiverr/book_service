package handlers

import (
	"book_service/pkg/api"
	"book_service/pkg/book_service"
	"book_service/pkg/errors"
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"gopkg.in/redis.v5"
	"net/http"
	"strconv"
)

type Handler struct {
	redisClient *redis.Client
	bookService *book_service.BookService
}

func NewHandler(redisClient *redis.Client, bookService *book_service.BookService) *Handler {
	return &Handler{
		redisClient: redisClient,
		bookService: bookService,
	}
}

func (h *Handler) CollectUserActivity(c *gin.Context) {
	// Store last 3 actions made by each user
	key := c.Query("user")
	if c.FullPath() == "/activity" {
		// avoid logging the activity api
		return
	}

	h.redisClient.LPush(key, fmt.Sprintf("{ 'Method':'%s', 'RequestUri':'%s'}",
		c.Request.Method, c.Request.URL.RequestURI()))
	h.redisClient.LTrim(key, 0, 2)
}

func (h *Handler) Activity(c *gin.Context) {
	user := c.Query("user")
	lastActions, err := h.redisClient.LRange(user, 0, 2).Result()
	if err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, lastActions)
}

func (h *Handler) Store(c *gin.Context) {
	resp, err := h.bookService.StoreStats(context.Background())
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, resp)
}

func (h *Handler) Search(c *gin.Context) {
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

	resp, err := h.bookService.SearchBooks(context.Background(), title, author, float32(minPrice), float32(maxPrice))
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, resp)
}

func (h *Handler) CreateBook(c *gin.Context) {
	book := &api.Book{}
	if err := c.ShouldBindJSON(book); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	id, err := h.bookService.AddBook(context.Background(), book)
	if err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, id)
}

func (h *Handler) UpdateBookTitle(c *gin.Context) {
	id := c.Param("id")
	newTitle := c.Query("title")

	if err := h.bookService.UpdateBookTitle(context.Background(), id, newTitle); err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, "success")
}

func (h *Handler) DeleteBook(c *gin.Context) {
	id := c.Param("id")
	if err := h.bookService.DeleteBook(context.Background(), id); err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, "deleted")
}

func (h *Handler) GetBook(c *gin.Context) {
	id := c.Param("id")
	book, err := h.bookService.GetBook(context.Background(), id)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, book)
}

func handleError(c *gin.Context, err error) {
	if httpErr, ok := err.(*errors.HttpError); ok && httpErr.Code == http.StatusNotFound {
		c.JSON(http.StatusNotFound, err)
	} else {
		c.JSON(http.StatusInternalServerError, err)
	}
}
