package bqmodels

import (
	"context"
	"fmt"

	"cloud.google.com/go/bigquery"
)

// BigQueryInserter is an interface for inserting records into BigQuery.
type BigQueryInserter interface {
	InsertRecords(ctx context.Context, datasetID string, tableID string, records []bigquery.ValueSaver) error
	Close() error
}

// DefaultBigQueryInserter is a struct that manages the BigQuery client and handles data insertion.
type DefaultBigQueryInserter struct {
	Client *bigquery.Client // BigQuery client for interacting with BigQuery API
}

// NewDefaultBigQueryInserter initializes a new DefaultBigQueryInserter with a BigQuery client.
func NewDefaultBigQueryInserter(ctx context.Context, projectID string) (*DefaultBigQueryInserter, error) {
	// Create a new BigQuery client using the provided project ID
	client, err := bigquery.NewClient(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("bigquery.NewClient: %w", err)
	}
	return &DefaultBigQueryInserter{Client: client}, nil
}

// Close releases any resources held by the BigQuery client.
func (d *DefaultBigQueryInserter) Close() error {
	// Close the BigQuery client to free up resources
	return d.Client.Close()
}

// InsertRecords inserts records into the specified BigQuery dataset and table.
func (d *DefaultBigQueryInserter) InsertRecords(ctx context.Context, datasetID string, tableID string, records []bigquery.ValueSaver) error {
	// Define the table inserter for inserting data into BigQuery
	u := d.Client.Dataset(datasetID).Table(tableID).Inserter()

	// Insert records into BigQuery using the table inserter
	if err := u.Put(ctx, records); err != nil {
		// Check if the error is a BigQuery-specific error and provide detailed information
		if bigqueryErr, ok := err.(*bigquery.Error); ok {
			return fmt.Errorf("BigQuery error: reason: %v, location: %v, message: %v", bigqueryErr.Reason, bigqueryErr.Location, bigqueryErr.Message)
		}
		return fmt.Errorf("u.Put: %w", err)
	}

	return nil
}