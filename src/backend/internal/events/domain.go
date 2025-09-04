package events

import (
	"context"
	"fmt"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/axiom-software-co/international-center/src/backend/internal/shared/domain"
	"github.com/google/uuid"
)

// Event entity represents an event in the system
type Event struct {
	// Core event fields
	EventID     string `json:"event_id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Content     *string `json:"content,omitempty"`
	Slug        string `json:"slug"`
	CategoryID  string `json:"category_id"`
	
	// Media and links
	ImageURL    *string `json:"image_url,omitempty"`
	VirtualLink *string `json:"virtual_link,omitempty"`
	
	// Organizer and location
	OrganizerName *string `json:"organizer_name,omitempty"`
	Location      string  `json:"location"`
	
	// Event scheduling
	EventDate time.Time `json:"event_date"`
	EventTime *string   `json:"event_time,omitempty"`
	EndDate   *time.Time `json:"end_date,omitempty"`
	EndTime   *string    `json:"end_time,omitempty"`
	
	// Registration management
	MaxCapacity        *int                `json:"max_capacity,omitempty"`
	RegistrationDeadline *time.Time        `json:"registration_deadline,omitempty"`
	RegistrationStatus RegistrationStatus `json:"registration_status"`
	PublishingStatus   PublishingStatus   `json:"publishing_status"`
	
	// Content metadata
	Tags          []string   `json:"tags"`
	EventType     EventType  `json:"event_type"`
	PriorityLevel PriorityLevel `json:"priority_level"`
	
	// Audit fields
	CreatedOn  time.Time `json:"created_on"`
	CreatedBy  *string   `json:"created_by,omitempty"`
	ModifiedOn *time.Time `json:"modified_on,omitempty"`
	ModifiedBy *string    `json:"modified_by,omitempty"`
	
	// Soft delete fields
	IsDeleted bool       `json:"is_deleted"`
	DeletedOn *time.Time `json:"deleted_on,omitempty"`
	DeletedBy *string    `json:"deleted_by,omitempty"`
}

// EventCategory entity represents an event category
type EventCategory struct {
	CategoryID          string  `json:"category_id"`
	Name                string  `json:"name"`
	Slug                string  `json:"slug"`
	Description         *string `json:"description,omitempty"`
	IsDefaultUnassigned bool    `json:"is_default_unassigned"`
	
	// Audit fields
	CreatedOn  time.Time  `json:"created_on"`
	CreatedBy  *string    `json:"created_by,omitempty"`
	ModifiedOn *time.Time `json:"modified_on,omitempty"`
	ModifiedBy *string    `json:"modified_by,omitempty"`
	
	// Soft delete fields
	IsDeleted bool       `json:"is_deleted"`
	DeletedOn *time.Time `json:"deleted_on,omitempty"`
	DeletedBy *string    `json:"deleted_by,omitempty"`
}

// FeaturedEvent entity represents the currently featured event
type FeaturedEvent struct {
	FeaturedEventID string    `json:"featured_event_id"`
	EventID         string    `json:"event_id"`
	
	// Audit fields
	CreatedOn  time.Time  `json:"created_on"`
	CreatedBy  *string    `json:"created_by,omitempty"`
	ModifiedOn *time.Time `json:"modified_on,omitempty"`
	ModifiedBy *string    `json:"modified_by,omitempty"`
}

// EventRegistration entity represents a registration for an event
type EventRegistration struct {
	RegistrationID     string               `json:"registration_id"`
	EventID            string               `json:"event_id"`
	ParticipantName    string               `json:"participant_name"`
	ParticipantEmail   string               `json:"participant_email"`
	ParticipantPhone   *string              `json:"participant_phone,omitempty"`
	RegistrationTimestamp time.Time         `json:"registration_timestamp"`
	RegistrationStatus RegistrationStatus   `json:"registration_status"`
	
	// Special requirements
	SpecialRequirements   *string `json:"special_requirements,omitempty"`
	DietaryRestrictions   *string `json:"dietary_restrictions,omitempty"`
	AccessibilityNeeds    *string `json:"accessibility_needs,omitempty"`
	
	// Audit fields
	CreatedOn  time.Time  `json:"created_on"`
	CreatedBy  *string    `json:"created_by,omitempty"`
	ModifiedOn *time.Time `json:"modified_on,omitempty"`
	ModifiedBy *string    `json:"modified_by,omitempty"`
	
	// Soft delete fields
	IsDeleted bool       `json:"is_deleted"`
	DeletedOn *time.Time `json:"deleted_on,omitempty"`
	DeletedBy *string    `json:"deleted_by,omitempty"`
}

// Enums

// EventType represents the type of event
type EventType string

const (
	EventTypeWorkshop    EventType = "workshop"
	EventTypeSeminar     EventType = "seminar"
	EventTypeWebinar     EventType = "webinar"
	EventTypeConference  EventType = "conference"
	EventTypeFundraiser  EventType = "fundraiser"
	EventTypeCommunity   EventType = "community"
	EventTypeMedical     EventType = "medical"
	EventTypeEducational EventType = "educational"
)

// PublishingStatus represents the publishing status of an event
type PublishingStatus string

const (
	PublishingStatusDraft     PublishingStatus = "draft"
	PublishingStatusPublished PublishingStatus = "published"
	PublishingStatusArchived  PublishingStatus = "archived"
)

// RegistrationStatus represents the registration status for events
type RegistrationStatus string

const (
	RegistrationStatusOpen               RegistrationStatus = "open"
	RegistrationStatusRegistrationRequired RegistrationStatus = "registration_required"
	RegistrationStatusFull               RegistrationStatus = "full"
	RegistrationStatusCancelled          RegistrationStatus = "cancelled"
	RegistrationStatusRegistered         RegistrationStatus = "registered"
	RegistrationStatusConfirmed          RegistrationStatus = "confirmed"
	RegistrationStatusNoShow             RegistrationStatus = "no_show"
)

// PriorityLevel represents the priority level of an event
type PriorityLevel string

const (
	PriorityLevelLow    PriorityLevel = "low"
	PriorityLevelNormal PriorityLevel = "normal"
	PriorityLevelHigh   PriorityLevel = "high"
	PriorityLevelUrgent PriorityLevel = "urgent"
)

// Request/Response DTOs for admin operations

// AdminCreateEventRequest represents the request to create a new event
type AdminCreateEventRequest struct {
	Title                string  `json:"title"`
	Description          string  `json:"description"`
	Content              *string `json:"content,omitempty"`
	CategoryID           string  `json:"category_id"`
	ImageURL             *string `json:"image_url,omitempty"`
	OrganizerName        *string `json:"organizer_name,omitempty"`
	EventDate            string  `json:"event_date"`  // YYYY-MM-DD format
	EventTime            *string `json:"event_time,omitempty"` // HH:MM format
	EndDate              *string `json:"end_date,omitempty"`   // YYYY-MM-DD format
	EndTime              *string `json:"end_time,omitempty"`   // HH:MM format
	Location             string  `json:"location"`
	VirtualLink          *string `json:"virtual_link,omitempty"`
	MaxCapacity          *int    `json:"max_capacity,omitempty"`
	RegistrationDeadline *string `json:"registration_deadline,omitempty"` // YYYY-MM-DDTHH:MM:SSZ format
	EventType            string  `json:"event_type"`
	PriorityLevel        *string `json:"priority_level,omitempty"`
	Tags                 []string `json:"tags,omitempty"`
}

// AdminUpdateEventRequest represents the request to update an event
type AdminUpdateEventRequest struct {
	Title                *string  `json:"title,omitempty"`
	Description          *string  `json:"description,omitempty"`
	Content              *string  `json:"content,omitempty"`
	CategoryID           *string  `json:"category_id,omitempty"`
	ImageURL             *string  `json:"image_url,omitempty"`
	OrganizerName        *string  `json:"organizer_name,omitempty"`
	EventDate            *string  `json:"event_date,omitempty"`  // YYYY-MM-DD format
	EventTime            *string  `json:"event_time,omitempty"`  // HH:MM format
	EndDate              *string  `json:"end_date,omitempty"`    // YYYY-MM-DD format
	EndTime              *string  `json:"end_time,omitempty"`    // HH:MM format
	Location             *string  `json:"location,omitempty"`
	VirtualLink          *string  `json:"virtual_link,omitempty"`
	MaxCapacity          *int     `json:"max_capacity,omitempty"`
	RegistrationDeadline *string  `json:"registration_deadline,omitempty"` // YYYY-MM-DDTHH:MM:SSZ format
	EventType            *string  `json:"event_type,omitempty"`
	PriorityLevel        *string  `json:"priority_level,omitempty"`
	Tags                 []string `json:"tags,omitempty"`
}

// AdminCreateEventCategoryRequest represents the request to create a new event category
type AdminCreateEventCategoryRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// Constructor functions

// NewEvent creates a new event entity with required validation
func NewEvent(title, description, categoryID, location string, eventDate time.Time, eventType EventType, userID string) (*Event, error) {
	if err := validateEventTitle(title); err != nil {
		return nil, err
	}
	
	if err := validateEventDescription(description); err != nil {
		return nil, err
	}
	
	if err := validateCategoryID(categoryID); err != nil {
		return nil, err
	}
	
	if err := validateEventLocation(location); err != nil {
		return nil, err
	}
	
	if err := validateEventType(eventType); err != nil {
		return nil, err
	}
	
	if err := validateEventDate(eventDate); err != nil {
		return nil, err
	}

	eventID := uuid.New().String()
	slug := generateEventSlug(title)
	
	return &Event{
		EventID:            eventID,
		Title:              title,
		Description:        description,
		Slug:               slug,
		CategoryID:         categoryID,
		Location:           location,
		EventDate:          eventDate,
		EventType:          eventType,
		PublishingStatus:   PublishingStatusDraft,
		RegistrationStatus: RegistrationStatusOpen,
		PriorityLevel:      PriorityLevelNormal,
		Tags:               []string{},
		CreatedOn:          time.Now(),
		CreatedBy:          &userID,
		IsDeleted:          false,
	}, nil
}

// NewEventCategory creates a new event category entity with required validation
func NewEventCategory(name, description string, userID string) (*EventCategory, error) {
	if err := validateCategoryName(name); err != nil {
		return nil, err
	}
	
	categoryID := uuid.New().String()
	slug := generateCategorySlug(name)
	
	return &EventCategory{
		CategoryID:          categoryID,
		Name:                name,
		Slug:                slug,
		Description:         &description,
		IsDefaultUnassigned: false,
		CreatedOn:           time.Now(),
		CreatedBy:           &userID,
		IsDeleted:           false,
	}, nil
}

// NewFeaturedEvent creates a new featured event entity
func NewFeaturedEvent(eventID string, userID string) (*FeaturedEvent, error) {
	if err := validateEventID(eventID); err != nil {
		return nil, err
	}
	
	featuredEventID := uuid.New().String()
	
	return &FeaturedEvent{
		FeaturedEventID: featuredEventID,
		EventID:         eventID,
		CreatedOn:       time.Now(),
		CreatedBy:       &userID,
	}, nil
}

// Business logic methods

// UpdateEvent updates the event with new information
func (e *Event) UpdateEvent(title, description, location *string, eventDate *time.Time, userID string) error {
	if title != nil {
		if err := validateEventTitle(*title); err != nil {
			return err
		}
		e.Title = *title
		e.Slug = generateEventSlug(*title)
	}
	
	if description != nil {
		if err := validateEventDescription(*description); err != nil {
			return err
		}
		e.Description = *description
	}
	
	if location != nil {
		if err := validateEventLocation(*location); err != nil {
			return err
		}
		e.Location = *location
	}
	
	if eventDate != nil {
		if err := validateEventDate(*eventDate); err != nil {
			return err
		}
		e.EventDate = *eventDate
	}
	
	e.ModifiedOn = &[]time.Time{time.Now()}[0]
	e.ModifiedBy = &userID
	
	return nil
}

// PublishEvent changes the event status to published
func (e *Event) PublishEvent(userID string) error {
	if e.PublishingStatus == PublishingStatusPublished {
		return domain.NewValidationError("event is already published")
	}
	
	if e.PublishingStatus == PublishingStatusArchived {
		return domain.NewValidationError("cannot publish archived event")
	}
	
	e.PublishingStatus = PublishingStatusPublished
	e.ModifiedOn = &[]time.Time{time.Now()}[0]
	e.ModifiedBy = &userID
	
	return nil
}

// ArchiveEvent changes the event status to archived
func (e *Event) ArchiveEvent(userID string) error {
	if e.PublishingStatus == PublishingStatusDraft {
		return domain.NewValidationError("cannot archive draft event, publish first")
	}
	
	if e.PublishingStatus == PublishingStatusArchived {
		return domain.NewValidationError("event is already archived")
	}
	
	e.PublishingStatus = PublishingStatusArchived
	e.ModifiedOn = &[]time.Time{time.Now()}[0]
	e.ModifiedBy = &userID
	
	return nil
}

// SetImageURL sets the image URL for the event
func (e *Event) SetImageURL(imageURL string, userID string) error {
	if imageURL != "" {
		if err := validateImageURL(imageURL); err != nil {
			return err
		}
	}
	
	if imageURL == "" {
		e.ImageURL = nil
	} else {
		e.ImageURL = &imageURL
	}
	
	e.ModifiedOn = &[]time.Time{time.Now()}[0]
	e.ModifiedBy = &userID
	
	return nil
}

// SetVirtualLink sets the virtual link for the event
func (e *Event) SetVirtualLink(virtualLink string, userID string) error {
	if virtualLink != "" {
		if err := validateVirtualLink(virtualLink); err != nil {
			return err
		}
	}
	
	if virtualLink == "" {
		e.VirtualLink = nil
	} else {
		e.VirtualLink = &virtualLink
	}
	
	e.ModifiedOn = &[]time.Time{time.Now()}[0]
	e.ModifiedBy = &userID
	
	return nil
}

// SetMaxCapacity sets the maximum capacity for the event
func (e *Event) SetMaxCapacity(maxCapacity *int, userID string) error {
	if maxCapacity != nil && *maxCapacity < 0 {
		return domain.NewValidationError("max capacity cannot be negative")
	}
	
	e.MaxCapacity = maxCapacity
	e.ModifiedOn = &[]time.Time{time.Now()}[0]
	e.ModifiedBy = &userID
	
	return nil
}

// SetRegistrationDeadline sets the registration deadline for the event
func (e *Event) SetRegistrationDeadline(deadline *time.Time, userID string) error {
	if deadline != nil {
		if deadline.After(e.EventDate) {
			return domain.NewValidationError("registration deadline cannot be after event date")
		}
	}
	
	e.RegistrationDeadline = deadline
	e.ModifiedOn = &[]time.Time{time.Now()}[0]
	e.ModifiedBy = &userID
	
	return nil
}

// SetEventType sets the event type
func (e *Event) SetEventType(eventType EventType, userID string) error {
	if err := validateEventType(eventType); err != nil {
		return err
	}
	
	e.EventType = eventType
	e.ModifiedOn = &[]time.Time{time.Now()}[0]
	e.ModifiedBy = &userID
	
	return nil
}

// SetPriorityLevel sets the priority level for the event
func (e *Event) SetPriorityLevel(priorityLevel PriorityLevel, userID string) error {
	if err := validatePriorityLevel(priorityLevel); err != nil {
		return err
	}
	
	e.PriorityLevel = priorityLevel
	e.ModifiedOn = &[]time.Time{time.Now()}[0]
	e.ModifiedBy = &userID
	
	return nil
}

// AddTags adds tags to the event
func (e *Event) AddTags(tags []string, userID string) error {
	if len(tags) > 10 {
		return domain.NewValidationError("cannot have more than 10 tags")
	}
	
	for _, tag := range tags {
		if err := validateTag(tag); err != nil {
			return err
		}
	}
	
	e.Tags = tags
	e.ModifiedOn = &[]time.Time{time.Now()}[0]
	e.ModifiedBy = &userID
	
	return nil
}

// CanBeFeatured checks if the event can be featured
func (e *Event) CanBeFeatured() error {
	if e.PublishingStatus != PublishingStatusPublished {
		return domain.NewValidationError("only published events can be featured")
	}
	
	if e.IsDeleted {
		return domain.NewValidationError("deleted events cannot be featured")
	}
	
	return nil
}

// IsRegistrationOpen checks if registration is currently open
func (e *Event) IsRegistrationOpen() bool {
	now := time.Now()
	
	// Check if registration is explicitly closed
	if e.RegistrationStatus != RegistrationStatusOpen {
		return false
	}
	
	// Check if registration deadline has passed
	if e.RegistrationDeadline != nil && now.After(*e.RegistrationDeadline) {
		return false
	}
	
	// Check if event date has passed
	if now.After(e.EventDate) {
		return false
	}
	
	return true
}

// Validation functions

func validateEventTitle(title string) error {
	if strings.TrimSpace(title) == "" {
		return domain.NewValidationError("event title cannot be empty")
	}
	
	if len(title) > 255 {
		return domain.NewValidationError("event title cannot exceed 255 characters")
	}
	
	return nil
}

func validateEventDescription(description string) error {
	if strings.TrimSpace(description) == "" {
		return domain.NewValidationError("event description cannot be empty")
	}
	
	if len(description) > 2000 {
		return domain.NewValidationError("event description cannot exceed 2000 characters")
	}
	
	return nil
}

func validateCategoryID(categoryID string) error {
	if strings.TrimSpace(categoryID) == "" {
		return domain.NewValidationError("category ID cannot be empty")
	}
	
	if _, err := uuid.Parse(categoryID); err != nil {
		return domain.NewValidationError("category ID must be a valid UUID")
	}
	
	return nil
}

func validateEventID(eventID string) error {
	if strings.TrimSpace(eventID) == "" {
		return domain.NewValidationError("event ID cannot be empty")
	}
	
	if _, err := uuid.Parse(eventID); err != nil {
		return domain.NewValidationError("event ID must be a valid UUID")
	}
	
	return nil
}

func validateEventLocation(location string) error {
	if strings.TrimSpace(location) == "" {
		return domain.NewValidationError("event location cannot be empty")
	}
	
	if len(location) > 500 {
		return domain.NewValidationError("event location cannot exceed 500 characters")
	}
	
	return nil
}

func validateEventType(eventType EventType) error {
	validTypes := []EventType{
		EventTypeWorkshop,
		EventTypeSeminar,
		EventTypeWebinar,
		EventTypeConference,
		EventTypeFundraiser,
		EventTypeCommunity,
		EventTypeMedical,
		EventTypeEducational,
	}
	
	for _, valid := range validTypes {
		if eventType == valid {
			return nil
		}
	}
	
	return domain.NewValidationError(fmt.Sprintf("invalid event type: %s", eventType))
}

func validateEventDate(eventDate time.Time) error {
	// Events can be in the past for historical records
	// but new events should typically be in the future
	return nil
}

func validatePriorityLevel(priorityLevel PriorityLevel) error {
	validLevels := []PriorityLevel{
		PriorityLevelLow,
		PriorityLevelNormal,
		PriorityLevelHigh,
		PriorityLevelUrgent,
	}
	
	for _, valid := range validLevels {
		if priorityLevel == valid {
			return nil
		}
	}
	
	return domain.NewValidationError(fmt.Sprintf("invalid priority level: %s", priorityLevel))
}

func validateCategoryName(name string) error {
	if strings.TrimSpace(name) == "" {
		return domain.NewValidationError("category name cannot be empty")
	}
	
	if len(name) > 255 {
		return domain.NewValidationError("category name cannot exceed 255 characters")
	}
	
	return nil
}

func validateImageURL(imageURL string) error {
	if strings.TrimSpace(imageURL) == "" {
		return domain.NewValidationError("image URL cannot be empty")
	}
	
	parsedURL, err := url.Parse(imageURL)
	if err != nil {
		return domain.NewValidationError("image URL must be a valid URL")
	}
	
	if parsedURL.Scheme != "https" {
		return domain.NewValidationError("image URL must use HTTPS")
	}
	
	if len(imageURL) > 500 {
		return domain.NewValidationError("image URL cannot exceed 500 characters")
	}
	
	return nil
}

func validateVirtualLink(virtualLink string) error {
	if strings.TrimSpace(virtualLink) == "" {
		return domain.NewValidationError("virtual link cannot be empty")
	}
	
	parsedURL, err := url.Parse(virtualLink)
	if err != nil {
		return domain.NewValidationError("virtual link must be a valid URL")
	}
	
	if parsedURL.Scheme != "https" {
		return domain.NewValidationError("virtual link must use HTTPS")
	}
	
	if len(virtualLink) > 500 {
		return domain.NewValidationError("virtual link cannot exceed 500 characters")
	}
	
	return nil
}

func validateTag(tag string) error {
	tag = strings.TrimSpace(tag)
	if tag == "" {
		return domain.NewValidationError("tag cannot be empty")
	}
	
	if len(tag) > 50 {
		return domain.NewValidationError("tag cannot exceed 50 characters")
	}
	
	// Tags should only contain alphanumeric characters, hyphens, and spaces
	matched, _ := regexp.MatchString(`^[a-zA-Z0-9\s\-]+$`, tag)
	if !matched {
		return domain.NewValidationError("tag contains invalid characters")
	}
	
	return nil
}

// Utility functions

func generateEventSlug(title string) string {
	// Convert to lowercase and replace spaces with hyphens
	slug := strings.ToLower(title)
	slug = regexp.MustCompile(`[^a-z0-9\s]`).ReplaceAllString(slug, "")
	slug = regexp.MustCompile(`\s+`).ReplaceAllString(slug, "-")
	slug = strings.Trim(slug, "-")
	
	// Truncate if too long
	if len(slug) > 100 {
		slug = slug[:100]
	}
	
	return slug
}

func generateCategorySlug(name string) string {
	// Convert to lowercase and replace spaces with hyphens
	slug := strings.ToLower(name)
	slug = regexp.MustCompile(`[^a-z0-9\s]`).ReplaceAllString(slug, "")
	slug = regexp.MustCompile(`\s+`).ReplaceAllString(slug, "-")
	slug = strings.Trim(slug, "-")
	
	// Truncate if too long
	if len(slug) > 50 {
		slug = slug[:50]
	}
	
	return slug
}

// IsValidEventType checks if the given string is a valid event type
func IsValidEventType(eventType string) bool {
	validTypes := []string{
		string(EventTypeWorkshop),
		string(EventTypeSeminar),
		string(EventTypeWebinar),
		string(EventTypeConference),
		string(EventTypeFundraiser),
		string(EventTypeCommunity),
		string(EventTypeMedical),
		string(EventTypeEducational),
	}
	
	for _, valid := range validTypes {
		if eventType == valid {
			return true
		}
	}
	
	return false
}

// IsValidPriorityLevel checks if the given string is a valid priority level
func IsValidPriorityLevel(priorityLevel string) bool {
	validLevels := []string{
		string(PriorityLevelLow),
		string(PriorityLevelNormal),
		string(PriorityLevelHigh),
		string(PriorityLevelUrgent),
	}
	
	for _, valid := range validLevels {
		if priorityLevel == valid {
			return true
		}
	}
	
	return false
}

// IsAdminUser checks if the user has admin privileges
func IsAdminUser(userID string) bool {
	return strings.HasPrefix(userID, "admin-")
}

// Repository interface for events domain
type EventsRepositoryInterface interface {
	// Event operations
	SaveEvent(ctx context.Context, event *Event) error
	GetEvent(ctx context.Context, eventID string) (*Event, error)
	DeleteEvent(ctx context.Context, eventID string, userID string) error
	
	// Event category operations
	SaveEventCategory(ctx context.Context, category *EventCategory) error
	GetEventCategory(ctx context.Context, categoryID string) (*EventCategory, error)
	DeleteEventCategory(ctx context.Context, categoryID string, userID string) error
	GetDefaultUnassignedCategory(ctx context.Context) (*EventCategory, error)
	
	// Featured event operations
	SaveFeaturedEvent(ctx context.Context, featuredEvent *FeaturedEvent) error
	GetFeaturedEvent(ctx context.Context) (*FeaturedEvent, error)
	DeleteFeaturedEvent(ctx context.Context) error
	
	// Event registration operations
	GetEventRegistrations(ctx context.Context, eventID string) ([]*EventRegistration, error)
	
	// Audit operations
	PublishAuditEvent(ctx context.Context, entityType domain.EntityType, entityID string, operationType domain.AuditEventType, userID string, beforeData, afterData interface{}) error
}