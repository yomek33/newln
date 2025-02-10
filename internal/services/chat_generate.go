package services

// func (s *chatService) StartChat(chatID uint, chat *models.Chat) (*models.Chat, error) {

// 	// 🔹 AI に会話開始をリクエスト
// 	prompt := "Hello! How can I help you today?"
// 	session := s.vertex.StartChat(prompt)

// 	// 🔹 AI からの最初のメッセージを取得
// 	ctx := context.Background()
// 	firstResponse, err := session.SendChatMessage(ctx, prompt)
// 	if err != nil {
// 		log.Printf("❌ Failed to get first AI message: %v", err)
// 		return nil, err
// 	}

// 	// 🔒 セッション登録（スレッドセーフ）
// 	s.mu.Lock()
// 	s.sessions[chat.ID] = session
// 	s.mu.Unlock()

// 	// ✅ AI の最初のメッセージを DB に保存
// 	firstMessage := &models.Message{
// 		ChatID:    chatID,
// 		UserID:     models.GeminiUserID, // Gemini AI の ID
// 		Content:    firstResponse,       // AI の返答
// 		SenderType: models.SenderSystem, // システムメッセージ
// 	}

// 	err = s.store.CreateMessage(firstMessage)
// 	if err != nil {
// 		log.Printf("❌ Failed to save first AI message: %v", err)
// 		return nil, err
// 	}

// 	log.Printf("✅ Chat started: ChatID=%d, FirstMessage=%s", chat.ID, firstResponse)
// 	return chat, nil
// }

// func (s *chatService) SendMessage(chatID uint, userID uuid.UUID, message string) (*models.Message, error) {
//     if chatID == 0 {
//         return nil, errors.New("chatID cannot be zero")
//     }
//     if message == "" {
//         return nil, errors.New("message cannot be empty")
//     }
//     if userID == uuid.Nil {
//         return nil, errors.New("userID cannot be empty")
//     }

//     // 現在のチャット情報を取得
//     chat, err := s.store.GetChatByID(chatID)
//     if err != nil {
//         return nil, err
//     }

//     // メッセージの回数をチェック（10回で終了）
//     if chat.PendingMessage >= 10 {
//         s.mu.Lock()
//         delete(s.sessions, chatID) // ✅ 10回以上送信したら削除
//         s.mu.Unlock()

//         return &models.Message{
//             ChatID:     chatID,
//             UserID:     models.GeminiUserID,
//             Content:    "🚀 チャット終了！次のステップへ進みます。（後で実装）",
//             SenderType: models.SenderSystem,
//         }, nil
//     }

//     // ユーザーのメッセージを DB に保存
//     userMessage := &models.Message{
//         ChatID:     chatID,
//         UserID:     userID,
//         Content:    message,
//         SenderType: models.SenderUser,
//     }
//     err = s.store.CreateMessage(userMessage)
//     if err != nil {
//         return nil, err
//     }

//     // Gemini にメッセージを送信
//     session, exists := s.sessions[chatID]
//     if !exists {
//         return nil, errors.New("chat session not found")
//     }

//     ctx := context.Background()
//     response, err := session.SendChatMessage(ctx, message)
//     if err != nil {
//         log.Printf("❌ Error in chat session %d: %v", chatID, err)

//         s.mu.Lock()
//         // delete(s.sessions, chatID) // ✅ エラー発生時にセッション削除
//         s.mu.Unlock()

//         return nil, err
//     }

//     // Gemini のレスポンスを DB に保存
//     botMessage := &models.Message{
//         ChatID:     chatID,
//         UserID:     models.GeminiUserID,
//         Content:    response,
//         SenderType: models.SenderSystem,
//     }
//     err = s.store.CreateMessage(botMessage)
//     if err != nil {
//         return nil, err
//     }

//     // メッセージ回数を更新
//     err = s.store.UpdatePendingMessages(chatID, chat.PendingMessage+1)
//     if err != nil {
//         return nil, err
//     }

//     log.Printf("💬 User: %s", message)
//     log.Printf("🤖 Gemini: %s", response)

//     return botMessage, nil
// }
