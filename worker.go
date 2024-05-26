package main

import (
	"compress/gzip"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path"
	"time"

	"github.com/go-co-op/gocron"
)

type Worker struct {
	Config *EnvConfig
	AWS    *AWSSession
}

func NewWorker() (*Worker, error) {
	config, err := LoadEnv()
	if err != nil {
		return nil, err
	}
	aws, err := NewAWSSession(config)
	if err != nil {
		return nil, err
	}
	return &Worker{Config: config, AWS: aws}, nil
}

func (w *Worker) RunScheduler() {
	schedulerTimezone := time.UTC
	if w.Config.Backup.CronTimezone != "" {
		customTimeZone, err := time.LoadLocation(w.Config.Backup.CronTimezone)
		if err != nil {
			log.Fatal(err)
		}
		schedulerTimezone = customTimeZone
	}
	s := gocron.NewScheduler(schedulerTimezone)
	log.Printf("Waiting for the cron...\n\n")
	s.Cron(w.Config.Backup.CronSchedule).Do(func() {
		log.Println("Cron is triggered.")
		_, nextScheduledTime := s.NextRun()
		if err := w.Backup(); err != nil {
			log.Println(err)
		}
		log.Printf("Next scheduled time: %s\n", nextScheduledTime)
	})
	if w.Config.RunOnStartup {
		log.Printf("RunOnStartup function is enabled.\n\n")
		if err := w.Backup(); err != nil {
			log.Println(err)
		}
		log.Printf("Waiting for the next cron schedule.\n\n")
	}
	s.StartBlocking()
}

func (w *Worker) Backup() error {
	log.Println("Initiating DB backup...")
	date := time.Now().Format("2006-01-02T15-04-05")
	filename := fmt.Sprintf("backup-%s.tar.gz", date)
	filepath := path.Join(os.TempDir(), filename)
	if err := w.DumpToFile(filepath); err != nil {
		return err
	}
	if err := w.AWS.UploadToS3(filename, filepath, w.Config.S3.Bucket); err != nil {
		return err
	}
	if err := w.DeleteFile(filepath); err != nil {
		return err
	}
	log.Println("DB backup complete...")
	return nil
}

func (w *Worker) DumpToFile(filepath string) error {
	log.Println("Dumping DB to file...")
	outputFile, err := os.Create(filepath)
	if err != nil {
		return fmt.Errorf("error creating a file: %w", err)
	}
	defer outputFile.Close()
	gzipWriter := gzip.NewWriter(outputFile)
	defer gzipWriter.Close()
	cmd := exec.Command("pg_dump", fmt.Sprintf("--dbname=%s", w.Config.Backup.DatabaseURL), "--format=tar")
	cmd.Stdout = gzipWriter
	cmd.Stderr = os.Stderr
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start pg_dump command: %w", err)
	}
	if err := cmd.Wait(); err != nil {
		return fmt.Errorf("failed to wait for pg_dump command: %w", err)
	}
	fileInfo, err := os.Stat(filepath)
	if err != nil {
		return fmt.Errorf("failed to get file size: %w", err)
	}
	fileSizeMB := float64(fileInfo.Size()) / (1024 * 1024)
	log.Printf("DB dumped to file: %s (Size: ~%.2f MB)\n", filepath, fileSizeMB)
	return nil
}

func (w *Worker) DeleteFile(filepath string) error {
	log.Println("Deleting file...")
	err := os.Remove(filepath)
	if err != nil {
		return fmt.Errorf("failed to delete file: %w", err)
	}
	return nil
}
