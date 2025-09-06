package queue

import (
	"fmt"
	"time"

	"github.com/ramil063/secondgodiplom/cmd/client/handlers/grpc"
	bankcardQueue "github.com/ramil063/secondgodiplom/cmd/client/handlers/queue/bankcard"
	binaryQueue "github.com/ramil063/secondgodiplom/cmd/client/handlers/queue/binary"
	passwordQueue "github.com/ramil063/secondgodiplom/cmd/client/handlers/queue/password"
	textdataQueue "github.com/ramil063/secondgodiplom/cmd/client/handlers/queue/textdata"
	"github.com/ramil063/secondgodiplom/internal/constants/queue"
)

type Sender struct {
	clients  grpc.Clients
	interval time.Duration
	stopChan chan struct{}
}

func NewSender(clients grpc.Clients) *Sender {
	return &Sender{
		clients:  clients,
		interval: queue.SendTimeInterval,
		stopChan: make(chan struct{}),
	}
}

func (s *Sender) Start() {
	passwordQueue.Init()
	textdataQueue.Init()
	bankcardQueue.Init()
	binaryQueue.Init()
	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()

	fmt.Println("Запуск сервиса отправки очереди...")

	for {
		select {
		case <-ticker.C:
			go passwordQueue.Process(s.clients.PasswordsClient)
			go textdataQueue.Process(s.clients.TextDataClient)
			go bankcardQueue.Process(s.clients.BankCardDataClient)
			go binaryQueue.Process(s.clients.BinaryDataClient)
		case <-s.stopChan:
			fmt.Println("Остановка сервиса отправки очереди")
			return
		}
	}
}

func (s *Sender) Stop() {
	close(s.stopChan)
}
