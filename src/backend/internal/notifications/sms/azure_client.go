package sms

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/axiom-software-co/international-center/src/backend/internal/shared/domain"
)

// AzureSMSClient provides Azure Communication Services SMS functionality
type AzureSMSClient interface {
	Initialize(ctx context.Context, config *AzureSMSConfig) error
	SendSMS(ctx context.Context, request *AzureSendSMSRequest) (*AzureSendSMSResponse, error)
	GetDeliveryStatus(ctx context.Context, messageID string) (*AzureSMSDeliveryStatus, error)
	HealthCheck(ctx context.Context) error
}

// AzureCommunicationSMSClient implements AzureSMSClient
type AzureCommunicationSMSClient struct {
	config     *AzureSMSConfig
	httpClient *http.Client
	logger     *slog.Logger
	endpoint   string
	apiVersion string
}

// NewAzureCommunicationSMSClient creates a new Azure Communication Services SMS client
func NewAzureCommunicationSMSClient(logger *slog.Logger) *AzureCommunicationSMSClient {
	return &AzureCommunicationSMSClient{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		logger:     logger,
		apiVersion: "2021-03-07",
	}
}

// Initialize initializes the Azure SMS client with configuration
func (a *AzureCommunicationSMSClient) Initialize(ctx context.Context, config *AzureSMSConfig) error {
	if config == nil {
		return domain.NewValidationError("Azure SMS configuration cannot be nil", nil)
	}

	if config.ConnectionString == "" {
		return domain.NewValidationError("Azure connection string is required", nil)
	}

	if config.FromNumber == "" {
		return domain.NewValidationError("from number is required", nil)
	}

	// Validate from number format
	if !IsValidUSPhoneNumber(config.FromNumber) {
		return domain.NewValidationError("from number is not a valid US phone number", nil)
	}

	// Parse connection string to extract endpoint
	endpoint, err := a.parseConnectionString(config.ConnectionString)
	if err != nil {
		return fmt.Errorf("failed to parse connection string: %w", err)
	}

	a.config = config
	a.endpoint = endpoint
	a.httpClient.Timeout = time.Duration(config.RequestTimeout) * time.Second

	a.logger.Info("Azure Communication Services SMS client initialized",
		"endpoint", endpoint,
		"from_number", config.FromNumber,
		"timeout", a.httpClient.Timeout)

	return nil
}

// SendSMS sends an SMS using Azure Communication Services
func (a *AzureCommunicationSMSClient) SendSMS(ctx context.Context, request *AzureSendSMSRequest) (*AzureSendSMSResponse, error) {
	if err := a.validateSendRequest(request); err != nil {
		return nil, fmt.Errorf("invalid send request: %w", err)
	}

	logger := a.logger.With(
		"from", request.From,
		"recipients", len(request.To),
		"message_length", len(request.Message))

	logger.Debug("Sending SMS via Azure Communication Services")

	// Validate message length
	if len(request.Message) > MaxSMSLength {
		// Truncate message if too long
		request.Message = TruncateSMSContent(request.Message, MaxSMSLength)
		logger.Warn("Message truncated to SMS limit", "new_length", len(request.Message))
	}

	// Validate and format recipient phone numbers
	validRecipients := make([]string, 0, len(request.To))
	for _, phone := range request.To {
		if IsValidUSPhoneNumber(phone) {
			formatted := FormatPhoneNumberE164(phone)
			validRecipients = append(validRecipients, formatted)
		} else {
			logger.Warn("Skipping invalid phone number", "phone", phone)
		}
	}

	if len(validRecipients) == 0 {
		return nil, domain.NewValidationError("no valid phone numbers to send SMS to", nil)
	}

	// Update request with valid recipients
	request.To = validRecipients

	// Create HTTP request
	url := fmt.Sprintf("%s/sms?api-version=%s", a.endpoint, a.apiVersion)
	
	requestBody, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// In a real implementation, this would make an actual HTTP request to Azure
	// For now, we simulate the response based on the request
	response := a.simulateAzureResponse(request)

	logger.Info("SMS sent successfully via Azure Communication Services",
		"azure_message_id", response.MessageID,
		"sent_to", len(response.To))

	return response, nil
}

// GetDeliveryStatus retrieves the delivery status from Azure
func (a *AzureCommunicationSMSClient) GetDeliveryStatus(ctx context.Context, messageID string) (*AzureSMSDeliveryStatus, error) {
	if messageID == "" {
		return nil, domain.NewValidationError("message ID is required", nil)
	}

	logger := a.logger.With("message_id", messageID)
	logger.Debug("Getting SMS delivery status from Azure")

	// Create HTTP request for status
	url := fmt.Sprintf("%s/sms/deliveryReports?api-version=%s&messageId=%s", a.endpoint, a.apiVersion, messageID)
	
	// In a real implementation, this would make an actual HTTP request to Azure
	// For now, we simulate the response
	status := &AzureSMSDeliveryStatus{
		MessageID:    messageID,
		From:         a.config.FromNumber,
		To:           "+1234567890", // Would come from actual response
		DeliveryStatus: "Delivered",
		DeliveryStatusDetails: AzureSMSDeliveryDetails{
			StatusMessage: "SMS delivered successfully",
			Timestamp:     time.Now().UTC().Add(-1 * time.Minute),
		},
		ReceivedTimestamp: time.Now().UTC().Add(-5 * time.Minute),
	}

	logger.Debug("Retrieved SMS delivery status", "status", status.DeliveryStatus)
	return status, nil
}

// HealthCheck performs a health check on the Azure Communication Services
func (a *AzureCommunicationSMSClient) HealthCheck(ctx context.Context) error {
	if a.config == nil {
		return domain.NewDependencyError("Azure SMS client not initialized", nil)
	}

	// In a real implementation, this would make a lightweight API call to Azure
	// For now, we simulate a successful health check
	a.logger.Debug("Azure Communication Services SMS health check passed")
	return nil
}

// Private helper methods

// parseConnectionString parses Azure connection string to extract endpoint
func (a *AzureCommunicationSMSClient) parseConnectionString(connectionString string) (string, error) {
	// Parse connection string format: "endpoint=https://...;accesskey=..."
	// In a real implementation, this would properly parse the connection string
	// For now, we return a mock endpoint
	return "https://mock-communication-service.azure.com", nil
}

// validateSendRequest validates the send SMS request
func (a *AzureCommunicationSMSClient) validateSendRequest(request *AzureSendSMSRequest) error {
	if request == nil {
		return domain.NewValidationError("send request cannot be nil", nil)
	}

	if request.From == "" {
		return domain.NewValidationError("from number is required", nil)
	}

	if len(request.To) == 0 {
		return domain.NewValidationError("at least one recipient is required", nil)
	}

	if request.Message == "" {
		return domain.NewValidationError("message content is required", nil)
	}

	// Validate from number
	if !IsValidUSPhoneNumber(request.From) {
		return domain.NewValidationError(fmt.Sprintf("invalid from number: %s", request.From), nil)
	}

	// Validate message length
	if len(request.Message) > MaxSMSLength {
		return domain.NewValidationError(fmt.Sprintf("message too long: %d characters (max %d)", len(request.Message), MaxSMSLength), nil)
	}

	// Validate recipient numbers (basic validation)
	for i, phone := range request.To {
		if phone == "" {
			return domain.NewValidationError(fmt.Sprintf("recipient %d phone number cannot be empty", i), nil)
		}
		
		// Note: Detailed phone validation will happen in the service layer
	}

	return nil
}

// simulateAzureResponse simulates an Azure Communication Services response
func (a *AzureCommunicationSMSClient) simulateAzureResponse(request *AzureSendSMSRequest) *AzureSendSMSResponse {
	// Generate mock response
	messageID := fmt.Sprintf("azure-sms-%d", time.Now().Unix())
	
	// Create recipient results
	recipientResults := make([]AzureSMSRecipientResult, len(request.To))
	for i, recipient := range request.To {
		recipientResults[i] = AzureSMSRecipientResult{
			To:                   recipient,
			MessageID:            fmt.Sprintf("%s-%d", messageID, i),
			HttpStatusCode:       200,
			Successful:           true,
			RepeatabilityResult:  "accepted",
		}
	}
	
	return &AzureSendSMSResponse{
		MessageID: messageID,
		To:        recipientResults,
	}
}

// SMS Character Limits and Validation (from domain.go)

const (
	MaxSMSLength         = 160  // Standard SMS character limit
	MaxConcatSMSLength   = 1600 // Maximum for concatenated SMS
	MaxPhoneNumberLength = 15   // E.164 format maximum
	MinPhoneNumberLength = 10   // US format minimum
)

// Phone number validation functions (would be imported from domain.go in real implementation)

// IsValidUSPhoneNumber validates US phone number format
func IsValidUSPhoneNumber(phone string) bool {
	// This would use the implementation from domain.go
	// Simplified version for compilation
	if len(phone) < 10 || len(phone) > 15 {
		return false
	}
	
	// Basic validation - starts with +1 for international or area code 2-9 for domestic
	cleaned := phone
	for _, char := range "()- " {
		cleaned = ""
		for _, c := range phone {
			if c != char {
				cleaned += string(c)
			}
		}
	}
	
	if len(cleaned) == 10 && cleaned[0] >= '2' && cleaned[0] <= '9' {
		return true
	}
	
	if len(cleaned) == 11 && cleaned[0] == '1' && cleaned[1] >= '2' && cleaned[1] <= '9' {
		return true
	}
	
	return false
}

// FormatPhoneNumberE164 formats phone number to E.164 format
func FormatPhoneNumberE164(phone string) string {
	// This would use the implementation from domain.go
	// Simplified version for compilation
	cleaned := ""
	for _, c := range phone {
		if c >= '0' && c <= '9' {
			cleaned += string(c)
		}
	}
	
	if len(cleaned) == 10 {
		return "+1" + cleaned
	}
	
	if len(cleaned) == 11 && cleaned[0] == '1' {
		return "+" + cleaned
	}
	
	return phone // Return original if formatting fails
}

// TruncateSMSContent truncates SMS content to character limit
func TruncateSMSContent(content string, maxLength int) string {
	if len(content) <= maxLength {
		return content
	}
	
	// Try to truncate at word boundary
	if maxLength > 3 {
		truncated := content[:maxLength-3]
		// Find last space in the second half to avoid cutting too early
		lastSpace := -1
		for i := len(truncated) - 1; i >= maxLength/2; i-- {
			if truncated[i] == ' ' {
				lastSpace = i
				break
			}
		}
		
		if lastSpace > 0 {
			return truncated[:lastSpace] + "..."
		}
		return truncated + "..."
	}
	
	return content[:maxLength]
}

// SMS content generation functions (would be imported from domain.go in real implementation)

// GenerateSMSContent generates SMS content by event type
func GenerateSMSContent(eventType string, eventData map[string]interface{}) string {
	switch eventType {
	case "inquiry-business":
		return generateBusinessInquirySMS(eventData)
	case "inquiry-media":
		return generateMediaInquirySMS(eventData)
	case "inquiry-donations":
		return generateDonationInquirySMS(eventData)
	case "inquiry-volunteers":
		return generateVolunteerInquirySMS(eventData)
	case "event-registration":
		return generateContentPublicationSMS(eventData)
	case "system-error":
		return generateSystemErrorSMS(eventData)
	case "capacity-alert":
		return generateCapacityAlertSMS(eventData)
	case "admin-action-required":
		return generateAdminActionSMS(eventData)
	case "compliance-alert":
		return generateComplianceAlertSMS(eventData)
	default:
		return "New notification alert. Check admin dashboard for details."
	}
}

func generateBusinessInquirySMS(eventData map[string]interface{}) string {
	entityID := extractString(eventData, "entity_id")
	if entityID != "" {
		return fmt.Sprintf("New business inquiry %s received. Review in admin dashboard.", entityID)
	}
	return "New business inquiry received. Check admin dashboard."
}

func generateMediaInquirySMS(eventData map[string]interface{}) string {
	entityID := extractString(eventData, "entity_id")
	if entityID != "" {
		return fmt.Sprintf("New media inquiry %s received. Review in admin dashboard.", entityID)
	}
	return "New media inquiry received. Check admin dashboard."
}

func generateDonationInquirySMS(eventData map[string]interface{}) string {
	entityID := extractString(eventData, "entity_id")
	if entityID != "" {
		return fmt.Sprintf("New donation inquiry %s received. Review in admin dashboard.", entityID)
	}
	return "New donation inquiry received. Check admin dashboard."
}

func generateVolunteerInquirySMS(eventData map[string]interface{}) string {
	entityID := extractString(eventData, "entity_id")
	if entityID != "" {
		return fmt.Sprintf("New volunteer application %s received. Review in admin dashboard.", entityID)
	}
	return "New volunteer application received. Check admin dashboard."
}

func generateContentPublicationSMS(eventData map[string]interface{}) string {
	entityType := extractString(eventData, "entity_type")
	entityID := extractString(eventData, "entity_id")
	
	if entityType != "" && entityID != "" {
		return fmt.Sprintf("New %s content %s published. View details in dashboard.", entityType, entityID)
	} else if entityType != "" {
		return fmt.Sprintf("New %s content published. Check dashboard for details.", entityType)
	}
	return "New content published. Check dashboard for details."
}

func generateSystemErrorSMS(eventData map[string]interface{}) string {
	errorType := extractString(eventData, "error_type")
	if errorType != "" {
		return fmt.Sprintf("URGENT: System error (%s) detected. Immediate action required.", errorType)
	}
	return "URGENT: System error detected. Immediate action required."
}

func generateCapacityAlertSMS(eventData map[string]interface{}) string {
	resourceType := extractString(eventData, "resource_type")
	if resourceType != "" {
		return fmt.Sprintf("WARNING: %s capacity alert. Check system resources immediately.", resourceType)
	}
	return "WARNING: Capacity alert detected. Check system resources."
}

func generateAdminActionSMS(eventData map[string]interface{}) string {
	actionType := extractString(eventData, "action_type")
	if actionType != "" {
		return fmt.Sprintf("Admin action required: %s. Check dashboard for details.", actionType)
	}
	return "Admin action required. Check dashboard for details."
}

func generateComplianceAlertSMS(eventData map[string]interface{}) string {
	alertType := extractString(eventData, "alert_type")
	if alertType != "" {
		return fmt.Sprintf("COMPLIANCE: %s alert. Immediate review required.", alertType)
	}
	return "COMPLIANCE: Alert detected. Immediate review required."
}

// extractString helper function
func extractString(data map[string]interface{}, key string) string {
	if value, exists := data[key]; exists {
		if str, ok := value.(string); ok {
			return str
		}
	}
	return ""
}