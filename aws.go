package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

type AWSSession struct {
	svc *s3.S3
}

func NewAWSSession(config *EnvConfig) (*AWSSession, error) {
	sess, err := session.NewSession(&aws.Config{
		Region:   aws.String(config.S3.Region),
		Endpoint: aws.String(config.S3.Endpoint),
		Credentials: credentials.NewStaticCredentials(
			config.AWS.AccessKeyID, config.AWS.SecretAccessKey, "",
		),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to start new aws session: %w", err)
	}
	svc := s3.New(sess)
	return &AWSSession{svc: svc}, nil
}

func (s *AWSSession) UploadToS3(filename, filepath string, bucket string) error {
	log.Println("Uploading backup to S3...")
	file, err := os.Open(filepath)
	if err != nil {
		return fmt.Errorf("failed to open a file: %w", err)
	}
	defer file.Close()
	if _, err = s.svc.PutObjectWithContext(context.TODO(), &s3.PutObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(filename),
		Body:   file,
	}); err != nil {
		return fmt.Errorf("failed to put object to S3 storage: %w", err)
	}
	log.Println("Backup uploaded to S3...")
	return nil
}
