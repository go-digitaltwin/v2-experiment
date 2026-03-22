package coverage

import "github.com/google/uuid"

// Asset exercises fields with types from an external module (outside stdlib
// and outside the current module).
type Asset struct {
	ID            string
	CorrelationID uuid.UUID
}

// AssetContainers exercises container fields whose element or key types come
// from an external module.
type AssetContainers struct {
	ID           string
	History      []uuid.UUID
	LabelsByUUID map[uuid.UUID]string
	UUIDsByLabel map[string]uuid.UUID
}
