package main

import (
	"com.github.rverst.wp-ics-gen/app/worker"
	"log"

	"com.github.rverst.wp-ics-gen/app/config"
	"com.github.rverst.wp-ics-gen/app/server"
)

func main() {
	cfg := config.Get()
	wrk := worker.New(cfg.WorkingDir, cfg.CheckInterval, cfg.EventsUrl)
	wrk.StartWorker()

	srv := server.NewServer(cfg.Server.Address, cfg.BaseUrl)
	go func() {
		for {
			select {
			case content := <-wrk.Update:
				srv.UpdateContent(content)
			}
		}
	}()

	if err := srv.Start(); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
