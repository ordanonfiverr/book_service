package book_service

import (
	"book_service/pkg/api"
	"book_service/pkg/consts"
	"context"
	"encoding/json"
	"errors"
	"github.com/olivere/elastic/v6"
	"net/http"
)

const (
	booksIndex = "books"
	bookType   = "book"
)

type BookService struct {
	elasticClient *elastic.Client
}

func NewBookService(elasticClient *elastic.Client) *BookService {
	return &BookService{
		elasticClient: elasticClient,
	}
}

func (b *BookService) StoreStats(ctx context.Context) (*api.StoreStats, error) {
	const authorsDcountAggName = "authors_dcount_i"

	searchResult, err := b.elasticClient.Search().
		Index(booksIndex).
		Size(120).
		Aggregation(authorsDcountAggName, elastic.NewCardinalityAggregation().Field("author_name.keyword")).
		Query(elastic.NewMatchAllQuery()).
		Do(ctx)

	if err != nil {
		return nil, ConvertError(err)
	}
	authorsDcountAgg, found := searchResult.Aggregations.Cardinality(authorsDcountAggName)
	if !found {
		return nil, errors.New("missing agg result")
	}

	return &api.StoreStats{
		Count:  searchResult.Hits.TotalHits,
		Dcount: int(*authorsDcountAgg.Value),
	}, nil
}

func (b *BookService) SearchBooks(ctx context.Context, title string, authorName string, minPrice float32, maxPrice float32) ([]*api.Book, error) {
	searchResult, err := b.elasticClient.Search().
		Index(booksIndex).
		Query(elastic.NewBoolQuery().
			Must(elastic.NewMatchPhraseQuery("title", title)).
			Must(elastic.NewMatchPhraseQuery("author_name", authorName)).
			Filter(elastic.NewRangeQuery("price").Gte(minPrice).Lte(maxPrice))).
		Do(ctx)

	if err != nil {
		return nil, ConvertError(err)
	}
	if searchResult.Hits == nil || searchResult.Hits.Hits == nil {
		return nil, errors.New("missing agg result")
	}

	books := make([]*api.Book, len(searchResult.Hits.Hits))
	for i, hit := range searchResult.Hits.Hits {
		book, err := ConvertElasticBookResponseToBook(hit.Id, hit.Source)
		if err != nil {
			return nil, err
		}
		books[i] = book
	}
	return books, nil
}

func (b *BookService) DeleteBook(ctx context.Context, id string) error {
	_, err := b.elasticClient.Delete().
		Index(booksIndex).
		Type(bookType).
		Id(id).
		Do(ctx)
	return ConvertError(err)
}

func (b *BookService) UpdateBookTitle(ctx context.Context, id string, title string) error {
	_, err := b.elasticClient.Update().
		Index(booksIndex).
		Type(bookType).
		Id(id).
		Doc(map[string]string{"title": title}).
		Do(ctx)

	return ConvertError(err)
}

func (b *BookService) AddBook(ctx context.Context, book *api.Book) (id string, err error) {
	indexResponse, err := b.elasticClient.Index().
		Index(booksIndex).
		Type(bookType).
		BodyJson(book).
		Do(ctx)
	if err != nil {
		return "", ConvertError(err)
	}

	return indexResponse.Id, nil
}

func (b *BookService) GetBook(ctx context.Context, id string) (*api.Book, error) {
	getResponse, err := b.elasticClient.Get().
		Index(booksIndex).
		Type(bookType).
		Id(id).
		Do(ctx)
	if err != nil {
		return nil, ConvertError(err)
	}

	book, err := ConvertElasticBookResponseToBook(getResponse.Id, getResponse.Source)
	if err != nil {
		return nil, err
	}

	return book, nil
}

func ConvertElasticBookResponseToBook(id string, raw *json.RawMessage) (*api.Book, error) {
	if raw == nil {
		return nil, errors.New("wrong response format")
	}
	book := &api.Book{}
	if err := json.Unmarshal(*raw, book); err != nil {
		return nil, err
	}

	book.Id = id
	return book, nil
}

func ConvertError(err error) error {
	if elasticErr, ok := err.(*elastic.Error); ok && elasticErr.Status == http.StatusNotFound {
		return consts.NotFoundErr
	}
	return err
}
