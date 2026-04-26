package main

import (
	"context"
	"log"

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
	// È buona pratica usare path assoluti che includano il nome del servizio
	def := nexus.ServiceDefinition{
		ServiceName: "registry-service",
		Version:     "1.0.0",
		Endpoints: []nexus.Endpoint{
			{
				Path:         "/api/v1/registry/entities",
				Method:       "POST",
				AuthRequired: true,
				Summary:      "Crea una nuova entità nel registro",
			},
		},
	}

	// Registra in automatico GET /_discover senza il Guard
	nexus.RegisterDiscovery(r, def)

	// Avvia l'handshake (PUSH verso il Gateway)
	nexus.StartGatewayHandshake(def)

	// 3. Rotte di Business (PROTETTE DAL GUARD)
	h := &handlers.EntityHandler{Repo: entityRepo}

	// Creiamo un gruppo specifico per le API di questo servizio
	api := r.Group("/api/v1/registry")

	// Applichiamo la sicurezza SOLO a questo gruppo!
	api.Use(nexus.Guard())
	{
		api.POST("/entities", h.Create)
		// api.GET("/entities", h.List)
		// ... altre rotte protette
	}

	log.Println("Avvio Registry Service sulla porta 8080...")
	if err := r.Run(":8080"); err != nil {
		log.Fatalf("Errore critico del server: %v", err)
	}
}
