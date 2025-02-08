package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/yomek33/newln/internal/logger"
	"github.com/yomek33/newln/internal/models"
	"github.com/yomek33/newln/internal/services"
	"golang.org/x/net/websocket"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/oklog/ulid/v2"
)

type MaterialHandler struct {
	MaterialService services.MaterialService
	PhraseService   services.PhraseService
	WordService     services.WordService
	jwtSecret       []byte
}

func NewMaterialHandler(materialService services.MaterialService, phraseService services.PhraseService, wordService services.WordService, jwtSecret []byte) *MaterialHandler {
	return &MaterialHandler{
		MaterialService: materialService,
		PhraseService:   phraseService,
		WordService:     wordService,
		jwtSecret: jwtSecret,
	}
}

type MaterialResponse struct {
	ID        uint
	ULID      string
	Content   string
	Title     string
	Status    string
	CreatedAt time.Time
	UpdatedAt time.Time
	HasPendingPhraseList bool
	HasPendingWordList	bool
}

func (h *MaterialHandler) CreateMaterial(c echo.Context) error {
	var material models.Material
	if err := bindAndValidateMaterial(c, &material); err != nil {
		return respondWithError(c, http.StatusBadRequest, err.Error())
	}

	UserID, err := getUserIDFromContext(c)
	if err != nil {
		return respondWithError(c, http.StatusUnauthorized, ErrInvalidUserToken)
	}

	material.UserID = UserID
	material.Status = "draft"
	material.ULID = ulid.Make().String()
	material.HasPendingPhraseList = true
	material.HasPendingWordList = true

	ctx, cancel := context.WithTimeout(c.Request().Context(), 10*time.Second)
	defer cancel()

	createdMaterial, err := h.MaterialService.CreateMaterial(&material)
	if err != nil {
		logger.Errorf("Error creating material: %v, UserID: %v", err, UserID)
		return respondWithError(c, http.StatusInternalServerError, ErrFailedCreateMaterial)
	}
	go h.processMaterialAsync(ctx, createdMaterial.ID, createdMaterial.ULID, UserID)
	logger.Infof("createdMaterial: %+v", createdMaterial)
	response := MaterialResponse{
		ULID:      material.ULID,
		Content:   material.Content,
		Title:     material.Title,
		Status:    createdMaterial.Status,
		CreatedAt: createdMaterial.CreatedAt,
		UpdatedAt: createdMaterial.UpdatedAt,
		HasPendingPhraseList: true,
		HasPendingWordList:   true,
	}

	logger.Infof("Material created successfully: %+v", createdMaterial)
	return c.JSON(http.StatusCreated, response)
}

func (h *MaterialHandler) GetMaterialByULID(c echo.Context) error {
	ulid := c.Param("ulid")

	UserID, err := getUserIDFromContext(c)
	if err != nil {
		return respondWithError(c, http.StatusUnauthorized, ErrInvalidUserToken)
	}

	material, err := h.MaterialService.GetMaterialByULID(ulid, UserID)
	if err != nil {
		return respondWithError(c, http.StatusNotFound, ErrMaterialNotFound)
	}

	logger.Infof("Retrieved material Material HasPendingPhraseList;%v", material.HasPendingPhraseList)
	return c.JSON(http.StatusOK, material)
}

func (h *MaterialHandler) UpdateMaterial(c echo.Context) error {
	UserID, err := getUserIDFromContext(c)
	if err != nil {
		return respondWithError(c, http.StatusUnauthorized, ErrInvalidUserToken)
	}

	ulid := c.Param("ulid")
	material, err := h.MaterialService.GetMaterialByULID(ulid, UserID)
	if err != nil {
		return respondWithError(c, http.StatusNotFound, ErrMaterialNotFound)
	}

	if err := bindAndValidateMaterial(c, material); err != nil {
		return respondWithError(c, http.StatusBadRequest, err.Error())
	}

	if material.UserID != UserID {
		return respondWithError(c, http.StatusForbidden, ErrForbiddenModify)
	}

	if err := h.MaterialService.UpdateMaterial(ulid, material); err != nil {
		logger.Errorf("Failed to update material: %v, MaterialID: %v, UserID: %v", err, ulid, UserID)
		return respondWithError(c, http.StatusInternalServerError, ErrFailedUpdateMaterial)
	}

	logger.Infof("Updated material, MaterialID: %v, UserID: %v", ulid, UserID)
	return c.JSON(http.StatusOK, material)
}

func (h *MaterialHandler) DeleteMaterial(c echo.Context) error {
	ulid := c.Param("ulid")

	UserID, err := getUserIDFromContext(c)
	if err != nil {
		return respondWithError(c, http.StatusUnauthorized, ErrInvalidUserToken)
	}

	if err := h.MaterialService.DeleteMaterial(ulid, UserID); err != nil {
		logger.Errorf("Failed to delete material: %v, MaterialID: %v, UserID: %v", err, ulid, UserID)
		return respondWithError(c, http.StatusInternalServerError, ErrFailedDeleteMaterial)
	}

	logger.Infof("Deleted material, MaterialID: %v, UserID: %v", ulid, UserID)
	return c.NoContent(http.StatusNoContent)
}

func (h *MaterialHandler) GetAllMaterials(c echo.Context) error {
	searchQuery := c.QueryParam("search")

	UserID, err := getUserIDFromContext(c)
	if err != nil {
		return respondWithError(c, http.StatusUnauthorized, ErrInvalidUserToken)
	}

	materials, err := h.MaterialService.GetAllMaterials(searchQuery, UserID)
	if err != nil {
		logger.Errorf("Failed to retrieve materials: %v, UserID: %v", err, UserID)
		return respondWithError(c, http.StatusInternalServerError, ErrFailedRetrieveMaterials)
	}

	logger.Infof("Retrieved materials, MaterialCount: %v, UserID: %v", len(materials), UserID)
	return c.JSON(http.StatusOK, materials)
}

func (h *MaterialHandler) CheckMaterialStatus(c echo.Context) error {
	ulid := c.Param("ulid")

	status, err := h.MaterialService.GetMaterialStatus(ulid)
	if err != nil {
		logger.Errorf("Failed to get material status: %v, MaterialID: %v", err, ulid)
		return respondWithError(c, http.StatusInternalServerError, err.Error())
	}

	logger.Infof("Checked material status, MaterialID: %v, Status: %v", ulid, status)
	return c.JSON(http.StatusOK, map[string]string{"status": status})
}

func (h *MaterialHandler) processMaterialAsync(ctx context.Context, materialID uint, materialULID string, userID uuid.UUID) {
	logger.Infof("🚀 Starting async processing for materialID: %v, userID: %v", materialID, userID)
	h.MaterialService.UpdateMaterialStatus(materialID, "processing")

	// ✅ ステータス変更を SSE で送信
	h.MaterialService.PublishMaterialUpdate(materialULID, `{"event": "processing"}`)

	// ✅ PhraseList を作成
	phraseList := models.PhraseList{
		MaterialID: materialID,
		Title:      "Default Phrase List",
	}

	// ✅ WordList を作成
	wordList := models.WordList{
		MaterialID: materialID,
		Title:      "Default Word List",
	}

	var wg sync.WaitGroup
	errChan := make(chan error, 4)

	// ✅ PhraseList を非同期で作成
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := h.PhraseService.CreatePhraseList(&phraseList); err != nil {
			errChan <- fmt.Errorf("❌ failed to create phrase list: %w", err)
		}
	}()

	// ✅ WordList を非同期で作成
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := h.WordService.CreateWordList(&wordList); err != nil {
			errChan <- fmt.Errorf("❌ failed to create word list: %w", err)
		}
	}()

	// ✅ `GeneratePhrases` を非同期で処理
	phrasesChan := make(chan []models.Phrase, 1)
	wg.Add(1)
	go func() {
		defer wg.Done()
		phrases, err := h.PhraseService.GeneratePhrases(ctx, materialID)
		if err != nil {
			errChan <- fmt.Errorf("⚠️ No phrases generated: %w", err)
			phrasesChan <- nil
			return
		}
		phrasesChan <- phrases
				if err := h.MaterialService.UpdateHasPendingPhraseStatus(materialULID, false); err != nil {
			logger.Errorf("❌ Failed to update HasPhraseList: %v", err)
			errChan <- fmt.Errorf("❌ failed to update HasPhraseList: %w", err)
		}
	}()

	// ✅ `GenerateWords` を非同期で処理
	wordsChan := make(chan []models.Word, 1)
	wg.Add(1)
	go func() {
		defer wg.Done()
		words, err := h.WordService.GenerateWords(ctx, materialID)
		if err != nil {
			errChan <- fmt.Errorf("❌ failed to generate words: %w", err)
			wordsChan <- nil
			return
		}
		wordsChan <- words
		if err := h.MaterialService.UpdateHasPendingWordStatus(materialULID, false); err != nil {
			logger.Errorf("❌ Failed to update HasPendingWordList: %v", err)
			errChan <- fmt.Errorf("❌ failed to update HasPendingWordList: %w", err)
		}
	}()

	wg.Wait()
	close(errChan)
	close(phrasesChan)
	close(wordsChan)

	// ✅ エラーチェック
	hasError := false
	for err := range errChan {
		logger.Errorf("❌ Error occurred: %v, materialID: %v, userID: %v", err, materialID, userID)
		h.MaterialService.PublishMaterialUpdate(materialULID, fmt.Sprintf(`{"event": "error", "message": "%s"}`, err.Error()))
		hasError = true
	}

	// ✅ フレーズの処理 & SSE 送信
	phrases := <-phrasesChan
	if phrases != nil && len(phrases) > 0 {
		for i := range phrases {
			phrases[i].PhraseListID = phraseList.ID
		}
		if err := h.PhraseService.BulkInsertPhrases(phrases); err != nil {
			hasError = true
			logger.Errorf("❌ Failed to store phrases: %v", err)
			h.MaterialService.PublishMaterialUpdate(materialULID, fmt.Sprintf(`{"event": "error", "message": "%s"}`, err.Error()))
		} else {
			// ✅ phrases を SSE で送信
			phrasesJSON, _ := json.Marshal(phrases)
			h.MaterialService.PublishMaterialUpdate(materialULID, fmt.Sprintf(`{"event": "phrases_stored", "data": %s}`, phrasesJSON))
		}
	} else {
		logger.Warnf("⚠️ No phrases were stored, materialULID: %v", materialULID)
	}

	// ✅ ワードの処理 & SSE 送信
	words := <-wordsChan
	if words != nil && len(words) > 0 {
		for i := range words {
			words[i].WordListID = wordList.ID
		}
		if err := h.WordService.BulkInsertWords(words); err != nil {
			hasError = true
			logger.Errorf("❌ Failed to store words: %v", err)
			h.MaterialService.PublishMaterialUpdate(materialULID, fmt.Sprintf(`{"event": "error", "message": "%s"}`, err.Error()))
		} else {
			// ✅ words を SSE で送信
			wordsJSON, _ := json.Marshal(words)
			h.MaterialService.PublishMaterialUpdate(materialULID, fmt.Sprintf(`{"event": "words_stored", "data": %s}`, wordsJSON))
		}
	} else {
		logger.Warnf("⚠️ No words were stored, materialULID: %v", materialULID)
	}

	// ✅ 最終ステータス更新 & SSE 送信
	if hasError {
		h.MaterialService.UpdateMaterialStatus(materialID, "failed")
		h.MaterialService.PublishMaterialUpdate(materialULID, `{"event": "failed"}`)
	} else {
		h.MaterialService.UpdateMaterialStatus(materialID, "completed")
		h.MaterialService.PublishMaterialUpdate(materialULID, `{"event": "completed"}`)
	}
}
func (h *MaterialHandler) StreamMaterialProgressWS(c echo.Context) error {
    materialULID := c.Param("ulid")
    tokenString := c.QueryParam("token")
    if tokenString == "" {
        return c.JSON(http.StatusUnauthorized, map[string]string{"error": "missing token"})
    }
	_, err := isValidJWTToken(tokenString, h.jwtSecret)
    if err != nil {
        return c.JSON(http.StatusUnauthorized, map[string]string{"error": err.Error()})
    }

    websocket.Handler(func(ws *websocket.Conn) {
        defer ws.Close()

        ch := h.MaterialService.SubscribeToMaterialUpdates(materialULID)
        defer h.MaterialService.UnsubscribeFromMaterialUpdates(materialULID, ch)

        for msg := range ch {
            if err := websocket.Message.Send(ws, msg); err != nil {
                log.Println("WebSocket send error:", err)
                break
            }
        }
    }).ServeHTTP(c.Response(), c.Request())

    return nil
}