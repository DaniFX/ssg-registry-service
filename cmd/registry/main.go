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
	// 1. Setup Firestore
	ctx := context.Background()
	firestoreClient := repository.InitFirestore(ctx)
	defer firestoreClient.Close()

	entityRepo := nexusRepo.NewRepository(firestoreClient, "entities")

	r := gin.Default()

	// 2. Service Discovery
	def := nexus.ServiceDefinition{
		ServiceName: "registry-service",
		Version:     "1.1.0",
		Endpoints: []nexus.Endpoint{
			{Path: "/api/v1/registry/entities",     Method: "GET",    AuthRequired: true, Summary: "Lista anagrafiche (Navigator)"},
			{Path: "/api/v1/registry/entities",     Method: "POST",   AuthRequired: true, Summary: "Crea anagrafica"},
			{Path: "/api/v1/registry/entities/:id", Method: "GET",    AuthRequired: true, Summary: "Dettaglio anagrafica"},
			{Path: "/api/v1/registry/entities/:id", Method: "PATCH",  AuthRequired: true, Summary: "Aggiorna anagrafica"},
			{Path: "/api/v1/registry/entities/:id", Method: "DELETE", AuthRequired: true, Summary: "Soft delete anagrafica"},
		},
	}
	nexus.RegisterDiscovery(r, def)
	nexus.StartGatewayHandshake(def)

	// 3. Rotte protette
	h := &handlers.EntityHandler{Repo: entityRepo}

	api := r.Group("/api/v1/registry")
	api.Use(nexus.Guard())
	{
		api.GET("/entities",     h.List)
		api.POST("/entities",    h.Create)
		api.GET("/entities/:id", h.GetByID)
		api.PATCH("/entities/:id", h.Update)
		api.DELETE("/entities/:id", h.Delete)
	}

	// PORT da variabile d'ambiente (fix issue aperto)
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Registry Service avviato sulla porta %s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("Errore critico del server: %v", err)
	}
}
