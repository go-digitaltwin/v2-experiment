package coverage

import "coverage/tenant"

// Account exercises fields with types from another package in the same Go
// module. The import path is a module-relative path, not a stdlib path.
type Account struct {
	ID    string
	Owner tenant.ID
}

// AccountContainers exercises container fields whose element or key types come
// from another package in the same module.
type AccountContainers struct {
	ID            string
	Members       []tenant.ID
	RolesByTenant map[tenant.ID]string
	TenantsByRole map[string]tenant.ID
}
