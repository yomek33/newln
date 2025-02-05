package handler

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/yomek33/newln/internal/logger"
	"github.com/yomek33/newln/internal/models"
	"github.com/yomek33/newln/internal/services"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/oklog/ulid/v2"
)

type MaterialHandler struct {
	MaterialService services.MaterialService
	PhraseService   services.PhraseService
	WordService     services.WordService
}

func NewMaterialHandler(materialService services.MaterialService, phraseService services.PhraseService, wordService services.WordService) *MaterialHandler {
	return &MaterialHandler{
		MaterialService: materialService,
		PhraseService:   phraseService,
		WordService:     wordService,
	}
}

type MaterialResponse struct {
	ID        uint
	LocalULID string
	Content   string
	Title     string
	Status    string
	CreatedAt time.Time
	UpdatedAt time.Time
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
	material.LocalULID = ulid.Make().String()

	ctx, cancel := context.WithTimeout(c.Request().Context(), 10*time.Second)
	defer cancel()

	createdMaterial, err := h.MaterialService.CreateMaterial(&material)
	if err != nil {
		logger.Errorf("Error creating material: %v, UserID: %v", err, UserID)
		return respondWithError(c, http.StatusInternalServerError, ErrFailedCreateMaterial)
	}
	go h.processMaterialAsync(ctx, createdMaterial.ID, UserID)

	response := MaterialResponse{
		LocalULID: createdMaterial.LocalULID,
		Content:   createdMaterial.Content,
		Title:     createdMaterial.Title,
		Status:    createdMaterial.Status,
		CreatedAt: createdMaterial.CreatedAt,
		UpdatedAt: createdMaterial.UpdatedAt,
	}

	logger.Infof("Material created successfully: %+v", createdMaterial)
	return c.JSON(http.StatusCreated, response)
}

func (h *MaterialHandler) GetMaterialByID(c echo.Context) error {
	ulid := c.Param("ulid")

	UserID, err := getUserIDFromContext(c)
	if err != nil {
		return respondWithError(c, http.StatusUnauthorized, ErrInvalidUserToken)
	}

	material, err := h.MaterialService.GetMaterialByULID(ulid, UserID)
	if err != nil {
		return respondWithError(c, http.StatusNotFound, ErrMaterialNotFound)
	}

	logger.Infof("Retrieved material MaterialULID;%v", ulid)
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

func (h *MaterialHandler) processMaterialAsync(ctx context.Context, materialID uint, userID uuid.UUID) {
	h.MaterialService.UpdateMaterialStatus(materialID, "processing")

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
			errChan <- fmt.Errorf("failed to create phrase list: %w", err)
		}
	}()

	// ✅ WordList を非同期で作成
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := h.WordService.CreateWordList(&wordList); err != nil {
			errChan <- fmt.Errorf("failed to create word list: %w", err)
		}
	}()

	// ✅ `GeneratePhrases` を非同期で処理
	phrasesChan := make(chan []models.Phrase, 1)
	wg.Add(1)
	go func() {
		defer wg.Done()
		phrases, err := h.PhraseService.GeneratePhrases(ctx, materialID)
		if err != nil {
			errChan <- fmt.Errorf("failed to generate phrases: %w", err)
			return
		}
		phrasesChan <- phrases
	}()

	// ✅ `GenerateWords` を非同期で処理
	wordsChan := make(chan []models.Word, 1)
	wg.Add(1)
	go func() {
		defer wg.Done()
		words, err := h.WordService.GenerateWords(ctx, materialID)
		if err != nil {
			errChan <- fmt.Errorf("failed to generate words: %w", err)
			return
		}
		wordsChan <- words
	}()

	wg.Wait()
	close(errChan)
	close(phrasesChan)
	close(wordsChan)

	// ✅ エラーチェック
	for err := range errChan {
		logger.Error(fmt.Errorf("Error: %v, materialID: %v, userID: %v", err, materialID, userID))
		h.MaterialService.UpdateMaterialStatus(materialID, "failed")
		return
	}

	// ✅ フレーズの処理
	phrases := <-phrasesChan
	if len(phrases) == 0 {
		log.Printf("No phrases generated, materialID: %v, userID: %v", materialID, userID)
		h.MaterialService.UpdateMaterialStatus(materialID, "failed")
		return
	}
	for i := range phrases {
		phrases[i].PhraseListID = phraseList.ID
	}

	if err := h.PhraseService.BulkInsertPhrases(phrases); err != nil {
		log.Printf("Failed to store phrases: %v, materialID: %v, userID: %v", err, materialID, userID)
		h.MaterialService.UpdateMaterialStatus(materialID, "failed")
		return
	}

	// ✅ ワードの処理
	words := <-wordsChan
	if len(words) == 0 {
		log.Printf("No words generated, materialID: %v, userID: %v", materialID, userID)
		h.MaterialService.UpdateMaterialStatus(materialID, "failed")
		return
	}
	for i := range words {
		words[i].WordListID = wordList.ID
	}

	if err := h.WordService.BulkInsertWords(words); err != nil {
		log.Printf("Failed to store words: %v, materialID: %v, userID: %v", err, materialID, userID)
		h.MaterialService.UpdateMaterialStatus(materialID, "failed")
		return
	}

	logger.Infof("Phrases and Words generated and stored successfully, materialID: %v, userID: %v", materialID, userID)
	h.MaterialService.UpdateMaterialStatus(materialID, "completed")
}
