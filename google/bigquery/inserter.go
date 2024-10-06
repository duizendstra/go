package bqmodels

import (
	"context"
	"fmt"
	"github.com/duizendstra/go/google/structuredlogger"

	"cloud.google.com/go/bigquery"
)

// BigQueryInserter is an interface for inserting records into BigQuery.
type BigQueryInserter interface {
	InsertRecords(ctx context.Context, datasetID string, tableID string, records []bigquery.ValueSaver) error
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
		return nil, fmt.Errorf("bigquery.NewClient for project %s: %w", projectID, err)
	}
	return &DefaultBigQueryInserter{Client: client}, nil
}

// InsertRecords inserts records into the specified BigQuery dataset and table.
func (d *DefaultBigQueryInserter) InsertRecords(ctx context.Context, logger *structuredlogger.StructuredLogger, datasetID string, tableID string, records []bigquery.ValueSaver) error {

	// Validate that records are not nil or empty
	if records == nil || len(records) == 0 {
		return fmt.Errorf("no records to insert")
	}

	// Define the table inserter for inserting data into BigQuery
	u := d.Client.Dataset(datasetID).Table(tableID).Inserter()

	// Insert records into BigQuery using the table inserter
	if err := u.Put(ctx, records); err != nil {
		// Log the error for better traceability
		logger.LogError(ctx, "Error inserting records into BigQuery", "error", err)
		// Check if the error is a BigQuery-specific error and provide detailed information
		if bigqueryErr, ok := err.(*bigquery.Error); ok {
			logger.LogError(ctx, "BigQuery error details", "reason", bigqueryErr.Reason, "location", bigqueryErr.Location, "message", bigqueryErr.Message)
			return fmt.Errorf("BigQuery error: reason: %v, location: %v, message: %v", bigqueryErr.Reason, bigqueryErr.Location, bigqueryErr.Message)
		}
		return fmt.Errorf("u.Put: %w", err)
	}

	return nil
}
