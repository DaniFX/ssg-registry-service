package handlers

import (
	"net/http"

	"cloud.google.com/go/firestore"
	"github.com/DaniFX/ssg-nexus-sdk/pkg/nexus"
	nexusRepo "github.com/DaniFX/ssg-nexus-sdk/pkg/nexus/repository"
	"github.com/DaniFX/ssg-registry-service/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type EntityHandler struct {
	Repo *nexusRepo.Repository
}

// validateEntity esegue le validazioni di business comuni a Create e Update.
// fix #1: vatNumber obbligatorio per type == ORGANIZATION
func validateEntity(payload *models.Entity, requireType bool) (string, bool) {
	if requireType && payload.Type == "" {
		return "Il campo 'type' è obbligatorio", false
	}
	if payload.Type == "ORGANIZATION" {
		vatNumber, _ := payload.CoreData["vatNumber"].(string)
		if vatNumber == "" {
			return "Il campo 'vatNumber' in coreData è obbligatorio per le entità di tipo ORGANIZATION", false
		}
	}
	return "", true
}

// Create crea una nuova entità nel registro.
func (h *EntityHandler) Create(c *gin.Context) {
	ctx := c.Request.Context()
	identity := nexus.FromContext(ctx)

	var payload models.Entity
	if err := c.ShouldBindJSON(&payload); err != nil {
		nexus.Failure(c, http.StatusBadRequest, nexus.ErrValidationFailed, "Formato JSON non valido", err.Error())
		return
	}

	// fix #1: validazione vatNumber per ORGANIZATION
	if msg, ok := validateEntity(&payload, true); !ok {
		nexus.Failure(c, http.StatusBadRequest, nexus.ErrValidationFailed, msg, "")
		return
	}

	entityID := uuid.New().String()

	data := map[string]interface{}{
		"type":     payload.Type,
		"subTypes": payload.SubTypes,
		"status":   payload.Status,
		"coreData": payload.CoreData,
		"extData":  payload.ExtData,
	}

	if err := h.Repo.Create(ctx, entityID, data); err != nil {
		nexus.Failure(c, http.StatusInternalServerError, nexus.ErrInternal, "Errore salvataggio database", err.Error())
		return
	}

	data["id"] = entityID
	nexus.Success(c, data, gin.H{"insertedBy": identity.UserID})
}

// List restituisce tutte le entità (escluse le soft-deleted, gestito dall'SDK).
// fix #2: endpoint GET /entities attivato
func (h *EntityHandler) List(c *gin.Context) {
	ctx := c.Request.Context()

	filters := map[string]string{}
	for k, v := range c.Request.URL.Query() {
		if len(v) > 0 {
			filters[k] = v[0]
		}
	}

	docs, err := h.Repo.List(ctx, filters)
	if err != nil {
		nexus.Failure(c, http.StatusInternalServerError, nexus.ErrInternal, "Errore recupero entità", err.Error())
		return
	}

	nexus.Success(c, docs, nil)
}

// GetByID recupera un'entità per ID.
// fix #2: endpoint GET /entities/:id attivato
func (h *EntityHandler) GetByID(c *gin.Context) {
	ctx := c.Request.Context()
	id := c.Param("id")

	doc, err := h.Repo.GetByID(ctx, id)
	if err != nil {
		nexus.Failure(c, http.StatusNotFound, nexus.ErrNotFound, "Entità non trovata", err.Error())
		return
	}

	nexus.Success(c, doc, nil)
}

// Update aggiorna un'entità esistente.
// fix #2: endpoint PATCH /entities/:id attivato
// fix #1: se il type è (o diventa) ORGANIZATION, vatNumber rimane obbligatorio
func (h *EntityHandler) Update(c *gin.Context) {
	ctx := c.Request.Context()
	id := c.Param("id")

	var payload models.Entity
	if err := c.ShouldBindJSON(&payload); err != nil {
		nexus.Failure(c, http.StatusBadRequest, nexus.ErrValidationFailed, "Formato JSON non valido", err.Error())
		return
	}

	// fix #1: se il payload indica type ORGANIZATION, vatNumber è obbligatorio
	if msg, ok := validateEntity(&payload, false); !ok {
		nexus.Failure(c, http.StatusBadRequest, nexus.ErrValidationFailed, msg, "")
		return
	}

	// L'SDK.Update verifica internamente IsLocked() e restituisce ErrImmutableRecord se bloccato
	updates := []firestore.Update{}
	if payload.Type != "" {
		updates = append(updates, firestore.Update{Path: "type", Value: payload.Type})
	}
	if payload.SubTypes != nil {
		updates = append(updates, firestore.Update{Path: "subTypes", Value: payload.SubTypes})
	}
	if payload.Status != "" {
		updates = append(updates, firestore.Update{Path: "status", Value: payload.Status})
	}
	if payload.CoreData != nil {
		updates = append(updates, firestore.Update{Path: "coreData", Value: payload.CoreData})
	}
	if payload.ExtData != nil {
		updates = append(updates, firestore.Update{Path: "extData", Value: payload.ExtData})
	}

	if len(updates) == 0 {
		nexus.Failure(c, http.StatusBadRequest, nexus.ErrValidationFailed, "Nessun campo da aggiornare fornito", "")
		return
	}

	if err := h.Repo.Update(ctx, id, updates); err != nil {
		if err.Error() == nexus.ErrImmutableRecord {
			nexus.Failure(c, http.StatusConflict, nexus.ErrImmutableRecord, "Il documento è bloccato e non può essere modificato", "")
			return
		}
		nexus.Failure(c, http.StatusInternalServerError, nexus.ErrInternal, "Errore aggiornamento entità", err.Error())
		return
	}

	nexus.Success(c, gin.H{"id": id}, nil)
}

// Delete esegue il soft delete di un'entità.
// fix #2: endpoint DELETE /entities/:id attivato con SoftDelete()
func (h *EntityHandler) Delete(c *gin.Context) {
	ctx := c.Request.Context()
	id := c.Param("id")

	// SoftDelete() dell'SDK verifica IsLocked() internamente
	if err := h.Repo.SoftDelete(ctx, id); err != nil {
		if err.Error() == nexus.ErrImmutableRecord {
			nexus.Failure(c, http.StatusConflict, nexus.ErrImmutableRecord, "Il documento è bloccato e non può essere eliminato", "")
			return
		}
		nexus.Failure(c, http.StatusInternalServerError, nexus.ErrInternal, "Errore eliminazione entità", err.Error())
		return
	}

	nexus.Success(c, gin.H{"id": id, "deleted": true}, nil)
}
