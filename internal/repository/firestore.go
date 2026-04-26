package repository

import (
	"context"
	"log"
	"os"

	"cloud.google.com/go/firestore"
)

// InitFirestore crea la connessione al database del progetto GCP
func InitFirestore(ctx context.Context) *firestore.Client {
	// In Cloud Run, questa variabile d'ambiente può essere iniettata o dedotta
	projectID := os.Getenv("GCP_PROJECT_ID")
	if projectID == "" {
		log.Fatalf("ERRORE CRITICO: Variabile di ambiente GCP_PROJECT_ID non impostata")
	}

	client, err := firestore.NewClient(ctx, projectID)
	if err != nil {
		log.Fatalf("Impossibile connettersi a Firestore: %v", err)
	}

	return client
}
