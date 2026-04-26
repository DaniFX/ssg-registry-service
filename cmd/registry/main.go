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

	// 2. Sicurezza (Middleware)
	r.Use(nexus.Guard())

	// 3. Service Discovery
	def := nexus.ServiceDefinition{
		ServiceName: "registry-service",
		Version:     "1.0.0",
		Endpoints: []nexus.Endpoint{
			{Path: "/entities", Method: "POST", AuthRequired: true},
		},
	}
	nexus.RegisterDiscovery(r, def)

	nexus.StartGatewayHandshake(def)

	// 4. Rotte
	h := &handlers.EntityHandler{Repo: entityRepo}
	r.POST("/entities", h.Create)

	log.Println("Avvio Registry Service sulla porta 8080...")
	r.Run(":8080")
}
