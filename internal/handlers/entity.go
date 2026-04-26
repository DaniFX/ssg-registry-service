package handlers

import (
	"net/http"

	"github.com/DaniFX/ssg-nexus-sdk/pkg/nexus"
	nexusRepo "github.com/DaniFX/ssg-nexus-sdk/pkg/nexus/repository"
	"github.com/DaniFX/ssg-registry-service/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid" // go get github.com/google/uuid
)

type EntityHandler struct {
	Repo *nexusRepo.Repository
}

func (h *EntityHandler) Create(c *gin.Context) {
	ctx := c.Request.Context()
	identity := nexus.FromContext(ctx)

	// 1. Parsing del JSON in ingresso
	var payload models.Entity
	if err := c.ShouldBindJSON(&payload); err != nil {
		nexus.Failure(c, http.StatusBadRequest, nexus.ErrValidationFailed, "Formato JSON non valido", err.Error())
		return
	}

	// 2. Generazione ID univoco
	entityID := uuid.New().String()

	// 3. Preparazione dati (L'SDK si aspetta una map[string]interface{})
	data := map[string]interface{}{
		"type":     payload.Type,
		"subTypes": payload.SubTypes,
		"status":   payload.Status,
		"coreData": payload.CoreData,
		"extData":  payload.ExtData,
	}

	// 4. Salvataggio tramite Nexus SDK!
	// Questo inietterà automaticamente createdAt, updatedAt e createdBy (usando identity.UserID)
	err := h.Repo.Create(ctx, entityID, data)
	if err != nil {
		nexus.Failure(c, http.StatusInternalServerError, nexus.ErrInternal, "Errore salvataggio database", err.Error())
		return
	}

	// 5. Risposta Standardizzata
	data["id"] = entityID
	nexus.Success(c, data, gin.H{"insertedBy": identity.UserID})
}
