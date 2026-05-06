package handlers

import (
	"net/http"

	"github.com/DaniFX/ssg-nexus-sdk/pkg/nexus"
	nexusRepo "github.com/DaniFX/ssg-nexus-sdk/pkg/nexus/repository"
	"github.com/DaniFX/ssg-registry-service/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// EntityHandler gestisce tutte le operazioni CRUD sulle anagrafiche
type EntityHandler struct {
	Repo *nexusRepo.Repository
}

// Create gestisce POST /api/v1/registry/entities
func (h *EntityHandler) Create(c *gin.Context) {
	ctx := c.Request.Context()
	identity := nexus.FromContext(ctx)

	var req models.CreateEntityRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		nexus.Failure(c, http.StatusBadRequest, nexus.ErrValidationFailed, "Payload non valido", err.Error())
		return
	}

	// Validazione: vatNumber obbligatoria per ORGANIZATION
	if req.Type == "ORGANIZATION" && req.CoreData.VatNumber == "" {
		nexus.Failure(c, http.StatusBadRequest, nexus.ErrValidationFailed, "P.IVA obbligatoria per ORGANIZATION", "")
		return
	}

	// Validazione: displayName e email obbligatori
	if req.CoreData.DisplayName == "" || req.CoreData.Email == "" {
		nexus.Failure(c, http.StatusBadRequest, nexus.ErrValidationFailed, "displayName ed email sono obbligatori in coreData", "")
		return
	}

	// Status default ACTIVE
	if req.Status == "" {
		req.Status = "ACTIVE"
	}

	entityID := uuid.New().String()

	data := map[string]interface{}{
		"type":     req.Type,
		"subType":  req.SubType,
		"status":   req.Status,
		"coreData": req.CoreData,
		"metadata": req.Metadata,
	}

	if err := h.Repo.Create(ctx, entityID, data); err != nil {
		nexus.Failure(c, http.StatusInternalServerError, nexus.ErrInternal, "Errore salvataggio", err.Error())
		return
	}

	data["id"] = entityID
	nexus.Success(c, data, gin.H{"insertedBy": identity.UserID})
}

// List gestisce GET /api/v1/registry/entities
// Supporta Navigator Pattern: ?status=ACTIVE&sort=-createdAt&limit=20&offset=0
func (h *EntityHandler) List(c *gin.Context) {
	ctx := c.Request.Context()

	var entities []models.Entity
	meta, err := h.Repo.ApplyNavigator(ctx, c.Request.URL.Query(), &entities)
	if err != nil {
		nexus.Failure(c, http.StatusInternalServerError, nexus.ErrInternal, "Errore lettura lista", err.Error())
		return
	}

	nexus.Success(c, entities, meta)
}

// GetByID gestisce GET /api/v1/registry/entities/:id
func (h *EntityHandler) GetByID(c *gin.Context) {
	ctx := c.Request.Context()
	id := c.Param("id")

	var entity models.Entity
	if err := h.Repo.GetByID(ctx, id, &entity); err != nil {
		nexus.Failure(c, http.StatusNotFound, nexus.ErrNotFound, "Entità non trovata", err.Error())
		return
	}

	nexus.Success(c, entity, nil)
}

// Update gestisce PATCH /api/v1/registry/entities/:id
func (h *EntityHandler) Update(c *gin.Context) {
	ctx := c.Request.Context()
	id := c.Param("id")

	// Verifica esistenza e lock
	var existing models.Entity
	if err := h.Repo.GetByID(ctx, id, &existing); err != nil {
		nexus.Failure(c, http.StatusNotFound, nexus.ErrNotFound, "Entità non trovata", err.Error())
		return
	}
	if h.Repo.IsLocked(existing.NexusDoc) {
		nexus.Failure(c, http.StatusConflict, nexus.ErrImmutableRecord, "Record non modificabile", "")
		return
	}

	var req models.UpdateEntityRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		nexus.Failure(c, http.StatusBadRequest, nexus.ErrValidationFailed, "Payload non valido", err.Error())
		return
	}

	updates := map[string]interface{}{}
	if req.SubType != "" {
		updates["subType"] = req.SubType
	}
	if req.Status != "" {
		updates["status"] = req.Status
	}
	if req.CoreData != nil {
		updates["coreData"] = req.CoreData
	}
	if req.Metadata != nil {
		updates["metadata"] = req.Metadata
	}

	if len(updates) == 0 {
		nexus.Failure(c, http.StatusBadRequest, nexus.ErrValidationFailed, "Nessun campo da aggiornare", "")
		return
	}

	if err := h.Repo.Update(ctx, id, updates); err != nil {
		nexus.Failure(c, http.StatusInternalServerError, nexus.ErrInternal, "Errore aggiornamento", err.Error())
		return
	}

	nexus.Success(c, gin.H{"id": id, "updated": true}, nil)
}

// Delete gestisce DELETE /api/v1/registry/entities/:id (soft delete)
func (h *EntityHandler) Delete(c *gin.Context) {
	ctx := c.Request.Context()
	id := c.Param("id")

	// Verifica esistenza
	var existing models.Entity
	if err := h.Repo.GetByID(ctx, id, &existing); err != nil {
		nexus.Failure(c, http.StatusNotFound, nexus.ErrNotFound, "Entità non trovata", err.Error())
		return
	}

	if err := h.Repo.SoftDelete(ctx, id); err != nil {
		nexus.Failure(c, http.StatusInternalServerError, nexus.ErrInternal, "Errore eliminazione", err.Error())
		return
	}

	nexus.Success(c, gin.H{"id": id, "deleted": true}, nil)
}
