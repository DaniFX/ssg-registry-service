package models

import "github.com/DaniFX/ssg-nexus-sdk/pkg/nexus/repository"

// Entity rappresenta un'anagrafica polimorfica (Socio, Cliente, etc.)
type Entity struct {
	repository.NexusDoc                        // Includiamo i metadati ERP (ID, CreatedAt, etc.)
	Type                string                 `json:"type" firestore:"type"` // PERSON | ORGANIZATION
	SubTypes            []string               `json:"subTypes" firestore:"subTypes"`
	Status              string                 `json:"status" firestore:"status"`
	CoreData            map[string]interface{} `json:"coreData" firestore:"coreData"`
	ExtData             map[string]interface{} `json:"extData" firestore:"extData"`
}
