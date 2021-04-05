package book_service

import (
	"book_service/pkg/api"
	//"book_service/pkg/errors"
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	errors "github.com/fiverr/go_errors"
)

type BookService struct {
	HostUrl string
}

func NewBookService(hostUrl string) *BookService {
	return &BookService{
		HostUrl: hostUrl,
	}
}

func (b *BookService) StoreStats() (*api.StoreStats, error) {
	url := fmt.Sprintf("%s/books/_search", b.HostUrl)
	var result struct{
		Aggregations struct{
			Dcount struct{
				Value int `json:"value"`
			} `json:"dcount"`
		} `json:"aggregations"`
		Hits struct{
			Total int `json:"total"`
		} `json:"hits"`
	}
	if err := SendRequest(http.MethodGet, url, StoreQuery, &result); err != nil {
		return nil, err
	}

	return &api.StoreStats{
		Count:  result.Hits.Total,
		Dcount: result.Aggregations.Dcount.Value,
	}, nil
}

func (b *BookService) SearchBooks(title string, authorName string, minPrice float32, maxPrice float32) (interface{}, error) {
	url := fmt.Sprintf("%s/books/_search", b.HostUrl)

	var result struct{
		Hits struct {
			Hits interface{} `json:"hits"`
		} `json:"hits"`
	}

	err := SendRequest(http.MethodGet, url, fmt.Sprintf(SearchBooksQuery, title, authorName, minPrice, maxPrice), &result)

	return result.Hits.Hits, err
}

func (b *BookService) DeleteBook(id string) error {
	url := fmt.Sprintf("%s/books/book/%s", b.HostUrl, id)
	var x interface{}
	err := SendRequest(http.MethodDelete, url, nil, x)
	return err
}

func (b *BookService) UpdateBookTitle(id string, title string) error {
	url := fmt.Sprintf("%s/books/book/%s/_update", b.HostUrl, id)
	body := fmt.Sprintf(UpdateBookTitleQuery, title)

	resp := &ElasticSearchPostResponse{}
	return SendRequest(http.MethodPost, url, body, resp)
}

func (b *BookService) AddBook(book *api.Book) (id string, err error) {
	url := fmt.Sprintf("%s/books/book/", b.HostUrl)
	resp := &ElasticSearchPostResponse{}
	if err := SendRequest(http.MethodPost, url, book, resp); err != nil {
		return "", err
	}
	return resp.Id, nil
}

func (b *BookService) GetBook(id string) (interface{}, error) {
	url := fmt.Sprintf("%s/books/book/%s", b.HostUrl, id)

	resp := &ElasticSearchResponse{}
	if err := SendRequest(http.MethodGet, url, nil, resp); err != nil {
		return nil, err
	}

	return resp.Source, nil
}

func SendRequest(method string, url string, body interface{}, responseObj interface{}) error {
	// initialize http client
	client := &http.Client{}

	// marshal to json
	var bodyBytes []byte
	var err error
	if bodyString, ok := body.(string); ok {
		bodyBytes = []byte(bodyString)
	} else {
		bodyBytes, err = json.Marshal(body)
		if err != nil {
			return err
		}
	}

	// set the HTTP method, url, and request body
	req, err := http.NewRequest(method, url, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return err
	}

	// set the request header Content-Type for json
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	if resp.StatusCode >= 300 || resp.StatusCode < 200 {
		return errors.NewHttpError(resp.StatusCode, resp.Status, nil)
	}

	//bodyBytes, err = ioutil.ReadAll(resp.Body)
	//if err != nil {
	//	log.Fatal(err)
	//}
	//bodyString := string(bodyBytes)
	//print(bodyString)

	if err := json.NewDecoder(resp.Body).Decode(responseObj); err != nil {
		return err
	}

	return nil
}

const (
	StoreQuery = `
{
  "size": 0,
  "aggs": {
    "dcount": {
      "cardinality": {
        "field": "author's name.keyword"
      }
    }
  }
}`

	SearchBooksQuery = `
{
 "query": {
   "bool": {
     "must": [
       { "match_phrase": {"title": "%s"}},
       { "match_phrase": {"author's name": "%s"}}
     ],
     "filter": [
       { 
         "range": {
           "price": {
             "gte": %f,
             "lte": %f
           }
         }
       }
     ]
   }
 }
}`

	UpdateBookTitleQuery = `
{
 "doc": {
   "title": "%s"
 }
}
`
)
