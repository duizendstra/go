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

package cloudtasks_test

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"cloud.google.com/go/cloudtasks/apiv2/cloudtaskspb"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type mockCloudTasksClient struct{}

func (m *mockCloudTasksClient) CreateTask(ctx context.Context, req *cloudtaskspb.CreateTaskRequest) (*cloudtaskspb.Task, error) {
	return &cloudtaskspb.Task{
		Name:            req.Task.GetName(),
		CreateTime:      timestamppb.Now(),
		ScheduleTime:    timestamppb.Now(),
		DispatchDeadline: durationpb.New(10 * time.Minute),
	}, nil
}

func EnqueueTaskWithMockClient(ctx context.Context, client *mockCloudTasksClient, body interface{}, url, queueName, taskName, serviceAccountEmail string) error {
	requestBody, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("failed to serialize request: %v", err)
	}

	httpRequest := &cloudtaskspb.HttpRequest{
		HttpMethod: cloudtaskspb.HttpMethod_POST,
		Url:        url,
		Body:       requestBody,
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

	req := &cloudtaskspb.CreateTaskRequest{
		Parent: queueName,
		Task: &cloudtaskspb.Task{
			Name: taskName,
			MessageType: &cloudtaskspb.Task_HttpRequest{
				HttpRequest: httpRequest,
			},
		},
	}

	_, err = client.CreateTask(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to create task: %v", err)
	}
	return nil
}

func TestEnqueueTask(t *testing.T) {
	mockClient := &mockCloudTasksClient{}
	body := map[string]interface{}{
		"key": "value",
	}

	taskQueue := "projects/test-project/locations/us-central1/queues/test-queue"
	taskName := "test-task"
	targetUrl := "https://example.com/task"
	serviceAccount := "test-service-account@example.com"

	err := EnqueueTaskWithMockClient(context.Background(), mockClient, body, targetUrl, taskQueue, taskName, serviceAccount)
	assert.NoError(t, err, "expected no error when enqueuing a task, but got: %v", err)
}