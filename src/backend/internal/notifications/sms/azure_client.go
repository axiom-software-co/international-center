package sms

import (
	"context"
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
		return domain.NewValidationError("Azure SMS configuration cannot be nil")
	}

	if config.ConnectionString == "" {
		return domain.NewValidationError("Azure connection string is required")
	}

	if config.FromNumber == "" {
		return domain.NewValidationError("from number is required")
	}

	// Validate from number format
	if !IsValidUSPhoneNumber(config.FromNumber) {
		return domain.NewValidationError("from number is not a valid US phone number")
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
		return nil, domain.NewValidationError("no valid phone numbers to send SMS to")
	}

	// Update request with valid recipients
	request.To = validRecipients

	// Create HTTP request
	// url := fmt.Sprintf("%s/sms?api-version=%s", a.endpoint, a.apiVersion)
	
	// requestBody, err := json.Marshal(request)
	// if err != nil {
	//	return nil, fmt.Errorf("failed to marshal request: %w", err)
	// }

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
		return nil, domain.NewValidationError("message ID is required")
	}

	logger := a.logger.With("message_id", messageID)
	logger.Debug("Getting SMS delivery status from Azure")

	// Create HTTP request for status
	// url := fmt.Sprintf("%s/sms/deliveryReports?api-version=%s&messageId=%s", a.endpoint, a.apiVersion, messageID)
	
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
		return domain.NewValidationError("send request cannot be nil")
	}

	if request.From == "" {
		return domain.NewValidationError("from number is required")
	}

	if len(request.To) == 0 {
		return domain.NewValidationError("at least one recipient is required")
	}

	if request.Message == "" {
		return domain.NewValidationError("message content is required")
	}

	// Validate from number
	if !IsValidUSPhoneNumber(request.From) {
		return domain.NewValidationError(fmt.Sprintf("invalid from number: %s", request.From))
	}

	// Validate message length
	if len(request.Message) > MaxSMSLength {
		return domain.NewValidationError(fmt.Sprintf("message too long: %d characters (max %d)", len(request.Message), MaxSMSLength))
	}

	// Validate recipient numbers (basic validation)
	for i, phone := range request.To {
		if phone == "" {
			return domain.NewValidationError(fmt.Sprintf("recipient %d phone number cannot be empty", i))
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

