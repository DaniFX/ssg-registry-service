package models

import "github.com/DaniFX/ssg-nexus-sdk/pkg/nexus/repository"

// Address rappresenta un indirizzo fisico
type Address struct {
	Street  string `json:"street,omitempty"  firestore:"street,omitempty"`
	City    string `json:"city,omitempty"    firestore:"city,omitempty"`
	Zip     string `json:"zip,omitempty"     firestore:"zip,omitempty"`
	Country string `json:"country,omitempty" firestore:"country,omitempty"`
}

// CoreData contiene i dati anagrafici principali dell'entità
type CoreData struct {
	DisplayName string   `json:"displayName"          firestore:"displayName"`
	Email       string   `json:"email"                firestore:"email"`
	Phone       string   `json:"phone,omitempty"      firestore:"phone,omitempty"`
	TaxCode     string   `json:"taxCode,omitempty"    firestore:"taxCode,omitempty"`
	VatNumber   string   `json:"vatNumber,omitempty" firestore:"vatNumber,omitempty"`
	Address     *Address `json:"address,omitempty"   firestore:"address,omitempty"`
}

// Entity rappresenta un'anagrafica polimorfica (Socio, Cliente, Fornitore, etc.)
// Allineata con NexusEntity in contracts/schemas.json v1.2.0
type Entity struct {
	repository.NexusDoc
	Type     string                 `json:"type"              firestore:"type"`     // PERSON | ORGANIZATION
	SubType  string                 `json:"subType"           firestore:"subType"`  // MEMBER | CUSTOMER | SUPPLIER | EMPLOYEE | PARTNER
	Status   string                 `json:"status"            firestore:"status"`   // ACTIVE | INACTIVE | SUSPENDED
	CoreData CoreData               `json:"coreData"          firestore:"coreData"`
	Metadata map[string]interface{} `json:"metadata,omitempty" firestore:"metadata,omitempty"`
}

// CreateEntityRequest è il payload accettato in POST /entities
// Non include i campi NexusDoc (iniettati dall'SDK)
type CreateEntityRequest struct {
	Type     string                 `json:"type"              binding:"required"`
	SubType  string                 `json:"subType"           binding:"required"`
	Status   string                 `json:"status"`
	CoreData CoreData               `json:"coreData"          binding:"required"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// UpdateEntityRequest è il payload accettato in PATCH /entities/:id
// Tutti i campi sono opzionali (partial update)
type UpdateEntityRequest struct {
	SubType  string                 `json:"subType,omitempty"`
	Status   string                 `json:"status,omitempty"`
	CoreData *CoreData              `json:"coreData,omitempty"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}
