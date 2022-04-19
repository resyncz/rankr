package elastic

import (
	"context"
	"encoding/json"
	"fmt"
	elasticLib "github.com/olivere/elastic/v7"
	"os"
	"strings"
)

const (
	defaultSearchLimit = 50
	defaultHost        = "127.0.0.1"
	defaultPort        = "9200"
)

type Client struct {
	clientLib *elasticLib.Client
	conf      *Config
}

type Config struct {
	Host     string
	Port     string
	Sniff    bool
	UseHTTPS bool
}

type Entity struct {
	ID      string
	Content interface{}
}

type SearchEntity struct {
	Term      string
	Fields    []string
	From      int
	Limit     int
	Sort      string
	SortOrder bool // true -> asc, false -> desc
}

func NewConfig() *Config {
	host := defaultHost
	hostEnv := os.Getenv("ELASTIC_HOST")
	if strings.TrimSpace(hostEnv) != "" {
		host = hostEnv
	}

	port := defaultPort
	portEnv := os.Getenv("ELASTIC_PORT")
	if strings.TrimSpace(portEnv) != "" {
		port = portEnv
	}

	return &Config{
		Host: host,
		Port: port,
	}
}

func NewClient(conf *Config) (*Client, error) {
	link := "http"

	if conf.UseHTTPS {
		link = "https"
	}

	clientLib, err := elasticLib.NewClient(
		elasticLib.SetURL(fmt.Sprintf("%s://%s:%s", link, conf.Host, conf.Port)),
		elasticLib.SetSniff(conf.Sniff),
	)

	return &Client{
		conf:      conf,
		clientLib: clientLib,
	}, err
}

func (client *Client) Insert(ctx context.Context, index string, entity *Entity) error {
	svc := client.clientLib.Index().Index(index)

	_, err := svc.Id(entity.ID).BodyJson(entity.Content).Do(ctx)

	return err
}

func (client *Client) BulkInsert(ctx context.Context, index string, entities ...*Entity) error {
	bulk := client.clientLib.Bulk().Index(index)

	for _, entity := range entities {
		bulk.Add(elasticLib.NewBulkIndexRequest().Id(entity.ID).Doc(entity.Content))
	}

	_, err := bulk.Do(ctx)

	return err
}

func (client *Client) SearchByTerm(ctx context.Context, index string, searchEntity *SearchEntity) ([]byte, error) {
	if searchEntity.Limit == 0 {
		searchEntity.Limit = defaultSearchLimit
	}

	searchService := client.clientLib.Search().
		Index(index).
		From(searchEntity.From).
		Size(searchEntity.Limit)

	if strings.TrimSpace(searchEntity.Term) == "" {
		searchService.Query(elasticLib.NewMatchAllQuery())
	} else {
		searchService.Query(elasticLib.NewMultiMatchQuery(searchEntity.Term, searchEntity.Fields...).Type("phrase_prefix"))
	}

	if searchEntity.Sort != "" {
		searchService.Sort(searchEntity.Sort+".keyword", searchEntity.SortOrder)
	}

	searchResult, err := searchService.Do(ctx)
	if err != nil {
		return nil, err
	}

	return client.buildResponse(searchResult.Hits.Hits)
}

func (client *Client) buildResponse(hits []*elasticLib.SearchHit) ([]byte, error) {
	var resp []interface{}

	for _, hit := range hits {
		var item interface{}

		if err := json.Unmarshal(*&hit.Source, &item); err != nil {
			continue
		}

		resp = append(resp, item)
	}

	b, err := json.Marshal(resp)

	return b, err
}
