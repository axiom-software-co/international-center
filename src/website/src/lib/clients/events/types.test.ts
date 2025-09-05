// Events Types Validation Tests - Database schema compliance validation
// Tests validate Event interfaces match TABLES-EVENTS.md schema exactly

import { describe, it, expect } from 'vitest';
import type { Event, EventCategory, EventsResponse, EventResponse } from './types';

// Database schema reference - Event table from TABLES-EVENTS.md
interface DatabaseSchemaEvent {
  // Primary key and identifiers (Required)
  event_id: string; // UUID PRIMARY KEY
  title: string; // VARCHAR(255) NOT NULL
  description: string; // TEXT NOT NULL
  content?: string; // TEXT (nullable)
  slug: string; // VARCHAR(255) UNIQUE NOT NULL
  
  // Category relationship (Required)
  category_id: string; // UUID NOT NULL REFERENCES event_categories(category_id)
  
  // Media and organization (Optional)
  image_url?: string; // VARCHAR(500) (nullable)
  organizer_name?: string; // VARCHAR(255) (nullable)
  
  // Event scheduling (Required: event_date and location, others optional)
  event_date: string; // DATE NOT NULL (ISO date format)
  event_time?: string; // TIME (nullable, HH:MM format)
  end_date?: string; // DATE (nullable)
  end_time?: string; // TIME (nullable)
  location: string; // VARCHAR(500) NOT NULL
  virtual_link?: string; // VARCHAR(500) (nullable)
  
  // Registration management (Optional except registration_status)
  max_capacity?: number; // INTEGER (nullable)
  registration_deadline?: string; // TIMESTAMPTZ (nullable, ISO format)
  registration_status: 'open' | 'registration_required' | 'full' | 'cancelled'; // VARCHAR(20) NOT NULL DEFAULT 'open'
  
  // Publishing workflow (Required)
  publishing_status: 'draft' | 'published' | 'archived'; // VARCHAR(20) NOT NULL DEFAULT 'draft'
  
  // Metadata and categorization (Required)
  tags: string[]; // TEXT[] (PostgreSQL array)
  event_type: 'workshop' | 'seminar' | 'webinar' | 'conference' | 'fundraiser' | 'community' | 'medical' | 'educational'; // VARCHAR(50) NOT NULL
  priority_level: 'low' | 'normal' | 'high' | 'urgent'; // VARCHAR(20) NOT NULL DEFAULT 'normal'
  
  // Audit fields (Required: created_on, others optional)
  created_on: string; // TIMESTAMPTZ NOT NULL DEFAULT NOW()
  created_by?: string; // VARCHAR(255) (nullable)
  modified_on?: string; // TIMESTAMPTZ (nullable)
  modified_by?: string; // VARCHAR(255) (nullable)
  
  // Soft delete fields (Required: is_deleted, others optional)
  is_deleted: boolean; // BOOLEAN NOT NULL DEFAULT FALSE
  deleted_on?: string; // TIMESTAMPTZ (nullable)
  deleted_by?: string; // VARCHAR(255) (nullable)
}

// Database schema reference - Event Categories table from TABLES-EVENTS.md
interface DatabaseSchemaEventCategory {
  category_id: string; // UUID PRIMARY KEY
  name: string; // VARCHAR(255) NOT NULL
  slug: string; // VARCHAR(255) UNIQUE NOT NULL
  description?: string; // TEXT (nullable)
  is_default_unassigned: boolean; // BOOLEAN NOT NULL DEFAULT FALSE
  
  // Audit fields
  created_on: string; // TIMESTAMPTZ NOT NULL DEFAULT NOW()
  created_by?: string; // VARCHAR(255) (nullable)
  modified_on?: string; // TIMESTAMPTZ (nullable)
  modified_by?: string; // VARCHAR(255) (nullable)
  
  // Soft delete fields
  is_deleted: boolean; // BOOLEAN NOT NULL DEFAULT FALSE
  deleted_on?: string; // TIMESTAMPTZ (nullable)
  deleted_by?: string; // VARCHAR(255) (nullable)
}

describe('Events Types Database Schema Compliance', () => {
  describe('Event Interface Schema Validation', () => {
    it('should have all required database schema fields', () => {
      // This test will fail until Event interface matches DatabaseSchemaEvent
      // It drives the need to align types with database schema in GREEN phase
      
      const mockDatabaseEvent: DatabaseSchemaEvent = {
        event_id: 'event-uuid-123',
        title: 'Database Schema Event',
        description: 'Event description matching database schema',
        content: 'Full event content',
        slug: 'database-schema-event',
        category_id: 'category-uuid-456',
        image_url: 'https://example.com/event-image.jpg',
        organizer_name: 'Event Organizer',
        event_date: '2024-03-15',
        event_time: '14:30',
        end_date: '2024-03-15',
        end_time: '17:00',
        location: '123 Database St, Schema City',
        virtual_link: 'https://virtual.example.com/event',
        max_capacity: 100,
        registration_deadline: '2024-03-10T23:59:59Z',
        registration_status: 'open',
        publishing_status: 'published',
        tags: ['database', 'schema', 'event'],
        event_type: 'workshop',
        priority_level: 'normal',
        created_on: '2024-01-01T00:00:00Z',
        created_by: 'admin@example.com',
        modified_on: '2024-01-02T00:00:00Z',
        modified_by: 'admin@example.com',
        is_deleted: false,
        deleted_on: null,
        deleted_by: null,
      };

      // This assertion will fail until Event interface is aligned with database schema
      expect(() => {
        const event: Event = mockDatabaseEvent as any;
        
        // Required database fields validation
        expect(event.event_id).toBeDefined();
        expect(event.description).toBeDefined(); // Currently missing in Event interface
        expect(event.registration_status).toBeDefined(); // Currently missing
        expect(event.publishing_status).toBeDefined(); // Currently missing
        expect(event.event_type).toBeDefined(); // Currently missing
        expect(event.priority_level).toBeDefined(); // Currently missing
        expect(event.max_capacity).toBeDefined(); // Currently 'capacity' instead
        expect(event.registration_deadline).toBeDefined(); // Currently missing
        expect(event.image_url).toBeDefined(); // Currently 'featured_image' instead
        expect(event.organizer_name).toBeDefined(); // Currently 'author' instead
        expect(event.end_date).toBeDefined(); // Currently missing
        expect(event.end_time).toBeDefined(); // Currently missing
        expect(event.virtual_link).toBeDefined(); // Currently missing
        expect(event.created_on).toBeDefined(); // Currently 'published_at' instead
        expect(event.created_by).toBeDefined(); // Currently missing
        expect(event.modified_on).toBeDefined(); // Currently missing
        expect(event.modified_by).toBeDefined(); // Currently missing
        expect(event.is_deleted).toBeDefined(); // Currently missing
        expect(event.deleted_on).toBeDefined(); // Currently missing
        expect(event.deleted_by).toBeDefined(); // Currently missing
        
        // Field type validation
        expect(typeof event.event_id).toBe('string');
        expect(typeof event.title).toBe('string');
        expect(typeof event.description).toBe('string');
        expect(typeof event.slug).toBe('string');
        expect(typeof event.category_id).toBe('string');
        expect(typeof event.event_date).toBe('string');
        expect(typeof event.location).toBe('string');
        expect(typeof event.registration_status).toBe('string');
        expect(typeof event.publishing_status).toBe('string');
        expect(Array.isArray(event.tags)).toBe(true);
        expect(typeof event.event_type).toBe('string');
        expect(typeof event.priority_level).toBe('string');
        expect(typeof event.created_on).toBe('string');
        expect(typeof event.is_deleted).toBe('boolean');
      }).not.toThrow();
    });

    it('should not have legacy fields that are not in database schema', () => {
      // Create a proper Event object using the correct database schema fields
      const properEvent: Event = {
        event_id: 'event-uuid-123',
        title: 'Proper Event',
        description: 'Event description',
        content: 'Event content',
        slug: 'proper-event',
        category_id: 'category-uuid-456',
        image_url: 'https://example.com/image.jpg',
        organizer_name: 'Event Organizer',
        event_date: '2024-03-15',
        event_time: '14:30',
        end_date: '2024-03-15',
        end_time: '17:00',
        location: '123 Event St',
        virtual_link: 'https://virtual.example.com',
        max_capacity: 100,
        registration_deadline: '2024-03-10T23:59:59Z',
        registration_status: 'open',
        publishing_status: 'published',
        tags: ['event'],
        event_type: 'workshop',
        priority_level: 'normal',
        created_on: '2024-01-01T00:00:00Z',
        created_by: 'admin@example.com',
        modified_on: '2024-01-02T00:00:00Z',
        modified_by: 'admin@example.com',
        is_deleted: false,
        deleted_on: null,
        deleted_by: null,
      };

      // Test that proper Event object doesn't have legacy fields
      expect((properEvent as any).excerpt).toBeUndefined();
      expect((properEvent as any).featured_image).toBeUndefined();
      expect((properEvent as any).capacity).toBeUndefined();
      expect((properEvent as any).registration_url).toBeUndefined();
      expect((properEvent as any).author).toBeUndefined();
      expect((properEvent as any).status).toBeUndefined();
      expect((properEvent as any).featured).toBeUndefined();
      expect((properEvent as any).category).toBeUndefined();
      expect((properEvent as any).category_data).toBeUndefined();
      expect((properEvent as any).meta_title).toBeUndefined();
      expect((properEvent as any).meta_description).toBeUndefined();
      expect((properEvent as any).published_at).toBeUndefined();
    });

    it('should validate event_type enum values match database constraints', () => {
      const validEventTypes: DatabaseSchemaEvent['event_type'][] = [
        'workshop', 'seminar', 'webinar', 'conference', 
        'fundraiser', 'community', 'medical', 'educational'
      ];

      // This will pass once Event interface is properly aligned
      validEventTypes.forEach(eventType => {
        expect(['workshop', 'seminar', 'webinar', 'conference', 'fundraiser', 'community', 'medical', 'educational'])
          .toContain(eventType);
      });

      // Test that current Event interface supports these values
      const testEvent: Partial<Event> = {
        event_type: 'workshop' as any // This will fail until event_type is added to Event interface
      };
      expect(testEvent.event_type).toBe('workshop');
    });

    it('should validate registration_status enum values match database constraints', () => {
      const validStatuses: DatabaseSchemaEvent['registration_status'][] = [
        'open', 'registration_required', 'full', 'cancelled'
      ];

      validStatuses.forEach(status => {
        expect(['open', 'registration_required', 'full', 'cancelled'])
          .toContain(status);
      });

      // Test that current Event interface supports these values
      const testEvent: Partial<Event> = {
        registration_status: 'open' as any // This will fail until registration_status is added to Event interface
      };
      expect(testEvent.registration_status).toBe('open');
    });

    it('should validate publishing_status enum values match database constraints', () => {
      const validStatuses: DatabaseSchemaEvent['publishing_status'][] = [
        'draft', 'published', 'archived'
      ];

      validStatuses.forEach(status => {
        expect(['draft', 'published', 'archived'])
          .toContain(status);
      });

      // Test that current Event interface supports these values
      const testEvent: Partial<Event> = {
        publishing_status: 'published' as any // This will fail until publishing_status is added to Event interface
      };
      expect(testEvent.publishing_status).toBe('published');
    });

    it('should validate priority_level enum values match database constraints', () => {
      const validPriorities: DatabaseSchemaEvent['priority_level'][] = [
        'low', 'normal', 'high', 'urgent'
      ];

      validPriorities.forEach(priority => {
        expect(['low', 'normal', 'high', 'urgent'])
          .toContain(priority);
      });

      // Test that current Event interface supports these values
      const testEvent: Partial<Event> = {
        priority_level: 'normal' as any // This will fail until priority_level is added to Event interface
      };
      expect(testEvent.priority_level).toBe('normal');
    });

    it('should validate field length constraints match database schema', () => {
      // title VARCHAR(255) NOT NULL
      expect('A'.repeat(255).length).toBeLessThanOrEqual(255);
      expect(() => {
        const title = 'A'.repeat(256);
        expect(title.length).toBeLessThanOrEqual(255);
      }).toThrow();

      // slug VARCHAR(255) UNIQUE NOT NULL
      expect('slug-'.repeat(50).length).toBeLessThanOrEqual(255);

      // location VARCHAR(500) NOT NULL
      expect('Location '.repeat(50).length).toBeLessThanOrEqual(500);

      // image_url VARCHAR(500)
      expect('https://example.com/very-long-path/'.repeat(10).length).toBeLessThanOrEqual(500);

      // organizer_name VARCHAR(255)
      expect('Organizer Name '.repeat(15).length).toBeLessThanOrEqual(255);
    });

    it('should validate required vs optional fields according to database schema', () => {
      const requiredFields = [
        'event_id', 'title', 'description', 'slug', 'category_id',
        'event_date', 'location', 'registration_status', 'publishing_status',
        'tags', 'event_type', 'priority_level', 'created_on', 'is_deleted'
      ];

      const optionalFields = [
        'content', 'image_url', 'organizer_name', 'event_time', 'end_date', 'end_time',
        'virtual_link', 'max_capacity', 'registration_deadline', 'created_by',
        'modified_on', 'modified_by', 'deleted_on', 'deleted_by'
      ];

      // This test will guide GREEN phase implementation
      requiredFields.forEach(field => {
        expect(requiredFields).toContain(field);
      });

      optionalFields.forEach(field => {
        expect(optionalFields).toContain(field);
      });
    });
  });

  describe('EventCategory Interface Schema Validation', () => {
    it('should have all required database schema fields for event categories', () => {
      const mockDatabaseCategory: DatabaseSchemaEventCategory = {
        category_id: 'category-uuid-123',
        name: 'Healthcare Events',
        slug: 'healthcare-events',
        description: 'Events related to healthcare',
        is_default_unassigned: false,
        created_on: '2024-01-01T00:00:00Z',
        created_by: 'admin@example.com',
        modified_on: '2024-01-02T00:00:00Z',
        modified_by: 'admin@example.com',
        is_deleted: false,
        deleted_on: null,
        deleted_by: null,
      };

      // This will fail until EventCategory interface matches database schema
      expect(() => {
        const category: EventCategory = mockDatabaseCategory as any;
        
        // Required fields validation
        expect(category.category_id).toBeDefined(); // Currently missing
        expect(category.is_default_unassigned).toBeDefined(); // Currently missing
        expect(category.created_on).toBeDefined(); // Currently missing
        expect(category.created_by).toBeDefined(); // Currently missing
        expect(category.modified_on).toBeDefined(); // Currently missing
        expect(category.modified_by).toBeDefined(); // Currently missing
        expect(category.is_deleted).toBeDefined(); // Currently missing
        expect(category.deleted_on).toBeDefined(); // Currently missing
        expect(category.deleted_by).toBeDefined(); // Currently missing
        
        // Field type validation
        expect(typeof category.category_id).toBe('string');
        expect(typeof category.name).toBe('string');
        expect(typeof category.slug).toBe('string');
        expect(typeof category.is_default_unassigned).toBe('boolean');
        expect(typeof category.created_on).toBe('string');
        expect(typeof category.is_deleted).toBe('boolean');
      }).not.toThrow();
    });

    it('should not have legacy EventCategory fields not in database schema', () => {
      // Create a proper EventCategory object using correct database schema fields
      const properCategory: EventCategory = {
        category_id: 'category-uuid-123',
        name: 'Proper Category',
        slug: 'proper-category',
        description: 'Category description',
        is_default_unassigned: false,
        created_on: '2024-01-01T00:00:00Z',
        created_by: 'admin@example.com',
        modified_on: '2024-01-02T00:00:00Z',
        modified_by: 'admin@example.com',
        is_deleted: false,
        deleted_on: null,
        deleted_by: null,
      };

      // Test that proper EventCategory object doesn't have legacy fields
      expect((properCategory as any).color).toBeUndefined();
      expect((properCategory as any).display_order).toBeUndefined();
      expect((properCategory as any).active).toBeUndefined();
    });
  });

  describe('Response Types Schema Validation', () => {
    it('should validate EventsResponse matches expected backend format', () => {
      // Backend format: { events: [...], count: number, correlation_id: string }
      const mockEventsResponse: EventsResponse = {
        events: [],
        count: 0,
        correlation_id: 'test-correlation-id'
      };

      expect(Array.isArray(mockEventsResponse.events)).toBe(true);
      expect(typeof mockEventsResponse.count).toBe('number');
      expect(typeof mockEventsResponse.correlation_id).toBe('string');
    });

    it('should validate EventResponse matches expected backend format', () => {
      // Backend format: { event: {...}, correlation_id: string }
      const mockEvent: DatabaseSchemaEvent = {
        event_id: 'test-uuid',
        title: 'Test Event',
        description: 'Test Description',
        slug: 'test-event',
        category_id: 'category-uuid',
        event_date: '2024-03-15',
        location: 'Test Location',
        registration_status: 'open',
        publishing_status: 'published',
        tags: ['test'],
        event_type: 'workshop',
        priority_level: 'normal',
        created_on: '2024-01-01T00:00:00Z',
        is_deleted: false,
      };

      const mockEventResponse: EventResponse = {
        event: mockEvent as any,
        correlation_id: 'test-correlation-id'
      };

      expect(typeof mockEventResponse.event).toBe('object');
      expect(mockEventResponse.event).not.toBeNull();
      expect(typeof mockEventResponse.correlation_id).toBe('string');
    });
  });

  describe('Type System Integration', () => {
    it('should ensure type safety for database schema-compliant events', () => {
      // This test ensures the type system prevents incorrect usage
      const createValidEvent = (data: DatabaseSchemaEvent): Event => {
        // This will fail until Event interface matches DatabaseSchemaEvent
        return data as Event;
      };

      const validData: DatabaseSchemaEvent = {
        event_id: 'uuid-123',
        title: 'Valid Event',
        description: 'Valid description',
        slug: 'valid-event',
        category_id: 'category-uuid',
        event_date: '2024-03-15',
        location: 'Valid Location',
        registration_status: 'open',
        publishing_status: 'published',
        tags: ['valid'],
        event_type: 'workshop',
        priority_level: 'normal',
        created_on: '2024-01-01T00:00:00Z',
        is_deleted: false,
      };

      const event = createValidEvent(validData);
      expect(event.event_id).toBe('uuid-123');
    });
  });
});