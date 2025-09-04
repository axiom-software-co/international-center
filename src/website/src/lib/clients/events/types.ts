// Events Domain Types - Database Schema Compliant
// Aligned with TABLES-EVENTS.md PostgreSQL schema

import {
  PaginationParams,
  FilterParams,
  StandardRestResponse,
  SingleRestResponse,
} from '../rest/types';

// Enums matching database constraints
export type RegistrationStatus = 'open' | 'registration_required' | 'full' | 'cancelled';
export type PublishingStatus = 'draft' | 'published' | 'archived';
export type EventType = 'workshop' | 'seminar' | 'webinar' | 'conference' | 'fundraiser' | 'community' | 'medical' | 'educational';
export type PriorityLevel = 'low' | 'normal' | 'high' | 'urgent';

// Database schema compliant interfaces
export interface EventCategory {
  category_id: string; // UUID
  name: string;
  slug: string;
  description?: string; // TEXT, can be NULL
  is_default_unassigned: boolean;
  
  // Audit fields
  created_on: string; // TIMESTAMPTZ as ISO string
  created_by?: string;
  modified_on?: string; // TIMESTAMPTZ as ISO string
  modified_by?: string;
  
  // Soft delete fields
  is_deleted: boolean;
  deleted_on?: string; // TIMESTAMPTZ as ISO string
  deleted_by?: string;
}

export interface Event {
  event_id: string; // UUID primary key
  title: string; // VARCHAR(255) NOT NULL
  description: string; // TEXT NOT NULL
  content?: string; // TEXT, can be NULL
  slug: string; // VARCHAR(255) UNIQUE NOT NULL
  category_id: string; // UUID NOT NULL
  image_url?: string; // VARCHAR(500) - Azure Blob Storage URL
  organizer_name?: string; // VARCHAR(255)
  event_date: string; // DATE NOT NULL as YYYY-MM-DD
  event_time?: string; // TIME as HH:MM:SS
  end_date?: string; // DATE as YYYY-MM-DD
  end_time?: string; // TIME as HH:MM:SS
  location: string; // VARCHAR(500) NOT NULL
  virtual_link?: string; // VARCHAR(500) for virtual events
  max_capacity?: number; // INTEGER
  registration_deadline?: string; // TIMESTAMPTZ as ISO string
  registration_status: RegistrationStatus;
  publishing_status: PublishingStatus;
  
  // Content metadata
  tags?: string[]; // TEXT[]
  event_type: EventType;
  priority_level: PriorityLevel;
  
  // Audit fields
  created_on: string; // TIMESTAMPTZ NOT NULL as ISO string
  created_by?: string;
  modified_on?: string; // TIMESTAMPTZ as ISO string
  modified_by?: string;
  
  // Soft delete fields
  is_deleted: boolean;
  deleted_on?: string; // TIMESTAMPTZ as ISO string
  deleted_by?: string;
  
  // Joined category data (when included)
  category_data?: EventCategory;
}

// Featured events relationship
export interface FeaturedEvent {
  featured_event_id: string; // UUID
  event_id: string; // UUID
  created_on: string; // TIMESTAMPTZ
  created_by?: string;
  modified_on?: string; // TIMESTAMPTZ
  modified_by?: string;
}

// API Response types
export type EventsResponse = StandardRestResponse<Event>;
export type EventResponse = SingleRestResponse<Event>;
export type EventCategoriesResponse = SingleRestResponse<EventCategory[]>;
export type FeaturedEventResponse = SingleRestResponse<FeaturedEvent>;

// Query parameter interfaces
export interface GetEventsParams extends PaginationParams, FilterParams {
  category_id?: string; // UUID filter
  event_type?: EventType;
  publishing_status?: PublishingStatus;
  registration_status?: RegistrationStatus;
  priority_level?: PriorityLevel;
  event_date_from?: string; // YYYY-MM-DD
  event_date_to?: string; // YYYY-MM-DD
  organizer_name?: string;
  featured?: boolean; // Query for featured events
  sortBy?: 'event_date_asc' | 'event_date_desc' | 'created_on_asc' | 'created_on_desc' | 'title_asc' | 'title_desc' | 'priority_level';
}

export interface SearchEventsParams extends PaginationParams {
  q: string; // Search query
  category_id?: string;
  event_type?: EventType;
  publishing_status?: PublishingStatus;
  event_date_from?: string;
  event_date_to?: string;
  sortBy?: string;
}

export interface CreateEventParams {
  title: string;
  description: string;
  content?: string;
  slug: string;
  category_id: string;
  image_url?: string;
  organizer_name?: string;
  event_date: string;
  event_time?: string;
  end_date?: string;
  end_time?: string;
  location: string;
  virtual_link?: string;
  max_capacity?: number;
  registration_deadline?: string;
  registration_status?: RegistrationStatus;
  event_type: EventType;
  priority_level?: PriorityLevel;
  tags?: string[];
}

export interface UpdateEventParams extends Partial<CreateEventParams> {
  event_id: string;
}

// Event registration interfaces
export interface EventRegistration {
  registration_id: string;
  event_id: string;
  participant_name: string;
  participant_email: string;
  participant_phone?: string;
  registration_timestamp: string;
  registration_status: 'registered' | 'confirmed' | 'cancelled' | 'no_show';
  special_requirements?: string;
  dietary_restrictions?: string;
  accessibility_needs?: string;
  
  // Audit fields
  created_on: string;
  created_by?: string;
  modified_on?: string;
  modified_by?: string;
  
  // Soft delete fields
  is_deleted: boolean;
  deleted_on?: string;
  deleted_by?: string;
}