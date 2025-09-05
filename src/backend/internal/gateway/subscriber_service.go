package gateway

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/axiom-software-co/international-center/src/backend/internal/shared/domain"
	"github.com/google/uuid"
)

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
func (s *DefaultSubscriberService) CreateSubscriber(ctx context.Context, req *CreateSubscriberRequest) (*NotificationSubscriber, error) {
	if req == nil {
		return nil, domain.NewValidationError("create request cannot be nil", nil)
	}

	// Validate request
	if err := s.validateCreateRequest(req); err != nil {
		return nil, err
	}

	// Generate new subscriber ID
	subscriberID := uuid.New().String()

	// Create subscriber domain model
	now := time.Now().UTC()
	subscriber := &NotificationSubscriber{
		SubscriberID:         subscriberID,
		Status:               SubscriberStatusActive, // Default status
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
func (s *DefaultSubscriberService) GetSubscriber(ctx context.Context, subscriberID string) (*NotificationSubscriber, error) {
	if subscriberID == "" {
		return nil, domain.NewValidationError("subscriber ID cannot be empty", nil)
	}

	// Validate UUID format
	if _, err := uuid.Parse(subscriberID); err != nil {
		return nil, domain.NewValidationError("invalid subscriber ID format", err)
	}

	subscriber, err := s.repository.GetSubscriber(ctx, subscriberID)
	if err != nil {
		return nil, fmt.Errorf("failed to get subscriber: %w", err)
	}

	return subscriber, nil
}

// UpdateSubscriber updates an existing subscriber
func (s *DefaultSubscriberService) UpdateSubscriber(ctx context.Context, subscriberID string, req *UpdateSubscriberRequest) (*NotificationSubscriber, error) {
	if subscriberID == "" {
		return nil, domain.NewValidationError("subscriber ID cannot be empty", nil)
	}

	if req == nil {
		return nil, domain.NewValidationError("update request cannot be nil", nil)
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
		return domain.NewValidationError("subscriber ID cannot be empty", nil)
	}

	if deletedBy == "" {
		return domain.NewValidationError("deleted by cannot be empty", nil)
	}

	// Validate UUID format
	if _, err := uuid.Parse(subscriberID); err != nil {
		return domain.NewValidationError("invalid subscriber ID format", err)
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
func (s *DefaultSubscriberService) ListSubscribers(ctx context.Context, status *SubscriberStatus, page, pageSize int) ([]*NotificationSubscriber, int, error) {
	// Validate pagination parameters
	if page < 1 {
		return nil, 0, domain.NewValidationError("page must be greater than 0", nil)
	}

	if pageSize < 1 || pageSize > 100 {
		return nil, 0, domain.NewValidationError("page size must be between 1 and 100", nil)
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
func (s *DefaultSubscriberService) GetSubscribersByEvent(ctx context.Context, eventType EventType, priority PriorityThreshold) ([]*NotificationSubscriber, error) {
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
	var filteredSubscribers []*NotificationSubscriber
	for _, subscriber := range eventSubscribers {
		if s.meetsPriorityThreshold(subscriber.PriorityThreshold, priority) {
			filteredSubscribers = append(filteredSubscribers, subscriber)
		}
	}

	return filteredSubscribers, nil
}

// ValidateSubscriber validates a subscriber model
func (s *DefaultSubscriberService) ValidateSubscriber(ctx context.Context, subscriber *NotificationSubscriber) error {
	if subscriber == nil {
		return domain.NewValidationError("subscriber cannot be nil", nil)
	}

	// Validate subscriber ID
	if subscriber.SubscriberID == "" {
		return domain.NewValidationError("subscriber ID is required", nil)
	}

	if _, err := uuid.Parse(subscriber.SubscriberID); err != nil {
		return domain.NewValidationError("invalid subscriber ID format", err)
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
		return domain.NewValidationError("notes cannot exceed 1000 characters", nil)
	}

	// Validate required fields
	if subscriber.CreatedBy == "" {
		return domain.NewValidationError("created by is required", nil)
	}

	if subscriber.UpdatedBy == "" {
		return domain.NewValidationError("updated by is required", nil)
	}

	// Validate business rules
	return s.validateBusinessRules(ctx, subscriber)
}

// Private validation methods

// validateCreateRequest validates a create subscriber request
func (s *DefaultSubscriberService) validateCreateRequest(req *CreateSubscriberRequest) error {
	if req.SubscriberName == "" {
		return domain.NewValidationError("subscriber name is required", nil)
	}

	if len(req.SubscriberName) < 2 {
		return domain.NewValidationError("subscriber name must be at least 2 characters", nil)
	}

	if len(req.SubscriberName) > 100 {
		return domain.NewValidationError("subscriber name cannot exceed 100 characters", nil)
	}

	if req.Email == "" {
		return domain.NewValidationError("email is required", nil)
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
		return domain.NewValidationError("at least one event type is required", nil)
	}

	if err := s.validateEventTypes(req.EventTypes); err != nil {
		return err
	}

	if len(req.NotificationMethods) == 0 {
		return domain.NewValidationError("at least one notification method is required", nil)
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
		return domain.NewValidationError("created by is required", nil)
	}

	if req.Notes != nil && len(*req.Notes) > 1000 {
		return domain.NewValidationError("notes cannot exceed 1000 characters", nil)
	}

	return nil
}

// validateUpdateRequest validates an update subscriber request
func (s *DefaultSubscriberService) validateUpdateRequest(req *UpdateSubscriberRequest) error {
	if req.UpdatedBy == "" {
		return domain.NewValidationError("updated by is required", nil)
	}

	if req.Status != nil {
		if err := s.validateSubscriberStatus(*req.Status); err != nil {
			return err
		}
	}

	if req.SubscriberName != nil {
		if len(*req.SubscriberName) < 2 {
			return domain.NewValidationError("subscriber name must be at least 2 characters", nil)
		}
		if len(*req.SubscriberName) > 100 {
			return domain.NewValidationError("subscriber name cannot exceed 100 characters", nil)
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
		return domain.NewValidationError("notes cannot exceed 1000 characters", nil)
	}

	return nil
}

// validateSubscriberName validates subscriber name
func (s *DefaultSubscriberService) validateSubscriberName(name string) error {
	if name == "" {
		return domain.NewValidationError("subscriber name is required", nil)
	}

	if len(name) < 2 {
		return domain.NewValidationError("subscriber name must be at least 2 characters", nil)
	}

	if len(name) > 100 {
		return domain.NewValidationError("subscriber name cannot exceed 100 characters", nil)
	}

	return nil
}

// validateEmail validates email format
func (s *DefaultSubscriberService) validateEmail(email string) error {
	if email == "" {
		return domain.NewValidationError("email is required", nil)
	}

	email = strings.TrimSpace(email)
	if len(email) < 3 || len(email) > 255 {
		return domain.NewValidationError("invalid email format", nil)
	}

	// Basic email validation
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	if !emailRegex.MatchString(email) {
		return domain.NewValidationError("invalid email format", nil)
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
		return domain.NewValidationError("invalid phone number format", nil)
	}

	return nil
}

// validateEventTypes validates event types array
func (s *DefaultSubscriberService) validateEventTypes(eventTypes []EventType) error {
	if len(eventTypes) == 0 {
		return domain.NewValidationError("at least one event type is required", nil)
	}

	validEventTypes := map[EventType]bool{
		EventTypeInquiryMedia:            true,
		EventTypeInquiryBusiness:         true,
		EventTypeInquiryDonations:        true,
		EventTypeInquiryVolunteers:       true,
		EventTypeEventRegistration:       true,
		EventTypeSystemError:             true,
		EventTypeCapacityAlert:           true,
		EventTypeAdminActionRequired:     true,
		EventTypeComplianceAlert:         true,
	}

	for _, eventType := range eventTypes {
		if !validEventTypes[eventType] {
			return domain.NewValidationError(fmt.Sprintf("invalid event type: %s", eventType), nil)
		}
	}

	return nil
}

// validateEventType validates a single event type
func (s *DefaultSubscriberService) validateEventType(eventType EventType) error {
	return s.validateEventTypes([]EventType{eventType})
}

// validateNotificationMethods validates notification methods array
func (s *DefaultSubscriberService) validateNotificationMethods(methods []NotificationMethod) error {
	if len(methods) == 0 {
		return domain.NewValidationError("at least one notification method is required", nil)
	}

	validMethods := map[NotificationMethod]bool{
		NotificationMethodEmail: true,
		NotificationMethodSMS:   true,
		NotificationMethodBoth:  true,
	}

	for _, method := range methods {
		if !validMethods[method] {
			return domain.NewValidationError(fmt.Sprintf("invalid notification method: %s", method), nil)
		}
	}

	return nil
}

// validateNotificationSchedule validates notification schedule
func (s *DefaultSubscriberService) validateNotificationSchedule(schedule NotificationSchedule) error {
	validSchedules := map[NotificationSchedule]bool{
		NotificationScheduleImmediate: true,
		NotificationScheduleHourly:    true,
		NotificationScheduleDaily:     true,
	}

	if !validSchedules[schedule] {
		return domain.NewValidationError(fmt.Sprintf("invalid notification schedule: %s", schedule), nil)
	}

	return nil
}

// validatePriorityThreshold validates priority threshold
func (s *DefaultSubscriberService) validatePriorityThreshold(priority PriorityThreshold) error {
	validPriorities := map[PriorityThreshold]bool{
		PriorityThresholdLow:    true,
		PriorityThresholdMedium: true,
		PriorityThresholdHigh:   true,
		PriorityThresholdUrgent: true,
	}

	if !validPriorities[priority] {
		return domain.NewValidationError(fmt.Sprintf("invalid priority threshold: %s", priority), nil)
	}

	return nil
}

// validateSubscriberStatus validates subscriber status
func (s *DefaultSubscriberService) validateSubscriberStatus(status SubscriberStatus) error {
	validStatuses := map[SubscriberStatus]bool{
		SubscriberStatusActive:    true,
		SubscriberStatusInactive:  true,
		SubscriberStatusSuspended: true,
	}

	if !validStatuses[status] {
		return domain.NewValidationError(fmt.Sprintf("invalid subscriber status: %s", status), nil)
	}

	return nil
}

// validateBusinessRules validates business-specific rules
func (s *DefaultSubscriberService) validateBusinessRules(ctx context.Context, subscriber *NotificationSubscriber) error {
	// Rule: SMS notification method requires phone number
	for _, method := range subscriber.NotificationMethods {
		if (method == NotificationMethodSMS || method == NotificationMethodBoth) && 
		   (subscriber.Phone == nil || *subscriber.Phone == "") {
			return domain.NewValidationError("phone number is required for SMS notifications", nil)
		}
	}

	// Rule: Email notification method requires valid email (already validated above)

	// Rule: High/Urgent priority events should have immediate notifications
	if (subscriber.PriorityThreshold == PriorityThresholdHigh || 
		subscriber.PriorityThreshold == PriorityThresholdUrgent) &&
	   subscriber.NotificationSchedule != NotificationScheduleImmediate {
		return domain.NewValidationError("high and urgent priority subscribers should use immediate notifications", nil)
	}

	return nil
}

// applyUpdates applies update request to existing subscriber
func (s *DefaultSubscriberService) applyUpdates(existing *NotificationSubscriber, req *UpdateSubscriberRequest) *NotificationSubscriber {
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
func (s *DefaultSubscriberService) meetsPriorityThreshold(subscriberThreshold, eventPriority PriorityThreshold) bool {
	priorityLevels := map[PriorityThreshold]int{
		PriorityThresholdLow:    1,
		PriorityThresholdMedium: 2,
		PriorityThresholdHigh:   3,
		PriorityThresholdUrgent: 4,
	}

	subscriberLevel := priorityLevels[subscriberThreshold]
	eventLevel := priorityLevels[eventPriority]

	return eventLevel >= subscriberLevel
}