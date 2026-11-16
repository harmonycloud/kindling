// Copyright (C) 2022 Kindling Project Authors.
// SPDX-License-Identifier: Apache-2.0

package esclient

import (
	"context"
	"fmt"
	"github.com/olivere/elastic/v6"
	"log"
	"os"
	"time"
)

type EsClient struct {
	client        *elastic.Client
	bulkProcessor *elastic.BulkProcessor
}

func NewEsClient(urls ...string) (*EsClient, error) {
	if len(urls) == 0 {
		return nil, fmt.Errorf("the url of elasticsearch is not provided")
	}
	errorLog := log.New(os.Stdout, "esclient", log.LstdFlags)
	esClient, err := elastic.NewClient(elastic.SetErrorLog(errorLog), elastic.SetURL(urls...), elastic.SetSniff(false))
	if err != nil {
		return nil, err
	}
	info, code, err := esClient.Ping(urls[0]).Do(context.Background())
	if err != nil {
		return nil, err
	}
	log.Printf("Elasticsearch returns with code %d and version %s", code, info.Version.Number)

	bulkProcessor, err := esClient.BulkProcessor().FlushInterval(3 * time.Second).After(
		func(executionId int64, requests []elastic.BulkableRequest, response *elastic.BulkResponse, err error) {
			if err != nil {
				log.Printf("After bulkProcessor with errors: %v", err)
			}
		}).Do(context.Background())
	if err != nil {
		return nil, err
	}
	ret := &EsClient{
		client:        esClient,
		bulkProcessor: bulkProcessor,
	}
	return ret, nil
}

func (e *EsClient) IndexJson(index string, json interface{}) error {
	_, err := e.client.Index().Index(index).Type("_doc").BodyJson(json).Do(context.Background())
	if err != nil {
		return err
	}
	return nil
}

func (e *EsClient) AddIndexRequestWithParams(index string, doc interface{}) {
	request := elastic.NewBulkIndexRequest().Index(index).Type("_doc").Doc(doc)
	e.bulkProcessor.Add(request)
}
