package handler

import (
	"context"
	"net/http"
	"time"

	"newln/internal/logger"
	"newln/internal/models"
	"newln/internal/services"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

type MaterialHandler struct {
	MaterialService services.MaterialService
	PhraseService   services.PhraseService
}

func NewMaterialHandler(materialService services.MaterialService, phraseService services.PhraseService) *MaterialHandler {
	return &MaterialHandler{
		MaterialService: materialService,
		PhraseService:   phraseService,
	}
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

	ctx, cancel := context.WithTimeout(c.Request().Context(), 10*time.Second)
	defer cancel()

	createdMaterial, err := h.MaterialService.CreateMaterial(&material)
	if err != nil {
		logger.Errorf("Error creating material: %v, UserID: %v", err, UserID)
		return respondWithError(c, http.StatusInternalServerError, ErrFailedCreateMaterial)
	}
	go h.processMaterialAsync(ctx, createdMaterial.LocalULID, UserID)

	logger.Info("Material created successfully")
	return c.JSON(http.StatusCreated, map[string]interface{}{
		"message": "Material created successfully",
		"localId": createdMaterial.LocalULID,
	})
}

func (h *MaterialHandler) GetMaterialByID(c echo.Context) error {
	ulid := c.Param("ulid")

	UserID, err := getUserIDFromContext(c)
	if err != nil {
		return respondWithError(c, http.StatusUnauthorized, ErrInvalidUserToken)
	}

	material, err := h.MaterialService.GetMaterialByID(ulid, UserID)
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
	material, err := h.MaterialService.GetMaterialByID(ulid, UserID)
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

func (h *MaterialHandler) processMaterialAsync(ctx context.Context, materialULID string, UserID uuid.UUID) {
	h.MaterialService.UpdateMaterialStatus(materialULID, "processing")

	phrases, err := h.PhraseService.GeneratePhrases(ctx, materialULID, UserID)
	if err != nil {
		logger.Errorf("Failed to generate phrases: %v, materialULID: %v, UserID: %v", err, materialULID, UserID)
		h.MaterialService.UpdateMaterialStatus(materialULID, "failed")
		return
	}

	if err = h.PhraseService.StorePhrases(materialULID, phrases); err != nil {
		logger.Errorf("Failed to store phrases: %v, materialULID: %v, UserID: %v", err, materialULID, UserID)
		h.MaterialService.UpdateMaterialStatus(materialULID, "failed")
		return
	}

	logger.Infof("Phrases generated and stored successfully, materialULID: %v, UserID: %v", materialULID, UserID)
	h.MaterialService.UpdateMaterialStatus(materialULID, "completed")
}
