// Events-specific filtering composable
// Integrates events data with generic filtering system

import { ref, computed, watch, type Ref } from 'vue';
import { useEvents } from '../useEvents';
import { useFiltering, type FilterConfig, type FilterOption } from './useFiltering';
import type { Event } from '../../lib/clients/events/types';

export interface UseEventsFilteringOptions {
  enableTypeFilter?: boolean;
  enableStatusFilter?: boolean;
  enableDateFilter?: boolean;
  includeCounts?: boolean;
}

export interface UseEventsFilteringResult {
  // Filtered events data
  filteredEvents: Ref<Event[]>;
  
  // Filter state
  filters: Ref<{ [key: string]: string | string[] }>;
  hasActiveFilters: Ref<boolean>;
  
  // Filter configurations for UI
  typeFilterConfig: Ref<FilterConfig | null>;
  statusFilterConfig: Ref<FilterConfig | null>;
  dateFilterConfig: Ref<FilterConfig | null>;
  
  // Filter actions
  setTypeFilter: (type: string) => void;
  setStatusFilter: (status: string) => void;
  setDateFilter: (dateRange: string) => void;
  clearAllFilters: () => void;
  
  // Utility methods
  isFilterActive: (filterName: string, value?: string) => boolean;
}

// Event type labels mapping
const EVENT_TYPE_LABELS = {
  conference: 'Conference',
  workshop: 'Workshop',
  seminar: 'Seminar',
  webinar: 'Webinar',
  community_event: 'Community Event',
  training: 'Training',
} as const;

// Event status labels mapping
const STATUS_LABELS = {
  published: 'Published',
  draft: 'Draft',
  archived: 'Archived',
} as const;

// Date range options
const DATE_RANGE_LABELS = {
  upcoming: 'Upcoming',
  this_month: 'This Month',
  next_month: 'Next Month',
  past: 'Past Events',
} as const;

export function useEventsFiltering(
  eventsOptions: Parameters<typeof useEvents>[0] = {},
  filterOptions: UseEventsFilteringOptions = {}
): UseEventsFilteringResult {

  const {
    enableTypeFilter = true,
    enableStatusFilter = true,
    enableDateFilter = true,
    includeCounts = true,
  } = filterOptions;

  // Load events data
  const { events, loading: eventsLoading } = useEvents(eventsOptions);

  // Create filter configurations
  const filterConfigs: FilterConfig[] = [];

  // Event type filter configuration
  const typeFilterConfig = computed(() => {
    if (!enableTypeFilter || eventsLoading.value) return null;

    const typeOptions: FilterOption[] = Object.entries(EVENT_TYPE_LABELS).map(([type, label]) => ({
      value: type,
      label,
      count: includeCounts ? countEventsByType(events.value, type) : undefined,
    }));

    return {
      name: 'event_type',
      label: 'Event Type',
      options: typeOptions,
      multiple: false,
      searchable: true,
    };
  });

  // Status filter configuration  
  const statusFilterConfig = computed(() => {
    if (!enableStatusFilter || eventsLoading.value) return null;

    const statusOptions: FilterOption[] = Object.entries(STATUS_LABELS).map(([status, label]) => ({
      value: status,
      label,
      count: includeCounts ? countEventsByStatus(events.value, status) : undefined,
    }));

    return {
      name: 'publishing_status',
      label: 'Status',
      options: statusOptions,
      multiple: false,
      searchable: false,
    };
  });

  // Date range filter configuration
  const dateFilterConfig = computed(() => {
    if (!enableDateFilter || eventsLoading.value) return null;

    const dateOptions: FilterOption[] = Object.entries(DATE_RANGE_LABELS).map(([range, label]) => ({
      value: range,
      label,
      count: includeCounts ? countEventsByDateRange(events.value, range) : undefined,
    }));

    return {
      name: 'date_range',
      label: 'Date Range',
      options: dateOptions,
      multiple: false,
      searchable: false,
    };
  });

  // Build dynamic filter configs
  watch([typeFilterConfig, statusFilterConfig, dateFilterConfig], ([typeConfig, statusConfig, dateConfig]) => {
    filterConfigs.length = 0; // Clear existing configs
    
    if (typeConfig) {
      filterConfigs.push(typeConfig);
    }
    if (statusConfig) {
      filterConfigs.push(statusConfig);
    }
    if (dateConfig) {
      filterConfigs.push(dateConfig);
    }
  }, { immediate: true });

  // Initialize filtering system
  const filteringResult = useFiltering(events, filterConfigs);

  // Convenient filter setters
  const setTypeFilter = (type: string): void => {
    filteringResult.setFilter('event_type', type);
  };

  const setStatusFilter = (status: string): void => {
    filteringResult.setFilter('publishing_status', status);
  };

  const setDateFilter = (dateRange: string): void => {
    filteringResult.setFilter('date_range', dateRange);
  };

  return {
    filteredEvents: filteringResult.filteredItems,
    filters: filteringResult.filters,
    hasActiveFilters: filteringResult.hasActiveFilters,
    typeFilterConfig,
    statusFilterConfig,
    dateFilterConfig,
    setTypeFilter,
    setStatusFilter,
    setDateFilter,
    clearAllFilters: filteringResult.clearAllFilters,
    isFilterActive: filteringResult.isFilterActive,
  };
}

// Helper functions
function countEventsByType(events: Event[], type: string): number {
  return events.filter(event => event.event_type === type).length;
}

function countEventsByStatus(events: Event[], status: string): number {
  return events.filter(event => event.publishing_status === status).length;
}

function countEventsByDateRange(events: Event[], dateRange: string): number {
  const now = new Date();
  const startOfMonth = new Date(now.getFullYear(), now.getMonth(), 1);
  const endOfMonth = new Date(now.getFullYear(), now.getMonth() + 1, 0);
  const startOfNextMonth = new Date(now.getFullYear(), now.getMonth() + 1, 1);
  const endOfNextMonth = new Date(now.getFullYear(), now.getMonth() + 2, 0);

  return events.filter(event => {
    const eventDate = new Date(event.event_date);
    
    switch (dateRange) {
      case 'upcoming':
        return eventDate >= now;
      case 'this_month':
        return eventDate >= startOfMonth && eventDate <= endOfMonth;
      case 'next_month':
        return eventDate >= startOfNextMonth && eventDate <= endOfNextMonth;
      case 'past':
        return eventDate < now;
      default:
        return true;
    }
  }).length;
}

// Type-only filtering composable for simpler use cases
export function useEventsTypeFiltering(
  eventsOptions: Parameters<typeof useEvents>[0] = {}
): UseEventsFilteringResult {
  return useEventsFiltering(eventsOptions, {
    enableTypeFilter: true,
    enableStatusFilter: false,
    enableDateFilter: false,
    includeCounts: true,
  });
}