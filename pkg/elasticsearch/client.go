package elasticsearch

import (
	"bytes"
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/elastic/go-elasticsearch/v7"
	"github.com/elastic/go-elasticsearch/v7/esutil"

	"bugfender-integration-elasticsearch/pkg/integration"
)

type Client struct {
	es          *elasticsearch.Client
	indexer     esutil.BulkIndexer
	failureFunc func(context.Context, esutil.BulkIndexerItem, esutil.BulkIndexerResponseItem, error) // Per item
}

// NewClient creates an ES client with the given parameters
// It is compulsory to call Close when done.
func NewClient(index string, addresses []string, username, password string) (*Client, error) {
	// uses Elasticsearch's BulkIndexer utility
	// example: https://github.com/elastic/go-elasticsearch/blob/v7.10.0/esutil/bulk_indexer_example_test.go
	es, err := elasticsearch.NewClient(elasticsearch.Config{
		Addresses:     addresses,
		Username:      username,
		Password:      password,
		RetryOnStatus: []int{502, 503, 504, 429},
		RetryBackoff:  func(i int) time.Duration { return time.Duration(i) * 100 * time.Millisecond },
		MaxRetries:    5,
	})
	if err != nil {
		return nil, err
	}
	indexer, err := esutil.NewBulkIndexer(esutil.BulkIndexerConfig{
		Client: es,
		Index:  index,
	})
	if err != nil {
		log.Fatalf("Error creating the indexer: %s", err)
	}
	failureFunc := func(
		ctx context.Context,
		item esutil.BulkIndexerItem,
		res esutil.BulkIndexerResponseItem, err error,
	) {
		if err != nil {
			log.Printf("ERROR: %s", err)
		} else {
			log.Printf("ERROR: %s: %s", res.Error.Type, res.Error.Reason)
		}
	}
	return &Client{
		es:          es,
		indexer:     indexer,
		failureFunc: failureFunc,
	}, nil
}

// WriteLogs writes logs to Elasticsearch
func (ec *Client) WriteLogs(ctx context.Context, page []integration.Log) error {
	for _, l := range page {
		doc, err := json.Marshal(l)
		if err != nil {
			panic(err)
		}
		err = ec.indexer.Add(
			ctx,
			esutil.BulkIndexerItem{
				Action:     "index",
				DocumentID: l.Uuid.String(),
				Body:       bytes.NewReader(doc),
				OnFailure:  ec.failureFunc,
			},
		)
		if err != nil {
			return err
		}
	}
	return nil
}

// Close flushes any pending logs and frees resources
func (ec *Client) Close(ctx context.Context) error {
	return ec.indexer.Close(ctx)
}
