package book_service

import (
	"book_service/pkg/api"
	"book_service/pkg/errors"
	"context"
	"encoding/json"
	"github.com/olivere/elastic/v6"
	"net/http"
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
	const authorsDcountAggName = "authors-dcount"

	searchResult, err := b.elasticClient.Search().
		Index("books").
		Size(120).
		Aggregation(authorsDcountAggName, elastic.NewCardinalityAggregation().Field("author's name.keyword")).
		Query(elastic.NewMatchAllQuery()).
		Do(ctx)

	if err != nil {
		return nil, ConvertError(err)
	}
	authorsDcountAgg, found := searchResult.Aggregations.Cardinality(authorsDcountAggName)
	if !found {
		return nil, errors.NewHttpError(http.StatusInternalServerError, "missing agg result", nil)
	}

	return &api.StoreStats{
		Count:  searchResult.Hits.TotalHits,
		Dcount: int(*authorsDcountAgg.Value),
	}, nil
}

func (b *BookService) SearchBooks(ctx context.Context, title string, authorName string, minPrice float32, maxPrice float32) (interface{}, error) {
	searchResult, err := b.elasticClient.Search().
		Index("books").
		Query(elastic.NewBoolQuery().
			Must(elastic.NewMatchPhraseQuery("title", title)).
			Must(elastic.NewMatchPhraseQuery("author's name", authorName)).
			Filter(elastic.NewRangeQuery("price").Gte(minPrice).Lte(maxPrice))).
		Do(ctx)

	if err != nil {
		return nil, ConvertError(err)
	}
	if searchResult.Hits == nil || searchResult.Hits.Hits == nil {
		return nil, errors.NewHttpError(http.StatusInternalServerError, "missing results", nil)
	}

	return searchResult.Hits.Hits, nil
}

func (b *BookService) DeleteBook(ctx context.Context, id string) error {
	_, err := b.elasticClient.Delete().
		Index("books").
		Type("book").
		Id(id).
		Do(ctx)
	return ConvertError(err)
}

func (b *BookService) UpdateBookTitle(ctx context.Context, id string, title string) error {
	_, err := b.elasticClient.Update().
		Index("books").
		Type("book").
		Id(id).
		Doc(map[string]string{ "title": title}).
		Do(ctx)

	return ConvertError(err)
}

func (b *BookService) AddBook(ctx context.Context, book *api.Book) (id string, err error) {
	indexResponse, err := b.elasticClient.Index().
		Index("books").
		Type("book").
		BodyJson(book).
		Do(ctx)
	if err != nil {
		return "", ConvertError(err)
	}

	return indexResponse.Id, nil
}

func (b *BookService) GetBook(ctx context.Context, id string) (*json.RawMessage, error) {
	getResponse, err := b.elasticClient.Get().
		Index("books").
		Type("book").
		Id(id).
		Do(ctx)
	if err != nil {
		return nil, ConvertError(err)
	}

	return getResponse.Source, nil
}

func ConvertError(err error) error {
	if elasticErr, ok := err.(*elastic.Error); ok {
		return errors.NewHttpError(elasticErr.Status, err.Error(), err)
	}
	return err
}