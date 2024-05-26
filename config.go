package main

import (
	"fmt"
	"os"

	"github.com/go-playground/validator/v10"
	"github.com/joho/godotenv"
)

type AWSCredentials struct {
	AccessKeyID     string `validate:"required"`
	SecretAccessKey string `validate:"required"`
}

type S3Config struct {
	Bucket   string `validate:"required"`
	Region   string `validate:"required"`
	Endpoint string `validate:"omitempty"`
}

type BackupConfig struct {
	DatabaseURL  string `validate:"required"`
	CronSchedule string `validate:"omitempty,cron"`
	CronTimezone string `validate:"omitempty"`
}

type EnvConfig struct {
	AWS          AWSCredentials `validate:"required"`
	S3           S3Config       `validate:"required"`
	Backup       BackupConfig   `validate:"required"`
	RunOnStartup bool           `validate:"omitempty"`
	RunScheduler bool           `validate:"omitempty"`
}

func LoadEnv() (*EnvConfig, error) {
	environment := os.Getenv("ENVIRONMENT")
	if environment != "production" {
		if err := godotenv.Load(); err != nil {
			return nil, fmt.Errorf("failed to load .env file: %w", err)
		}
	}
	config := EnvConfig{
		AWS: AWSCredentials{
			AccessKeyID:     os.Getenv("AWS_ACCESS_KEY_ID"),
			SecretAccessKey: os.Getenv("AWS_SECRET_ACCESS_KEY"),
		},
		S3: S3Config{
			Bucket:   os.Getenv("AWS_S3_BUCKET"),
			Region:   os.Getenv("AWS_S3_REGION"),
			Endpoint: os.Getenv("AWS_S3_ENDPOINT"),
		},
		Backup: BackupConfig{
			DatabaseURL:  os.Getenv("BACKUP_DATABASE_URL"),
			CronSchedule: os.Getenv("BACKUP_CRON_SCHEDULE"),
			CronTimezone: os.Getenv("BACKUP_CRON_TIMEZONE"),
		},
		RunOnStartup: os.Getenv("RUN_ON_STARTUP") == "true",
		RunScheduler: os.Getenv("RUN_SCHEDULER") == "true",
	}
	validate := validator.New()
	if err := validate.Struct(config); err != nil {
		return nil, fmt.Errorf("invalid environment variables: %w", err)
	}
	return &config, nil
}
