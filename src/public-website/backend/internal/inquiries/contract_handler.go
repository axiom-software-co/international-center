package inquiries

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/axiom-software-co/international-center/src/backend/internal/inquiries/media"
	"github.com/axiom-software-co/international-center/src/backend/internal/shared/domain"
	"github.com/axiom-software-co/international-center/src/backend/internal/contracts/admin"
	"github.com/google/uuid"
	openapi_types "github.com/oapi-codegen/runtime/types"
)

// ContractCompliantInquiryHandler implements the generated admin API interface for inquiry management
type ContractCompliantInquiryHandler struct {
	mediaService *media.MediaService
}

// NewContractCompliantInquiryHandler creates a new contract-compliant inquiry handler
func NewContractCompliantInquiryHandler(mediaService *media.MediaService) *ContractCompliantInquiryHandler {
	return &ContractCompliantInquiryHandler{
		mediaService: mediaService,
	}
}

// GetInquiries implements GET /inquiries - Get all inquiries
func (h *ContractCompliantInquiryHandler) GetInquiries(w http.ResponseWriter, r *http.Request, params admin.GetInquiriesParams) {
	ctx := r.Context()
	
	// Extract user ID from context (would come from authentication middleware)
	userID := r.Header.Get("X-User-ID")
	
	// Add correlation context
	correlationCtx := domain.FromContext(ctx)
	correlationCtx.SetUserContext(userID, "admin-api-1.0.0")
	ctx = correlationCtx.ToContext(ctx)

	// Convert parameters to internal format
	listParams := media.ListInquiriesParams{
		Page:  1,
		Limit: 20,
	}
	
	if params.Page != nil {
		listParams.Page = int(*params.Page)
	}
	if params.Limit != nil {
		listParams.Limit = int(*params.Limit)
	}
	if params.Search != nil {
		listParams.Search = *params.Search
	}
	
	// Map inquiry type filter
	if params.InquiryType != nil {
		switch *params.InquiryType {
		case admin.GetInquiriesParamsInquiryTypeMedia:
			listParams.InquiryType = "media"
		case admin.GetInquiriesParamsInquiryTypeBusiness:
			listParams.InquiryType = "business"
		case admin.GetInquiriesParamsInquiryTypeDonation:
			listParams.InquiryType = "donation"
		case admin.GetInquiriesParamsInquiryTypeVolunteer:
			listParams.InquiryType = "volunteer"
		}
	}

	// Map status filter
	if params.Status != nil {
		switch *params.Status {
		case admin.GetInquiriesParamsStatusPending:
			listParams.Status = "new"
		case admin.GetInquiriesParamsStatusInProgress:
			listParams.Status = "in_progress"
		case admin.GetInquiriesParamsStatusCompleted:
			listParams.Status = "resolved"
		case admin.GetInquiriesParamsStatusClosed:
			listParams.Status = "closed"
		}
	}

	// Call media service (for now, we'll focus on media inquiries)
	inquiries, pagination, err := h.mediaService.AdminListInquiries(ctx, listParams, userID)
	if err != nil {
		h.handleError(w, r, err, correlationCtx.CorrelationID)
		return
	}

	// Convert to contract-compliant types
	contractInquiries := make([]admin.Inquiry, len(inquiries))
	for i, inq := range inquiries {
		contractInquiries[i] = h.convertMediaInquiryToContract(inq)
	}

	// Build contract-compliant response
	response := struct {
		Data       []admin.Inquiry        `json:"data"`
		Pagination admin.PaginationInfo   `json:"pagination"`
	}{
		Data: contractInquiries,
		Pagination: admin.PaginationInfo{
			CurrentPage:  pagination.CurrentPage,
			TotalPages:   pagination.TotalPages,
			TotalItems:   pagination.TotalItems,
			ItemsPerPage: pagination.ItemsPerPage,
			HasNext:      pagination.HasNext,
			HasPrevious:  pagination.HasPrevious,
		},
	}

	h.writeContractResponse(w, http.StatusOK, response, correlationCtx.CorrelationID)
}

// GetInquiryById implements GET /inquiries/{id} - Get inquiry by ID
func (h *ContractCompliantInquiryHandler) GetInquiryById(w http.ResponseWriter, r *http.Request, id openapi_types.UUID) {
	ctx := r.Context()
	
	// Extract user ID from context
	userID := r.Header.Get("X-User-ID")
	
	// Add correlation context
	correlationCtx := domain.FromContext(ctx)
	correlationCtx.SetUserContext(userID, "admin-api-1.0.0")
	ctx = correlationCtx.ToContext(ctx)

	// Call media service - note the method returns a MediaInquiry directly
	mediaInquiry, err := h.mediaService.AdminGetInquiry(ctx, id.String(), userID)
	if err != nil {
		h.handleError(w, r, err, correlationCtx.CorrelationID)
		return
	}

	// Convert MediaInquiry to contract-compliant type
	contractInquiry := h.convertMediaInquiryToContract(h.mediaService.ConvertToContract(mediaInquiry))

	// Build contract-compliant response
	response := struct {
		Data admin.Inquiry `json:"data"`
	}{
		Data: contractInquiry,
	}

	h.writeContractResponse(w, http.StatusOK, response, correlationCtx.CorrelationID)
}

// UpdateInquiryStatus implements PUT /inquiries/{id} - Update inquiry status
func (h *ContractCompliantInquiryHandler) UpdateInquiryStatus(w http.ResponseWriter, r *http.Request, id openapi_types.UUID) {
	ctx := r.Context()
	
	// Extract user ID from context
	userID := r.Header.Get("X-User-ID")
	
	// Add correlation context
	correlationCtx := domain.FromContext(ctx)
	correlationCtx.SetUserContext(userID, "admin-api-1.0.0")
	ctx = correlationCtx.ToContext(ctx)

	// Parse request body
	var request admin.UpdateInquiryStatusJSONBody
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		h.handleError(w, r, domain.NewValidationError("Invalid request body: "+err.Error()), correlationCtx.CorrelationID)
		return
	}

	// Map contract status to internal status
	var internalStatus string
	switch request.Status {
	case admin.UpdateInquiryStatusJSONBodyStatusPending:
		internalStatus = "new"
	case admin.UpdateInquiryStatusJSONBodyStatusInProgress:
		internalStatus = "in_progress"
	case admin.UpdateInquiryStatusJSONBodyStatusCompleted:
		internalStatus = "resolved"
	case admin.UpdateInquiryStatusJSONBodyStatusClosed:
		internalStatus = "closed"
	default:
		h.handleError(w, r, domain.NewValidationError("Invalid status value"), correlationCtx.CorrelationID)
		return
	}

	// Create internal update request
	updateReq := media.AdminUpdateInquiryStatusRequest{
		Status: media.InquiryStatus(internalStatus),
		Notes:  request.Notes,
	}
	
	if request.AssignedTo != nil {
		assignedToStr := request.AssignedTo.String()
		updateReq.AssignedTo = &assignedToStr
	}

	// Call media service
	inquiry, err := h.mediaService.AdminUpdateInquiryStatus(ctx, id.String(), updateReq, userID)
	if err != nil {
		h.handleError(w, r, err, correlationCtx.CorrelationID)
		return
	}

	// Convert to contract-compliant type
	contractInquiry := h.convertMediaInquiryToContract(*inquiry)

	// Build contract-compliant response using UpdatedResponse format
	response := admin.UpdatedResponse{
		Success:       true,
		Message:       "Inquiry status updated successfully",
		Data:          map[string]interface{}{"inquiry": contractInquiry},
		Timestamp:     time.Now().UTC(),
		CorrelationId: openapi_types.UUID(uuid.MustParse(correlationCtx.CorrelationID)),
	}

	h.writeContractResponse(w, http.StatusOK, response, correlationCtx.CorrelationID)
}

// convertMediaInquiryToContract converts media inquiry type to contract-compliant type
func (h *ContractCompliantInquiryHandler) convertMediaInquiryToContract(inq media.Inquiry) admin.Inquiry {
	// Map status (inq.Status is already a string from media.Inquiry)
	var contractStatus admin.InquiryStatus
	switch inq.Status {
	case "new":
		contractStatus = admin.InquiryStatusPending
	case "acknowledged", "in_progress":
		contractStatus = admin.InquiryStatusInProgress
	case "resolved":
		contractStatus = admin.InquiryStatusCompleted
	case "closed":
		contractStatus = admin.InquiryStatusClosed
	default:
		contractStatus = admin.InquiryStatusPending
	}

	// Map inquiry type
	var contractInquiryType admin.InquiryInquiryType = admin.InquiryInquiryTypeMedia

	// Convert to contract format (using media.Inquiry fields)
	inquiryUUID, _ := uuid.Parse(inq.ID) // Parse string ID to UUID
	contractInquiry := admin.Inquiry{
		InquiryId:      openapi_types.UUID(inquiryUUID),
		InquiryType:    contractInquiryType,
		Status:         contractStatus,
		SubmitterName:  inq.SubmitterName,
		SubmitterEmail: openapi_types.Email(inq.SubmitterEmail),
		Subject:        inq.Subject,
		Message:        inq.Message,
		SubmittedOn:    inq.SubmittedOn,
		LastUpdated:    &inq.ModifiedOn,
		Notes:          inq.Notes,
	}

	// Handle assigned to (optional)
	if inq.AssignedTo != nil {
		assignedToUUID, err := uuid.Parse(*inq.AssignedTo)
		if err == nil {
			assignedToOpenAPIUUID := openapi_types.UUID(assignedToUUID)
			contractInquiry.AssignedTo = &assignedToOpenAPIUUID
		}
	}

	return contractInquiry
}

// handleError handles errors in a contract-compliant way
func (h *ContractCompliantInquiryHandler) handleError(w http.ResponseWriter, r *http.Request, err error, correlationID string) {
	var statusCode int
	var errorCode, message string

	// Determine error type using domain functions
	switch {
	case domain.IsValidationError(err):
		statusCode = http.StatusBadRequest
		errorCode = "BAD_REQUEST"
		message = err.Error()
	case domain.IsNotFoundError(err):
		statusCode = http.StatusNotFound
		errorCode = "NOT_FOUND"
		message = err.Error()
	case domain.IsUnauthorizedError(err):
		statusCode = http.StatusUnauthorized
		errorCode = "UNAUTHORIZED"
		message = err.Error()
	case domain.IsForbiddenError(err):
		statusCode = http.StatusForbidden
		errorCode = "FORBIDDEN"
		message = err.Error()
	case domain.IsTimeoutError(err):
		statusCode = http.StatusRequestTimeout
		errorCode = "TIMEOUT"
		message = err.Error()
	default:
		statusCode = http.StatusInternalServerError
		errorCode = "INTERNAL_ERROR"
		message = "An internal error occurred"
	}

	// Create standardized error response
	errorResponse := map[string]interface{}{
		"error": map[string]interface{}{
			"code":           errorCode,
			"message":        message,
			"correlation_id": correlationID,
			"timestamp":      time.Now().UTC().Format(time.RFC3339),
		},
	}

	// Set response headers
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Correlation-ID", correlationID)
	w.WriteHeader(statusCode)
	
	// Write error response
	json.NewEncoder(w).Encode(errorResponse)
}

// writeContractResponse writes a contract-compliant JSON response
func (h *ContractCompliantInquiryHandler) writeContractResponse(w http.ResponseWriter, statusCode int, data interface{}, correlationID string) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Correlation-ID", correlationID)
	
	// Add security headers
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("X-Frame-Options", "DENY")
	w.Header().Set("X-XSS-Protection", "1; mode=block")
	
	w.WriteHeader(statusCode)
	
	if data != nil {
		json.NewEncoder(w).Encode(data)
	}
}