package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/ramil063/secondgodiplom/cmd/gophkeeper/server"
	"github.com/ramil063/secondgodiplom/internal/logger"
)

func main() {
	// регистрируем перенаправление прерываний
	ctxGrSh, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT)
	defer stop()

	config, grpcStorage, manager, err := server.PrepareServerEnvironment()
	if err != nil {
		logger.WriteErrorLog(err.Error())
	}

	grpcServer, lis, err := server.GetGRPCServer(config)
	if err != nil {
		logger.WriteErrorLog(err.Error())
	}

	server.RegisterServiceServers(grpcServer, grpcStorage, config, manager)

	// через этот канал сообщим основному потоку, что соединения закрыты
	idleConnectsClosed := make(chan struct{})

	serverErr := make(chan error, 1)
	go func() {
		fmt.Println("Server gRPC started")
		if err = grpcServer.Serve(lis); err != nil {
			log.Printf("gRPC server Serve error: %v", err)
		}
	}()

	select {
	case <-ctxGrSh.Done():
		log.Println("Starting graceful shutdown...")
		grpcServer.GracefulStop()
		// сообщаем основному потоку,
		// что все сетевые соединения обработаны и закрыты
		close(idleConnectsClosed)
		log.Println("All connections closed")
	case err = <-serverErr:
		log.Printf("Server error: %v", err)
	}

	// ждём завершения процедуры graceful shutdown
	<-idleConnectsClosed
	grpcStorage.GetRepository().Pool.Close()
	fmt.Println("Server Shutdown gracefully")
}
