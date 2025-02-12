package handler

import (
	"fmt"
	"net/http"
	"strconv"
	"sync"

	"github.com/labstack/echo/v4"
	"github.com/yomek33/newln/internal/logger"
	"github.com/yomek33/newln/internal/services"
)

type ChatWSHandler struct {
	service     services.ChatService
	subscribers map[uint][]chan string
	mu          sync.Mutex
	jwtSecret   []byte
}

func NewChatWSHandler(service services.ChatService, jwtSecret []byte) *ChatWSHandler {
	return &ChatWSHandler{
		service:     service,
		subscribers: make(map[uint][]chan string),
		jwtSecret:   jwtSecret,
	}
}


func(h *ChatWSHandler)CreateFirstchat(c echo.Context) error {
	logger.Info("🔹 CreateFirstchat")
	//UserID, err := getUserIDFromContext(c)
	// if err != nil {
	// 	return respondWithError(c, http.StatusUnauthorized, ErrInvalidUserToken)
	// }

	chat, _, err := h.service.CreateFirstChat(1)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create chat"})
	}
	return c.JSON(http.StatusOK, chat)
}

var chatHistory = make(map[uint][]ChatMessage)
var mu sync.Mutex // 排他制御
type ChatRequest struct {
	ChatID  string `json:"chat_id"`
	Message string `json:"message"`
}

// **ChatMessage: チャットの 1 メッセージ**
type ChatMessage struct {
	Sender  string `json:"sender"`  // "AI" or "User"
	Message string `json:"message"` // 発言内容
}

// **ChatResponse: AI の応答**
type ChatResponse struct {
	ChatID  string        `json:"chat_id"`
	History []ChatMessage `json:"history"`
}
func (c *ChatWSHandler) fetchChatHistory(chatID uint) ([]ChatMessage, error) {
	mu.Lock()
	history, exists := chatHistory[chatID]
	mu.Unlock()

	if exists {
		return history, nil
	}

	_, his, err := c.service.CreateFirstChat(chatID)
	if err != nil {
		return nil, fmt.Errorf("Failed to create first chat: %w", err)
	}

	newHistory := []ChatMessage{}
	for _, h := range his {
		if len(h.Parts) == 0 {
			continue
		}
		newHistory = append(newHistory, ChatMessage{
			Sender:  h.Role,
			Message: fmt.Sprintf("%v", h.Parts[0]),
		})
	}

	// **履歴を保存**
	mu.Lock()
	chatHistory[chatID] = newHistory
	mu.Unlock()

	return newHistory, nil
}

// **AI にメッセージを送り、履歴を更新**
func (c *ChatWSHandler) sendMessageToAI(chatID uint, history []ChatMessage, message string) ([]ChatMessage, error) {
	// **ユーザーのメッセージを履歴に追加**
	history = append(history, ChatMessage{Sender: "User", Message: message})

	// **型変換: handlers.ChatMessage → services.ChatMessage**
	serviceHistory := make([]services.ChatMessage, len(history))
	for i, h := range history {
		serviceHistory[i] = services.ChatMessage{
			Sender:  h.Sender,
			Message: h.Message,
		}
	}

	// **AI に問い合わせ (service 経由)**
	updatedServiceHistory, err := c.service.ContinueChat(chatID, serviceHistory, message)
	if err != nil {
		return nil, fmt.Errorf("Failed to send message: %w", err)
	}

	// **型変換: services.ChatMessage → handlers.ChatMessage**
	updatedHistory := make([]ChatMessage, len(updatedServiceHistory))
	for i, h := range updatedServiceHistory {
		updatedHistory[i] = ChatMessage{
			Sender:  h.Sender,
			Message: h.Message,
		}
	}

	mu.Lock()
	chatHistory[chatID] = updatedHistory
	mu.Unlock()

	return updatedHistory, nil
}

// **POST /chat: チャット処理**
func (c *ChatWSHandler) chatHandler(ctx echo.Context) error {
	req := new(ChatRequest)
	// **リクエストのパース**
	if err := ctx.Bind(req); err != nil {
		return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request"})
	}

	// **chatID を uint に変換**
	chatID, err := strconv.ParseUint(req.ChatID, 10, 64)
	if err != nil {
		return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid chat_id"})
	}
	logger.Infof("🔹 ChatID: %d", chatID)
	// **履歴を取得**
	history, err := c.fetchChatHistory(uint(chatID))
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	// **メッセージを送信**
	updatedHistory, err := c.sendMessageToAI(uint(chatID), history, req.Message)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	// **レスポンスを返す**
	return ctx.JSON(http.StatusOK, ChatResponse{
		ChatID:  string(chatID),
		History: updatedHistory,
	})
}


// func (h *ChatWSHandler) ChatWebSocket(c echo.Context) error {
// 	logger.Info("🔹 ChatWebSocket")

// 	chatID, err := strconv.Atoi(c.Param("chatID"))
// 	if err != nil {
// 		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid chat ID"})
// 	}

// 	// ユーザー認証（UserID 取得）
// 	tokenString := c.QueryParam("token")
// 	if tokenString == "" {
// 		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "missing token"})
// 	}
// 	userID, err := isValidJWTToken(tokenString, h.jwtSecret)
// 	if err != nil {
// 		return c.JSON(http.StatusUnauthorized, map[string]string{"error": err.Error()})
// 	}
// 	userUUID, err := uuid.Parse(userID)
// 	if err != nil {
// 		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid user ID"})
// 	}

// 	// ✅ WebSocket に接続する前に SubscribeToChat する
// 	ch := h.SubscribeToChat(uint(chatID))
// 	defer h.UnsubscribeFromChat(uint(chatID), ch)

// 	// 🔍 チャットにメッセージがあるかチェック
// 	hasMessages, err := h.service.CheckChatHasMessages(uint(chatID))
// 	if err != nil {
// 		log.Println("❌ Failed to check chat messages:", err)
// 		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Chat check failed"})
// 	}

// 	// 🔥 メッセージがない場合、StartChat を実行
// 	if !hasMessages {
// 		newChat, err := h.service.StartChat(uint(chatID), &models.Chat{})
// 		if err != nil {
// 			log.Println("❌ Failed to start chat:", err)
// 			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to start chat"})
// 		}

// 		// WebSocket で AI の最初のメッセージを送信
// 		if len(newChat.Messages) > 0 {
// 			h.PublishChatUpdate(uint(chatID), "system" + newChat.Messages[0].Content)
// 			logger.Infof("🚀 Chat started: ChatID=%d, FirstMessage=%s", chatID, newChat.Messages[0].Content)
// 		}
// 	}

// 	// ✅ WebSocket ハンドラーを登録
// 	websocket.Handler(func(ws *websocket.Conn) {
// 		defer ws.Close()

// 		// ✅ WebSocket 送信ループ
// 		go func() {
// 			for msg := range ch {
// 				if err := websocket.Message.Send(ws, msg); err != nil {
// 					log.Println("❌ WebSocket send error:", err)
// 					break
// 				}
// 			}
// 		}()

// 		// ✅ WebSocket 受信ループ（ユーザーからのメッセージ受信）
// 		for {
// 			var message string
// 			if err := websocket.Message.Receive(ws, &message); err != nil {
// 				log.Printf("❌ WebSocket read error (ChatID=%d, UserID=%s): %v", chatID, userUUID, err)
// 				break
// 			}

// 			// 🔍 受信したメッセージをログに出力
// 			log.Printf("📥 Received message from UserID=%s in ChatID=%d: %s", userUUID, chatID, message)

// 			// メッセージを処理
// 			h.SendMessageWS(c, uint(chatID), userUUID, message)
// 		}
// 	}).ServeHTTP(c.Response(), c.Request())

// 	return nil
// }



// // 🔹 メッセージを WebSocket で送信（echo.Context を引数に追加）
// func (h *ChatWSHandler) SendMessageWS(c echo.Context, chatID uint, userID uuid.UUID, message string) {
// 	response, err := h.service.SendMessage(chatID, userID, message)
// 	if err != nil {
// 		log.Println("❌ Failed to send message:", err)
// 		return
// 	}

// 	// チャットが終了している場合、特別なメッセージを送る
// 	if response.Content == "🚀 チャット終了！次のステップへ進みます。（後で実装）" {
// 		h.PublishChatUpdate(chatID, "🚀 チャットが終了しました。")
// 		return
// 	}

// 	// クライアントにリアルタイム配信
// 	h.PublishChatUpdate(chatID, response.Content)
// }

// func (h *ChatWSHandler) SubscribeToChat(chatID uint) chan string {
// 	h.mu.Lock()
// 	defer h.mu.Unlock()

// 	ch := make(chan string, 10)
// 	h.subscribers[chatID] = append(h.subscribers[chatID], ch)

// 	return ch
// }

// func (h *ChatWSHandler) UnsubscribeFromChat(chatID uint, ch chan string) {
// 	h.mu.Lock()
// 	defer h.mu.Unlock()

// 	channels, exists := h.subscribers[chatID]
// 	if !exists {
// 		return
// 	}

// 	newChannels := make([]chan string, 0, len(channels))
// 	for _, c := range channels {
// 		if c != ch {
// 			newChannels = append(newChannels, c)
// 		}
// 	}

// 	if len(newChannels) == 0 {
// 		delete(h.subscribers, chatID)
// 	} else {
// 		h.subscribers[chatID] = newChannels
// 	}

// 	close(ch)
// }

// func (h *ChatWSHandler) PublishChatUpdate(chatID uint, message string) {
// 	h.mu.Lock()
// 	defer h.mu.Unlock()

// 	subscribers, ok := h.subscribers[chatID]
// 	if !ok {
// 		return
// 	}

// 	for _, ch := range subscribers {
// 		select {
// 		case ch <- message:
// 		default:
// 			log.Println("⚠️ WebSocket channel full, skipping:", chatID)
// 		}
// 	}

// 	log.Printf("📢 Sent message to ChatID=%d: %s", chatID, message)
// }
