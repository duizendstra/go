package handlers

import (
	"context"
	"encoding/json"
	"fmt"

	cloudtasks "cloud.google.com/go/cloudtasks/apiv2"
	"cloud.google.com/go/cloudtasks/apiv2/cloudtaskspb"
)

func EnqueueTask(ctx context.Context, body *interface{}, url string, queueName string, taskSuffix string, serviceAccountEmail string) error {
	taskClient, err := cloudtasks.NewClient(ctx)
	if err != nil {
		return fmt.Errorf("failed to create client: %v", err)
	}
	defer taskClient.Close()

	requestBody, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("failed to serialize request: %v", err)
	}

	taskFullName := fmt.Sprintf("%s/tasks/%s-%04d", queueName, taskSuffix)

	httpRequest := &cloudtaskspb.HttpRequest{
		HttpMethod: cloudtaskspb.HttpMethod_POST,
		Url:        url,
		Body:       requestBody,
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
		AuthorizationHeader: &cloudtaskspb.HttpRequest_OidcToken{ // Corrected line
			OidcToken: &cloudtaskspb.OidcToken{
				ServiceAccountEmail: serviceAccountEmail,
				Audience:            url,
			},
		},
	}

	req := &cloudtaskspb.CreateTaskRequest{
		Parent: queueName,
		Task: &cloudtaskspb.Task{
			Name: taskFullName,
			MessageType: &cloudtaskspb.Task_HttpRequest{
				HttpRequest: httpRequest,
			},
		},
	}

	_, err = taskClient.CreateTask(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to create task: %v", err)
	}
	return nil
}
