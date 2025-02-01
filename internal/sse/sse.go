package sse

import (
	"fmt"
	"net/http"
	"sync"
)

type SSEManager struct {
	clients map[chan string]bool
	mu      sync.Mutex
}

func NewSSEManager() *SSEManager {
	return &SSEManager{
		clients: make(map[chan string]bool),
	}
}

// ✅ クライアントを登録
func (s *SSEManager) AddClient(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
		return
	}

	messageChan := make(chan string)
	s.mu.Lock()
	s.clients[messageChan] = true
	s.mu.Unlock()

	defer func() {
		s.mu.Lock()
		delete(s.clients, messageChan)
		s.mu.Unlock()
		close(messageChan)
	}()

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	for msg := range messageChan {
		fmt.Fprintf(w, "data: %s\n\n", msg)
		flusher.Flush()
	}
}

// ✅ 全クライアントにメッセージ送信
func (s *SSEManager) Broadcast(message string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for client := range s.clients {
		client <- message
	}
}
