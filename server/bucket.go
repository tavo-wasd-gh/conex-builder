package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

func UploadFile(s3Client *s3.Client, endpoint string, bucketName string,
	publicEndpoint string, fileContent []byte,
	objectKey string) (string, error) {
	if _, err := s3Client.PutObject(context.TODO(), &s3.PutObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(objectKey),
		Body:   bytes.NewReader(fileContent),
	}); err != nil {
		return "", fmt.Errorf("unable to upload file: %w", err)
	}

	return fmt.Sprintf("%s/%s", publicEndpoint, objectKey), nil
}

func BucketSizeLimit(apiEndpoint string, apiToken string) error {
	req, err := http.NewRequest("GET", apiEndpoint, nil)
	req.Header.Set("Authorization", "Bearer "+apiToken)
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("unable to check bucket size 1: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("unable to check bucket size 2: %w", err)
	}

	var bucket struct {
		Result struct {
			PayloadSize string `json:"payloadSize"`
		} `json:"result"`
	}

	err = json.Unmarshal(body, &bucket)
	if err != nil {
		return fmt.Errorf("unable to check bucket size 3: %w", err)
	}

	payloadBytes, err := strconv.Atoi(bucket.Result.PayloadSize)
	if err != nil {
		return fmt.Errorf("unable to check bucket size 4: %w", err)
	}

	if payloadBytes > maxBucketSize {
		return fmt.Errorf("unable to check bucket size 5: %w", err)
	}

	return nil
}
