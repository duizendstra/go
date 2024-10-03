package bqmodels

import (
	"context"
	"fmt"

	"cloud.google.com/go/bigquery"
)

type DefaultBigQueryInserter struct{}

func (d *DefaultBigQueryInserter) InsertRecords(ctx context.Context, projectID string, datasetID string, tableID string, records interface{}) error {
	client, err := bigquery.NewClient(ctx, projectID)
	if err != nil {
		return fmt.Errorf("bigquery.NewClient: %w", err)
	}
	defer client.Close()

	// Define the table inserter
	u := client.Dataset(datasetID).Table(tableID).Inserter()

	// Insert records into BigQuery
	if err := u.Put(ctx, records); err != nil {
		return fmt.Errorf("u.Put: %w", err)
	}

	return nil
}
