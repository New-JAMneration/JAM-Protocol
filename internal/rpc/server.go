package rpc

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/New-JAMneration/JAM-Protocol/logger"
	"github.com/gorilla/websocket"
)

type RPCServer struct {
	addr         string
	upgrader     websocket.Upgrader
	mux          *http.ServeMux
	chainWatcher *ChainWatcher
	chainReader  ChainReader
	eventSub     EventSubscriber
}

func NewRPCServer(addr string, chainReader ChainReader, publisher EventPublisher, subscriber EventSubscriber) *RPCServer {
	return &RPCServer{
		addr: addr,
		mux:  http.NewServeMux(),
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true // Allow all origins now, TODO: implement origin checking
			},
		},
		chainWatcher: NewChainWatcher(chainReader, publisher),
		chainReader:  chainReader,
		eventSub:     subscriber,
	}
}

func (s *RPCServer) Start(ctx context.Context) error {
	s.mux.HandleFunc("/", s.handleWebSocket)
	server := &http.Server{
		Addr:    s.addr,
		Handler: s.mux,
	}

	// Start ChainWatcher to publish block/sync events for subscriptions
	s.chainWatcher.Start()

	// Graceful shutdown
	go func() {
		<-ctx.Done()
		logger.Info("RPC server shutting down...")
		s.chainWatcher.Stop()
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer shutdownCancel()
		server.Shutdown(shutdownCtx)
	}()
	logger.Infof("RPC server starting on %s", s.addr)
	return server.ListenAndServe()
}

func (s *RPCServer) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		logger.Errorf("WebSocket upgrade error: %v", err)
		return
	}
	defer conn.Close()

	logger.Info(fmt.Sprintf("Client connected: %s", conn.RemoteAddr()))

	var writeMu sync.Mutex
	subManager := NewSubscriptionManager(conn, &writeMu, s.eventSub)
	handler := NewHandler(NewRPCService(s.chainReader), subManager)

	defer subManager.UnsubscribeAll()

	for {
		messageType, message, err := conn.ReadMessage()
		if err != nil {
			logger.Debug(fmt.Sprintf("Connection closed: %v", err))
			break
		}

		response := handler.HandleMessage(message)

		writeMu.Lock()
		err = conn.WriteMessage(messageType, response)
		writeMu.Unlock()
		if err != nil {
			logger.Error(fmt.Sprintf("Write message error: %v", err))
			break
		}
	}

	logger.Info(fmt.Sprintf("Client disconnected: %s", conn.RemoteAddr()))
}
