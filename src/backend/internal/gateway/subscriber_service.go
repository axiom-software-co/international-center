package gateway

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/axiom-software-co/international-center/src/backend/internal/notifications"
	"github.com/axiom-software-co/international-center/src/backend/internal/shared/domain"
	"github.com/google/uuid"
)

// Gateway-specific request/response types using notification domain models
type CreateSubscriberRequest struct {
	SubscriberName       string                              `json:"subscriber_name"`
	Email                string                              `json:"email"`
	Phone                *string                             `json:"phone,omitempty"`
	EventTypes           []notifications.EventType           `json:"event_types"`
	NotificationMethods  []notifications.NotificationMethod  `json:"notification_methods"`
	NotificationSchedule notifications.NotificationSchedule  `json:"notification_schedule"`
	PriorityThreshold    notifications.PriorityThreshold     `json:"priority_threshold"`
	Notes                *string                             `json:"notes,omitempty"`
	CreatedBy            string                              `json:"created_by"`
}

type UpdateSubscriberRequest struct {
	Status               *notifications.SubscriberStatus      `json:"status,omitempty"`
	SubscriberName       *string                             `json:"subscriber_name,omitempty"`
	Email                *string                             `json:"email,omitempty"`
	Phone                *string                             `json:"phone,omitempty"`
	EventTypes           []notifications.EventType           `json:"event_types,omitempty"`
	NotificationMethods  []notifications.NotificationMethod  `json:"notification_methods,omitempty"`
	NotificationSchedule *notifications.NotificationSchedule `json:"notification_schedule,omitempty"`
	PriorityThreshold    *notifications.PriorityThreshold    `json:"priority_threshold,omitempty"`
	Notes                *string                             `json:"notes,omitempty"`
	UpdatedBy            string                              `json:"updated_by"`
}

// SubscriberService interface using notification domain types
type SubscriberService interface {
	CreateSubscriber(ctx context.Context, req *CreateSubscriberRequest) (*notifications.NotificationSubscriber, error)
	GetSubscriber(ctx context.Context, subscriberID string) (*notifications.NotificationSubscriber, error)
	UpdateSubscriber(ctx context.Context, subscriberID string, req *UpdateSubscriberRequest) (*notifications.NotificationSubscriber, error)
	DeleteSubscriber(ctx context.Context, subscriberID string, deletedBy string) error
	ListSubscribers(ctx context.Context, status *notifications.SubscriberStatus, page, pageSize int) ([]*notifications.NotificationSubscriber, int, error)
	GetSubscribersByEvent(ctx context.Context, eventType notifications.EventType, priority notifications.PriorityThreshold) ([]*notifications.NotificationSubscriber, error)
	ValidateSubscriber(ctx context.Context, subscriber *notifications.NotificationSubscriber) error
}

// SubscriberRepository interface using notification domain types
type SubscriberRepository interface {
	CreateSubscriber(ctx context.Context, subscriber *notifications.NotificationSubscriber) error
	GetSubscriber(ctx context.Context, subscriberID string) (*notifications.NotificationSubscriber, error)
	UpdateSubscriber(ctx context.Context, subscriber *notifications.NotificationSubscriber) error
	DeleteSubscriber(ctx context.Context, subscriberID string, deletedBy string) error
	ListSubscribers(ctx context.Context, status *notifications.SubscriberStatus, limit, offset int) ([]*notifications.NotificationSubscriber, int, error)
	GetSubscribersByEventType(ctx context.Context, eventType notifications.EventType) ([]*notifications.NotificationSubscriber, error)
	CheckEmailExists(ctx context.Context, email string, excludeSubscriberID *string) (bool, error)
}

// DefaultSubscriberService implements SubscriberService
type DefaultSubscriberService struct {
	repository SubscriberRepository
}

// NewDefaultSubscriberService creates a new default subscriber service
func NewDefaultSubscriberService(repository SubscriberRepository) *DefaultSubscriberService {
	return &DefaultSubscriberService{
		repository: repository,
	}
}

// CreateSubscriber creates a new notification subscriber
func (s *DefaultSubscriberService) CreateSubscriber(ctx context.Context, req *CreateSubscriberRequest) (*notifications.NotificationSubscriber, error) {
	if req == nil {
		return nil, domain.NewValidationError("create request cannot be nil")
	}

	// Validate request
	if err := s.validateCreateRequest(req); err != nil {
		return nil, err
	}

	// Generate new subscriber ID
	subscriberID := uuid.New().String()

	// Create subscriber domain model
	now := time.Now().UTC()
	subscriber := &notifications.NotificationSubscriber{
		SubscriberID:         subscriberID,
		Status:               notifications.SubscriberStatusActive, // Default status
		SubscriberName:       req.SubscriberName,
		Email:                strings.ToLower(strings.TrimSpace(req.Email)),
		Phone:                req.Phone,
		EventTypes:           req.EventTypes,
		NotificationMethods:  req.NotificationMethods,
		NotificationSchedule: req.NotificationSchedule,
		PriorityThreshold:    req.PriorityThreshold,
		Notes:                req.Notes,
		CreatedAt:            now,
		UpdatedAt:            now,
		CreatedBy:            req.CreatedBy,
		UpdatedBy:            req.CreatedBy,
		IsDeleted:            false,
	}

	// Validate the subscriber model
	if err := s.ValidateSubscriber(ctx, subscriber); err != nil {
		return nil, err
	}

	// Create subscriber in repository
	if err := s.repository.CreateSubscriber(ctx, subscriber); err != nil {
		return nil, fmt.Errorf("failed to create subscriber: %w", err)
	}

	return subscriber, nil
}

// GetSubscriber retrieves a subscriber by ID
func (s *DefaultSubscriberService) GetSubscriber(ctx context.Context, subscriberID string) (*notifications.NotificationSubscriber, error) {
	if subscriberID == "" {
		return nil, domain.NewValidationError("subscriber ID cannot be empty")
	}

	// Validate UUID format
	if _, err := uuid.Parse(subscriberID); err != nil {
		return nil, domain.NewValidationError("invalid subscriber ID format")
	}

	subscriber, err := s.repository.GetSubscriber(ctx, subscriberID)
	if err != nil {
		return nil, fmt.Errorf("failed to get subscriber: %w", err)
	}

	return subscriber, nil
}

// UpdateSubscriber updates an existing subscriber
func (s *DefaultSubscriberService) UpdateSubscriber(ctx context.Context, subscriberID string, req *UpdateSubscriberRequest) (*notifications.NotificationSubscriber, error) {
	if subscriberID == "" {
		return nil, domain.NewValidationError("subscriber ID cannot be empty")
	}

	if req == nil {
		return nil, domain.NewValidationError("update request cannot be nil")
	}

	// Validate request
	if err := s.validateUpdateRequest(req); err != nil {
		return nil, err
	}

	// Get existing subscriber
	existing, err := s.repository.GetSubscriber(ctx, subscriberID)
	if err != nil {
		return nil, fmt.Errorf("failed to get existing subscriber: %w", err)
	}

	// Apply updates
	updated := s.applyUpdates(existing, req)

	// Validate the updated subscriber
	if err := s.ValidateSubscriber(ctx, updated); err != nil {
		return nil, err
	}

	// Update in repository
	if err := s.repository.UpdateSubscriber(ctx, updated); err != nil {
		return nil, fmt.Errorf("failed to update subscriber: %w", err)
	}

	return updated, nil
}

// DeleteSubscriber soft deletes a subscriber
func (s *DefaultSubscriberService) DeleteSubscriber(ctx context.Context, subscriberID string, deletedBy string) error {
	if subscriberID == "" {
		return domain.NewValidationError("subscriber ID cannot be empty")
	}

	if deletedBy == "" {
		return domain.NewValidationError("deleted by cannot be empty")
	}

	// Validate UUID format
	if _, err := uuid.Parse(subscriberID); err != nil {
		return domain.NewValidationError("invalid subscriber ID format")
	}

	// Check if subscriber exists
	_, err := s.repository.GetSubscriber(ctx, subscriberID)
	if err != nil {
		return fmt.Errorf("failed to verify subscriber exists: %w", err)
	}

	// Delete subscriber
	if err := s.repository.DeleteSubscriber(ctx, subscriberID, deletedBy); err != nil {
		return fmt.Errorf("failed to delete subscriber: %w", err)
	}

	return nil
}

// ListSubscribers retrieves subscribers with pagination and filtering
func (s *DefaultSubscriberService) ListSubscribers(ctx context.Context, status *notifications.SubscriberStatus, page, pageSize int) ([]*notifications.NotificationSubscriber, int, error) {
	// Validate pagination parameters
	if page < 1 {
		return nil, 0, domain.NewValidationError("page must be greater than 0")
	}

	if pageSize < 1 || pageSize > 100 {
		return nil, 0, domain.NewValidationError("page size must be between 1 and 100")
	}

	// Validate status filter
	if status != nil {
		if err := s.validateSubscriberStatus(*status); err != nil {
			return nil, 0, err
		}
	}

	// Calculate offset
	offset := (page - 1) * pageSize

	subscribers, total, err := s.repository.ListSubscribers(ctx, status, pageSize, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list subscribers: %w", err)
	}

	return subscribers, total, nil
}

// GetSubscribersByEvent retrieves subscribers for a specific event and priority
func (s *DefaultSubscriberService) GetSubscribersByEvent(ctx context.Context, eventType notifications.EventType, priority notifications.PriorityThreshold) ([]*notifications.NotificationSubscriber, error) {
	// Validate event type
	if err := s.validateEventType(eventType); err != nil {
		return nil, err
	}

	// Validate priority threshold
	if err := s.validatePriorityThreshold(priority); err != nil {
		return nil, err
	}

	// Get subscribers by event type
	eventSubscribers, err := s.repository.GetSubscribersByEventType(ctx, eventType)
	if err != nil {
		return nil, fmt.Errorf("failed to get subscribers by event type: %w", err)
	}

	// Filter by priority threshold
	var filteredSubscribers []*notifications.NotificationSubscriber
	for _, subscriber := range eventSubscribers {
		if s.meetsPriorityThreshold(subscriber.PriorityThreshold, priority) {
			filteredSubscribers = append(filteredSubscribers, subscriber)
		}
	}

	return filteredSubscribers, nil
}

// ValidateSubscriber validates a subscriber model
func (s *DefaultSubscriberService) ValidateSubscriber(ctx context.Context, subscriber *notifications.NotificationSubscriber) error {
	if subscriber == nil {
		return domain.NewValidationError("subscriber cannot be nil")
	}

	// Validate subscriber ID
	if subscriber.SubscriberID == "" {
		return domain.NewValidationError("subscriber ID is required")
	}

	if _, err := uuid.Parse(subscriber.SubscriberID); err != nil {
		return domain.NewValidationError("invalid subscriber ID format")
	}

	// Validate subscriber name
	if err := s.validateSubscriberName(subscriber.SubscriberName); err != nil {
		return err
	}

	// Validate email
	if err := s.validateEmail(subscriber.Email); err != nil {
		return err
	}

	// Validate phone (if provided)
	if subscriber.Phone != nil && *subscriber.Phone != "" {
		if err := s.validatePhone(*subscriber.Phone); err != nil {
			return err
		}
	}

	// Validate event types
	if err := s.validateEventTypes(subscriber.EventTypes); err != nil {
		return err
	}

	// Validate notification methods
	if err := s.validateNotificationMethods(subscriber.NotificationMethods); err != nil {
		return err
	}

	// Validate notification schedule
	if err := s.validateNotificationSchedule(subscriber.NotificationSchedule); err != nil {
		return err
	}

	// Validate priority threshold
	if err := s.validatePriorityThreshold(subscriber.PriorityThreshold); err != nil {
		return err
	}

	// Validate status
	if err := s.validateSubscriberStatus(subscriber.Status); err != nil {
		return err
	}

	// Validate notes length
	if subscriber.Notes != nil && len(*subscriber.Notes) > 1000 {
		return domain.NewValidationError("notes cannot exceed 1000 characters")
	}

	// Validate required fields
	if subscriber.CreatedBy == "" {
		return domain.NewValidationError("created by is required")
	}

	if subscriber.UpdatedBy == "" {
		return domain.NewValidationError("updated by is required")
	}

	// Validate business rules
	return s.validateBusinessRules(ctx, subscriber)
}

// Private validation methods

// validateCreateRequest validates a create subscriber request
func (s *DefaultSubscriberService) validateCreateRequest(req *CreateSubscriberRequest) error {
	if req.SubscriberName == "" {
		return domain.NewValidationError("subscriber name is required")
	}

	if len(req.SubscriberName) < 2 {
		return domain.NewValidationError("subscriber name must be at least 2 characters")
	}

	if len(req.SubscriberName) > 100 {
		return domain.NewValidationError("subscriber name cannot exceed 100 characters")
	}

	if req.Email == "" {
		return domain.NewValidationError("email is required")
	}

	if err := s.validateEmail(req.Email); err != nil {
		return err
	}

	if req.Phone != nil && *req.Phone != "" {
		if err := s.validatePhone(*req.Phone); err != nil {
			return err
		}
	}

	if len(req.EventTypes) == 0 {
		return domain.NewValidationError("at least one event type is required")
	}

	if err := s.validateEventTypes(req.EventTypes); err != nil {
		return err
	}

	if len(req.NotificationMethods) == 0 {
		return domain.NewValidationError("at least one notification method is required")
	}

	if err := s.validateNotificationMethods(req.NotificationMethods); err != nil {
		return err
	}

	if err := s.validateNotificationSchedule(req.NotificationSchedule); err != nil {
		return err
	}

	if err := s.validatePriorityThreshold(req.PriorityThreshold); err != nil {
		return err
	}

	if req.CreatedBy == "" {
		return domain.NewValidationError("created by is required")
	}

	if req.Notes != nil && len(*req.Notes) > 1000 {
		return domain.NewValidationError("notes cannot exceed 1000 characters")
	}

	return nil
}

// validateUpdateRequest validates an update subscriber request
func (s *DefaultSubscriberService) validateUpdateRequest(req *UpdateSubscriberRequest) error {
	if req.UpdatedBy == "" {
		return domain.NewValidationError("updated by is required")
	}

	if req.Status != nil {
		if err := s.validateSubscriberStatus(*req.Status); err != nil {
			return err
		}
	}

	if req.SubscriberName != nil {
		if len(*req.SubscriberName) < 2 {
			return domain.NewValidationError("subscriber name must be at least 2 characters")
		}
		if len(*req.SubscriberName) > 100 {
			return domain.NewValidationError("subscriber name cannot exceed 100 characters")
		}
	}

	if req.Email != nil && *req.Email != "" {
		if err := s.validateEmail(*req.Email); err != nil {
			return err
		}
	}

	if req.Phone != nil && *req.Phone != "" {
		if err := s.validatePhone(*req.Phone); err != nil {
			return err
		}
	}

	if len(req.EventTypes) > 0 {
		if err := s.validateEventTypes(req.EventTypes); err != nil {
			return err
		}
	}

	if len(req.NotificationMethods) > 0 {
		if err := s.validateNotificationMethods(req.NotificationMethods); err != nil {
			return err
		}
	}

	if req.NotificationSchedule != nil {
		if err := s.validateNotificationSchedule(*req.NotificationSchedule); err != nil {
			return err
		}
	}

	if req.PriorityThreshold != nil {
		if err := s.validatePriorityThreshold(*req.PriorityThreshold); err != nil {
			return err
		}
	}

	if req.Notes != nil && len(*req.Notes) > 1000 {
		return domain.NewValidationError("notes cannot exceed 1000 characters")
	}

	return nil
}

// validateSubscriberName validates subscriber name
func (s *DefaultSubscriberService) validateSubscriberName(name string) error {
	if name == "" {
		return domain.NewValidationError("subscriber name is required")
	}

	if len(name) < 2 {
		return domain.NewValidationError("subscriber name must be at least 2 characters")
	}

	if len(name) > 100 {
		return domain.NewValidationError("subscriber name cannot exceed 100 characters")
	}

	return nil
}

// validateEmail validates email format
func (s *DefaultSubscriberService) validateEmail(email string) error {
	if email == "" {
		return domain.NewValidationError("email is required")
	}

	email = strings.TrimSpace(email)
	if len(email) < 3 || len(email) > 255 {
		return domain.NewValidationError("invalid email format")
	}

	// Basic email validation
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	if !emailRegex.MatchString(email) {
		return domain.NewValidationError("invalid email format")
	}

	return nil
}

// validatePhone validates phone number format
func (s *DefaultSubscriberService) validatePhone(phone string) error {
	phone = strings.TrimSpace(phone)
	if phone == "" {
		return nil // Phone is optional
	}

	// E.164 format validation
	e164Regex := regexp.MustCompile(`^\+[1-9]\d{1,14}$`)
	if !e164Regex.MatchString(phone) {
		return domain.NewValidationError("invalid phone number format")
	}

	return nil
}

// validateEventTypes validates event types array
func (s *DefaultSubscriberService) validateEventTypes(eventTypes []notifications.EventType) error {
	if len(eventTypes) == 0 {
		return domain.NewValidationError("at least one event type is required")
	}

	validEventTypes := map[notifications.EventType]bool{
		notifications.EventTypeInquiryMedia:            true,
		notifications.EventTypeInquiryBusiness:         true,
		notifications.EventTypeInquiryDonations:        true,
		notifications.EventTypeInquiryVolunteers:       true,
		notifications.EventTypeEventRegistration:       true,
		notifications.EventTypeSystemError:             true,
		notifications.EventTypeCapacityAlert:           true,
		notifications.EventTypeAdminActionRequired:     true,
		notifications.EventTypeComplianceAlert:         true,
	}

	for _, eventType := range eventTypes {
		if !validEventTypes[eventType] {
			return domain.NewValidationError(fmt.Sprintf("invalid event type: %s", eventType))
		}
	}

	return nil
}

// validateEventType validates a single event type
func (s *DefaultSubscriberService) validateEventType(eventType notifications.EventType) error {
	return s.validateEventTypes([]notifications.EventType{eventType})
}

// validateNotificationMethods validates notification methods array
func (s *DefaultSubscriberService) validateNotificationMethods(methods []notifications.NotificationMethod) error {
	if len(methods) == 0 {
		return domain.NewValidationError("at least one notification method is required")
	}

	validMethods := map[notifications.NotificationMethod]bool{
		notifications.NotificationMethodEmail: true,
		notifications.NotificationMethodSMS:   true,
		notifications.NotificationMethodBoth:  true,
	}

	for _, method := range methods {
		if !validMethods[method] {
			return domain.NewValidationError(fmt.Sprintf("invalid notification method: %s", method))
		}
	}

	return nil
}

// validateNotificationSchedule validates notification schedule
func (s *DefaultSubscriberService) validateNotificationSchedule(schedule notifications.NotificationSchedule) error {
	validSchedules := map[notifications.NotificationSchedule]bool{
		notifications.ScheduleImmediate: true,
		notifications.ScheduleHourly:    true,
		notifications.ScheduleDaily:     true,
	}

	if !validSchedules[schedule] {
		return domain.NewValidationError(fmt.Sprintf("invalid notification schedule: %s", schedule))
	}

	return nil
}

// validatePriorityThreshold validates priority threshold
func (s *DefaultSubscriberService) validatePriorityThreshold(priority notifications.PriorityThreshold) error {
	validPriorities := map[notifications.PriorityThreshold]bool{
		notifications.PriorityLow:    true,
		notifications.PriorityMedium: true,
		notifications.PriorityHigh:   true,
		notifications.PriorityUrgent: true,
	}

	if !validPriorities[priority] {
		return domain.NewValidationError(fmt.Sprintf("invalid priority threshold: %s", priority))
	}

	return nil
}

// validateSubscriberStatus validates subscriber status
func (s *DefaultSubscriberService) validateSubscriberStatus(status notifications.SubscriberStatus) error {
	validStatuses := map[notifications.SubscriberStatus]bool{
		notifications.SubscriberStatusActive:    true,
		notifications.SubscriberStatusInactive:  true,
		notifications.SubscriberStatusSuspended: true,
	}

	if !validStatuses[status] {
		return domain.NewValidationError(fmt.Sprintf("invalid subscriber status: %s", status))
	}

	return nil
}

// validateBusinessRules validates business-specific rules
func (s *DefaultSubscriberService) validateBusinessRules(ctx context.Context, subscriber *notifications.NotificationSubscriber) error {
	// Rule: SMS notification method requires phone number
	for _, method := range subscriber.NotificationMethods {
		if (method == notifications.NotificationMethodSMS || method == notifications.NotificationMethodBoth) && 
		   (subscriber.Phone == nil || *subscriber.Phone == "") {
			return domain.NewValidationError("phone number is required for SMS notifications")
		}
	}

	// Rule: Email notification method requires valid email (already validated above)

	// Rule: High/Urgent priority events should have immediate notifications
	if (subscriber.PriorityThreshold == notifications.PriorityHigh || 
		subscriber.PriorityThreshold == notifications.PriorityUrgent) &&
	   subscriber.NotificationSchedule != notifications.ScheduleImmediate {
		return domain.NewValidationError("high and urgent priority subscribers should use immediate notifications")
	}

	return nil
}

// applyUpdates applies update request to existing subscriber
func (s *DefaultSubscriberService) applyUpdates(existing *notifications.NotificationSubscriber, req *UpdateSubscriberRequest) *notifications.NotificationSubscriber {
	updated := *existing // Copy the existing subscriber

	// Apply updates
	if req.Status != nil {
		updated.Status = *req.Status
	}

	if req.SubscriberName != nil {
		updated.SubscriberName = *req.SubscriberName
	}

	if req.Email != nil {
		updated.Email = strings.ToLower(strings.TrimSpace(*req.Email))
	}

	if req.Phone != nil {
		updated.Phone = req.Phone
	}

	if len(req.EventTypes) > 0 {
		updated.EventTypes = req.EventTypes
	}

	if len(req.NotificationMethods) > 0 {
		updated.NotificationMethods = req.NotificationMethods
	}

	if req.NotificationSchedule != nil {
		updated.NotificationSchedule = *req.NotificationSchedule
	}

	if req.PriorityThreshold != nil {
		updated.PriorityThreshold = *req.PriorityThreshold
	}

	if req.Notes != nil {
		updated.Notes = req.Notes
	}

	// Always update timestamp and updater
	updated.UpdatedAt = time.Now().UTC()
	updated.UpdatedBy = req.UpdatedBy

	return &updated
}

// meetsPriorityThreshold checks if a subscriber's priority threshold meets the given priority
func (s *DefaultSubscriberService) meetsPriorityThreshold(subscriberThreshold, eventPriority notifications.PriorityThreshold) bool {
	priorityLevels := map[notifications.PriorityThreshold]int{
		notifications.PriorityLow:    1,
		notifications.PriorityMedium: 2,
		notifications.PriorityHigh:   3,
		notifications.PriorityUrgent: 4,
	}

	subscriberLevel := priorityLevels[subscriberThreshold]
	eventLevel := priorityLevels[eventPriority]

	return eventLevel >= subscriberLevel
}