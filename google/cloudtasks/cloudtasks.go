// Copyright 2024 Jasper Duizendstra
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package cloudtasks

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"

	cloudtasks "cloud.google.com/go/cloudtasks/apiv2"
	"cloud.google.com/go/cloudtasks/apiv2/cloudtaskspb"
)

// createCloudTasksClient creates a new Cloud Tasks client.
func createCloudTasksClient(ctx context.Context) (*cloudtasks.Client, error) {
	return cloudtasks.NewClient(ctx)
}

// createHttpRequest creates an HTTP request for a Cloud Task.
func createHttpRequest(body []byte, url, serviceAccountEmail string) *cloudtaskspb.HttpRequest {
	return &cloudtaskspb.HttpRequest{
		HttpMethod: cloudtaskspb.HttpMethod_POST,
		Url:        url,
		Body:       body,
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
		AuthorizationHeader: &cloudtaskspb.HttpRequest_OidcToken{
			OidcToken: &cloudtaskspb.OidcToken{
				ServiceAccountEmail: serviceAccountEmail,
				Audience:            url,
			},
		},
	}
}

// EnqueueTask enqueues a task to a specified Cloud Tasks queue with the given parameters.
func EnqueueTask(ctx context.Context, body interface{}, url string, queueName string, taskName string, serviceAccountEmail string) error {
	// Create a Cloud Tasks client
	taskClient, err := createCloudTasksClient(ctx)
	if err != nil {
		return fmt.Errorf("failed to create client: %v", err)
	}
	defer func() {
		if closeErr := taskClient.Close(); closeErr != nil {
			fmt.Printf("failed to close task client: %v\n", closeErr)
		}
	}()

	// Serialize the task payload
	requestBody, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("failed to serialize request: %v", err)
	}

	// Validate task name format
	taskNamePattern := `^[a-zA-Z0-9_-]{1,500}$`
	if matched, _ := regexp.MatchString(taskNamePattern, taskName); !matched {
		return fmt.Errorf("task name '%s' does not meet Cloud Tasks naming conventions", taskName)
	}

	// Define the full task name
	taskFullName := fmt.Sprintf("%s/tasks/%s", queueName, taskName)

	// Create the HTTP request for the task
	httpRequest := createHttpRequest(requestBody, url, serviceAccountEmail)

	// Create the task request
	req := &cloudtaskspb.CreateTaskRequest{
		Parent: queueName,
		Task: &cloudtaskspb.Task{
			Name: taskFullName,
			MessageType: &cloudtaskspb.Task_HttpRequest{
				HttpRequest: httpRequest,
			},
		},
	}

	// Enqueue the task
	_, err = taskClient.CreateTask(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to create task: %v", err)
	}
	return nil
}