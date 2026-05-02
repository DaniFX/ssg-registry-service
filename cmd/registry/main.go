package main

import (
	"context"
	"log"
	"os"

	"github.com/gin-gonic/gin"

	"github.com/DaniFX/ssg-nexus-sdk/pkg/nexus"
	nexusRepo "github.com/DaniFX/ssg-nexus-sdk/pkg/nexus/repository"

	"github.com/DaniFX/ssg-registry-service/internal/handlers"
	"github.com/DaniFX/ssg-registry-service/internal/repository"
)

func main() {
	// 1. Setup Firestore e Nexus Repository
	ctx := context.Background()
	firestoreClient := repository.InitFirestore(ctx)
	defer firestoreClient.Close()

	// Inizializziamo l'ORM dell'SDK per la collezione "entities"
	entityRepo := nexusRepo.NewRepository(firestoreClient, "entities")

	r := gin.Default()

	// 2. Service Discovery (DEVE ESSERE PRIMA DEL GUARD O FUORI DAL GRUPPO)
	def := nexus.ServiceDefinition{
		ServiceName: "registry-service",
		Version:     "1.0.0",
		Endpoints: []nexus.Endpoint{
			{
				Path:         "/api/v1/registry/entities",
				Method:       "GET",
				AuthRequired: true,
				Summary:      "Lista tutte le entità nel registro",
			},
			{
				Path:         "/api/v1/registry/entities",
				Method:       "POST",
				AuthRequired: true,
				Summary:      "Crea una nuova entità nel registro",
			},
			{
				Path:         "/api/v1/registry/entities/:id",
				Method:       "GET",
				AuthRequired: true,
				Summary:      "Recupera un'entità per ID",
			},
			{
				Path:         "/api/v1/registry/entities/:id",
				Method:       "PATCH",
				AuthRequired: true,
				Summary:      "Aggiorna un'entità nel registro",
			},
			{
				Path:         "/api/v1/registry/entities/:id",
				Method:       "DELETE",
				AuthRequired: true,
				Summary:      "Elimina (soft delete) un'entità dal registro",
			},
		},
	}

	// Registra in automatico GET /_discover senza il Guard
	nexus.RegisterDiscovery(r, def)

	// Avvia l'handshake (PUSH verso il Gateway)
	nexus.StartGatewayHandshake(def)

	// 3. Rotte di Business (PROTETTE DAL GUARD)
	h := &handlers.EntityHandler{Repo: entityRepo}

	api := r.Group("/api/v1/registry")
	api.Use(nexus.Guard())
	{
		api.GET("/entities", h.List)
		api.POST("/entities", h.Create)
		api.GET("/entities/:id", h.GetByID)
		api.PATCH("/entities/:id", h.Update)
		api.DELETE("/entities/:id", h.Delete)
	}

	// fix #3: PORT letta da variabile d'ambiente (Cloud Run compliance)
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080" // fallback locale
	}

	log.Printf("Avvio Registry Service sulla porta %s...", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("Errore critico del server: %v", err)
	}
}
