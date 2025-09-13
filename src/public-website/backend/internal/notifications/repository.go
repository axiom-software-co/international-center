package notifications

import (
	"context"

	"github.com/axiom-software-co/international-center/src/backend/internal/shared/domain"
)

// SubscriberRepository provides operations for notification subscribers through Dapr state store
type SubscriberRepository interface {
	CreateSubscriber(ctx context.Context, subscriber *NotificationSubscriber) error
	GetSubscriber(ctx context.Context, subscriberID string) (*NotificationSubscriber, error)
	GetSubscriberByEmail(ctx context.Context, email string) (*NotificationSubscriber, error)
	UpdateSubscriber(ctx context.Context, subscriber *NotificationSubscriber) error
	DeleteSubscriber(ctx context.Context, subscriberID string, deletedBy string) error
	ListSubscribers(ctx context.Context, status *SubscriberStatus, limit, offset int) ([]*NotificationSubscriber, int, error)
	GetSubscribersByEventType(ctx context.Context, eventType EventType) ([]*NotificationSubscriber, error)
	GetActiveSubscribersByPriority(ctx context.Context, priority PriorityThreshold) ([]*NotificationSubscriber, error)
	CheckEmailExists(ctx context.Context, email string, excludeID *string) (bool, error)
	HealthCheck(ctx context.Context) error
	PublishAuditEvent(ctx context.Context, entityType domain.EntityType, entityID string, operationType domain.AuditEventType, userID string, beforeData, afterData interface{}) error
}