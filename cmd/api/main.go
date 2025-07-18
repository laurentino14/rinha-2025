package main

import (
	"context"
	"log"
	"net"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/laurentino14/rinha-2025/internal/api"
	"github.com/laurentino14/rinha-2025/internal/application/payment"
	"github.com/laurentino14/rinha-2025/internal/config"

	"github.com/laurentino14/rinha-2025/internal/infra/pg"

	"github.com/valyala/fasthttp"
)

func main() {
	appCtx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	var err error
	var pgPool *pgxpool.Pool
	var ln net.Listener
	pollWorkerSize, _ := strconv.Atoi(config.GetDefaultEnv("WORKERS_POOL_SIZE", "10"))

	if pgPool, err = pg.NewPGPool(appCtx); err != nil {
		log.Fatalf("Error connecting to PostgreSQL: %v", err)
	}
	defer pgPool.Close()

	cache := payment.NewCache()
	repository := pg.NewPaymentRepository(pgPool)
	healthWorker := payment.NewHealthCheckWorker(appCtx, cache)
	processWorkerPool := payment.NewProcessWorkerPool(repository, cache)

	healthWorker.Run()
	processWorkerPool.Run(pollWorkerSize)

	if ln, err = net.Listen("tcp4", config.ADDR); err != nil {
		log.Fatalf("Failed to start tcp listener: %v", err)
	}

	s := api.Setup(repository, processWorkerPool)

	serverIsDown := make(chan struct{})
	// for range 4 {
	go func() {
		log.Printf("Server listening on %s", config.ADDR)
		if err := s.Serve(ln); err != nil && err != fasthttp.ErrConnectionClosed {
			log.Fatalf("Failed to start reuseport: %v", err)
		}
		close(serverIsDown)
	}()
	// }

	defer ln.Close()

	<-appCtx.Done()
	log.Println("Byyyye")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), time.Second*15)
	defer shutdownCancel()

	if err := s.ShutdownWithContext(shutdownCtx); err != nil {
		log.Printf("Error at shutdown: %v", err)
	}
	<-serverIsDown
	log.Println("Server closed!")
}
