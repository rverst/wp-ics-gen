package main

import (
	"log/slog"
	"os"

	"com.github.rverst.wp-ics-gen/app/config"
	"com.github.rverst.wp-ics-gen/app/server"
	"com.github.rverst.wp-ics-gen/app/worker"
)

func main() {
	cfg := config.Get()
	wrk := worker.New(cfg.WorkingDir, cfg.CheckInterval, cfg.EventsUrl)
	wrk.StartWorker()

	slog.Info("Starting server", "address", cfg.Server.Address)
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
		slog.Error("Server failed", "error", err)
		os.Exit(1)
	}
}
