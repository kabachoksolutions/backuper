package main

import (
	"log"
)

func main() {
	w, err := NewWorker()
	if err != nil {
		log.Fatal(err)
	}
	if w.Config.RunScheduler {
		w.RunScheduler()
	}
	if err := w.Backup(); err != nil {
		log.Fatal(err)
	}
	log.Println("Job is done. Shutting down...")
}
