package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/anmho/agent-control-api/internal/api"
	"github.com/anmho/agent-control-api/internal/dispatch"
	"github.com/anmho/agent-control-api/internal/runner"
	"github.com/anmho/agent-control-api/internal/store"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

func main() {
	log, _ := zap.NewProduction()
	defer log.Sync()

	ctx := context.Background()

	dbURL := mustEnv("DATABASE_URL")
	db, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		log.Fatal("connect to postgres", zap.Error(err))
	}
	defer db.Close()

	runStore := store.NewRunStore(db)
	if err := runStore.Migrate(ctx); err != nil {
		log.Fatal("migrate", zap.Error(err))
	}

	// Cloud Run runner: used in production when CLOUD_RUN_JOB_NAME is set.
	// Falls back to local runtime execution for local dev.
	var cloudRunner *runner.CloudRunClient
	if jobName := os.Getenv("CLOUD_RUN_JOB_NAME"); jobName != "" {
		cloudRunner, err = runner.NewCloudRunClient(ctx, jobName)
		if err != nil {
			log.Fatal("init cloud run client", zap.Error(err))
		}
		log.Info("cloud run mode", zap.String("job", jobName))
	} else {
		log.Info("local mode — runs will be executed via local runtime providers")
	}

	bus := runner.NewBus()
	localRunner := runner.NewLocalRunner(bus, runStore.AppendOutput)

	h := api.NewHandler(log, runStore, cloudRunner, localRunner, bus)
	dispatcher := dispatch.NewFromEnv(log.Named("mailbox-dispatcher"), h.DeliverMailboxMessage)
	if dispatcher != nil {
		dispatcher.Start(ctx)
		log.Info("mailbox dispatcher enabled", zap.String("events_url", os.Getenv("MCP_MAILBOX_EVENTS_URL")))
	} else {
		log.Info("mailbox dispatcher disabled")
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%s", port),
		Handler:      h.Routes(),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 15 * time.Minute,
		IdleTimeout:  60 * time.Second,
	}

	log.Info("agent-control-api listening", zap.String("port", port))
	if err := srv.ListenAndServe(); err != nil {
		log.Fatal("listen", zap.Error(err))
	}
}

func mustEnv(key string) string {
	v := os.Getenv(key)
	if v == "" {
		fmt.Fprintf(os.Stderr, "fatal: %s env var is required\n", key)
		os.Exit(1)
	}
	return v
}
