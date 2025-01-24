package handler

const (
	ErrUnauthorized       = "Unauthorized"
    ErrInvalidUserToken   = "Invalid user token"
    ErrMissingOrInvalidToken = "Missing or invalid token"
    ErrInvalidToken       = "Invalid token"
    ErrInvalidClaims      = "Invalid claims"
    	ErrInvalidMaterialID       = "invalid material ID format"
	ErrInvalidMaterialData     = "invalid material data"
	ErrForbiddenModify         = "forbidden to modify this material"
	ErrFailedUpdateMaterial    = "failed to update material"
	ErrInvalidID               = "invalid ID"
	ErrFailedDeleteMaterial    = "failed to delete material"
	ErrFailedRetrieveMaterials = "failed to retrieve materials"
	ErrFailedCreateMaterial    = "failed to create material"
	ErrMaterialNotFound        = "material not found"

)