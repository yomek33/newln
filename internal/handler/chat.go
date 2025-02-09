package handler

import (
	"bufio"
	"log"
	"net/http"
	"strconv"
	"sync"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/yomek33/newln/internal/services"
)

// ChatWSHandler 構造体
type ChatWSHandler struct {
	service     services.ChatService
	subscribers map[uint][]chan string
	mu          sync.Mutex
}

// NewChatWSHandler コンストラクタ
func NewChatWSHandler(service services.ChatService) *ChatWSHandler {
	return &ChatWSHandler{
		service:     service,
		subscribers: make(map[uint][]chan string),
	}
}
func (h *ChatWSHandler) ChatWebSocket(c echo.Context) error {
	chatID, err := strconv.Atoi(c.Param("chatID"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid chat ID"})
	}

	// ユーザー認証（UserID 取得）
	userID, err := getUserIDFromContext(c)
	if err != nil {
		return respondWithError(c, http.StatusUnauthorized, ErrInvalidUserToken)
	}

	// WebSocket 接続を確立
	conn, _, err := c.Response().Hijack()
	if err != nil {
		log.Println("WebSocket hijack failed:", err)
		return err
	}
	defer conn.Close()

	// クライアントの WebSocket 用のチャネルを作成
	ch := h.SubscribeToChat(uint(chatID))
	defer h.UnsubscribeFromChat(uint(chatID), ch)

	// WebSocket の送信ループ
	go func() {
		for msg := range ch {
			_, err := conn.Write([]byte(msg + "\n")) // 改行付きで送信
			if err != nil {
				log.Println("WebSocket send error:", err)
				return
			}
		}
	}()

	// WebSocket の受信ループ
	reader := bufio.NewReader(conn)
	for {
		message, err := reader.ReadString('\n')
		if err != nil {
			log.Println("WebSocket read error:", err)
			break
		}

		// 🔥 `echo.Context` を渡すように修正
		h.SendMessageWS(c, uint(chatID), userID, message)
	}

	return nil
}

// 🔹 メッセージを WebSocket で送信（echo.Context を引数に追加）
func (h *ChatWSHandler) SendMessageWS(c echo.Context, chatID uint, userID uuid.UUID, message string) {
	response, err := h.service.SendMessage(chatID, userID, message)
	if err != nil {
		log.Println("❌ Failed to send message:", err)
		return
	}

	// チャットが終了している場合、特別なメッセージを送る
	if response.Content == "🚀 チャット終了！次のステップへ進みます。（後で実装）" {
		h.PublishChatUpdate(chatID, "🚀 チャットが終了しました。")
		return
	}

	// クライアントにリアルタイム配信
	h.PublishChatUpdate(chatID, response.Content)
}


func (h *ChatWSHandler) SubscribeToChat(chatID uint) chan string {
	h.mu.Lock()
	defer h.mu.Unlock()

	ch := make(chan string, 10)
	h.subscribers[chatID] = append(h.subscribers[chatID], ch)

	return ch
}

func (h *ChatWSHandler) UnsubscribeFromChat(chatID uint, ch chan string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	channels, exists := h.subscribers[chatID]
	if !exists {
		return
	}

	newChannels := make([]chan string, 0, len(channels))
	for _, c := range channels {
		if c != ch {
			newChannels = append(newChannels, c)
		}
	}

	if len(newChannels) == 0 {
		delete(h.subscribers, chatID)
	} else {
		h.subscribers[chatID] = newChannels
	}

	close(ch)
}

// 🔹 クライアントにメッセージを送信
func (h *ChatWSHandler) PublishChatUpdate(chatID uint, message string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	subscribers, ok := h.subscribers[chatID]
	if !ok {
		return
	}

	for _, ch := range subscribers {
		select {
		case ch <- message:
		default:
			log.Println("⚠️ WebSocket channel full, skipping:", chatID)
		}
	}
}
