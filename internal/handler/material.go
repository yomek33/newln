package handler

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"newln/internal/logger"
	"newln/internal/models"
	"newln/internal/services"
	"newln/internal/sse"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/oklog/ulid/v2"
)

type MaterialHandler struct {
	MaterialService services.MaterialService
	PhraseService   services.PhraseService
	WordService     services.WordService
	SSEManager	  *sse.SSEManager
}

func NewMaterialHandler(materialService services.MaterialService, phraseService services.PhraseService, wordService services.WordService, sseManager *sse.SSEManager) *MaterialHandler {
    return &MaterialHandler{
        MaterialService: materialService,
        PhraseService:   phraseService,
        WordService:     wordService,
        SSEManager:      sseManager,
    }
}

type MaterialResponse struct {
	ID        uint      
	LocalULID string   
	Content   string    
	Title    string   
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
		Title: createdMaterial.Title,
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
	h.SSEManager.Broadcast(fmt.Sprintf(`{"material_id":%d, "status":"processing"}`, materialID))

	var wg sync.WaitGroup
	errChan := make(chan error, 2)

	// Phraseを生成する
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := h.PhraseService.HandlePhraseGeneration(ctx, materialID, h.SSEManager); err != nil {
			errChan <- err
		}
	}()

	// Wordを生成する
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := h.WordService.HandleWordGeneration(ctx, materialID, h.SSEManager); err != nil {
			errChan <- err
		}
	}()

	wg.Wait()
	close(errChan)

	for err := range errChan {
		logger.Error(fmt.Errorf("Error: %v, materialID: %v, userID: %v", err, materialID, userID))
		h.MaterialService.UpdateMaterialStatus(materialID, "failed")
		h.SSEManager.Broadcast(fmt.Sprintf(`{"material_id":%d, "status":"failed"}`, materialID))
		return
	}

	logger.Infof("Processing completed, materialID: %v, userID: %v", materialID, userID)
	h.MaterialService.UpdateMaterialStatus(materialID, "completed")
	h.SSEManager.Broadcast(fmt.Sprintf(`{"material_id":%d, "status":"completed"}`, materialID))
}
