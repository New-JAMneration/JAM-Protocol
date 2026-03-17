package rpc

import (
	"context"
	"fmt"
	"net/http"

	"github.com/New-JAMneration/JAM-Protocol/logger"
	"github.com/gorilla/websocket"
)

type RPCServer struct {
	addr     string
	upgrader websocket.Upgrader
	mux      *http.ServeMux
}

func NewRPCServer(addr string) *RPCServer {
	return &RPCServer{
		addr: addr,
		mux:  http.NewServeMux(),
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true // Allow all origins now, TODO: implement origin checking
			},
		},
	}
}

func (s *RPCServer) Start(ctx context.Context) error {
	s.mux.HandleFunc("/", s.handleWebSocket)
	server := &http.Server{
		Addr:    s.addr,
		Handler: s.mux,
	}

	// Graceful shutdown
	go func() {
		<-ctx.Done()
		logger.Info("RPC server shutting down...")
		server.Shutdown(context.Background())
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

	handler := NewHandler()
	subManager := NewSubscriptionManager(conn)
	handler.SetSubscriptionManager(subManager)

	defer subManager.UnsubscribeAll()

	for {
		messageType, message, err := conn.ReadMessage()
		if err != nil {
			logger.Debug(fmt.Sprintf("Connection closed: %v", err))
			break
		}

		response := handler.HandleMessage(message)

		err = conn.WriteMessage(messageType, response)
		if err != nil {
			logger.Error(fmt.Sprintf("Write message error: %v", err))
			break
		}
	}

	logger.Info(fmt.Sprintf("Client disconnected: %s", conn.RemoteAddr()))
}
