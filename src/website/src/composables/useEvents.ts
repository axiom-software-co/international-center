// Events Composables - Database Schema Compliant
// Updated to work with TABLES-EVENTS.md aligned types

import { ref, computed, onMounted, watch, type Ref } from 'vue';
import { eventsClient } from '../lib/clients';
import type { Event, GetEventsParams, FeaturedEvent } from '../lib/clients';

export interface UseEventsResult {
  events: Ref<Event[]>;
  loading: Ref<boolean>;
  error: Ref<string | null>;
  total: Ref<number>;
  page: Ref<number>;
  pageSize: Ref<number>;
  totalPages: Ref<number>;
  refetch: () => Promise<void>;
}

export interface UseEventsOptions extends GetEventsParams {
  enabled?: boolean;
  immediate?: boolean;
}

export function useEvents(options: UseEventsOptions = {}): UseEventsResult {
  const { enabled = true, immediate = true, ...params } = options;

  const events = ref<Event[]>([]);
  const loading = ref(enabled && immediate);
  const error = ref<string | null>(null);
  const total = ref(0);
  const page = ref(params.page || 1);
  const pageSize = ref(params.pageSize || 10);
  const totalPages = ref(0);

  const fetchEvents = async () => {
    if (!enabled) return;

    try {
      loading.value = true;
      error.value = null;

      // Backend returns: { data: [...], pagination: {...}, success: boolean, message?, errors? }
      const response = await eventsClient.getEvents(params);

      if (response.success && response.data) {
        events.value = response.data;
        total.value = response.pagination.total;
        page.value = response.pagination.page;
        pageSize.value = response.pagination.pageSize;
        totalPages.value = response.pagination.totalPages;
      } else {
        throw new Error(response.message || 'Failed to fetch events');
      }
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : 'Failed to fetch events';
      error.value = errorMessage;
      console.error('Error fetching events:', err);
      events.value = [];
      total.value = 0;
      totalPages.value = 0;
    } finally {
      loading.value = false;
    }
  };

  // Watch for parameter changes
  watch(() => params, fetchEvents, { deep: true });
  
  if (immediate) {
    onMounted(fetchEvents);
  }

  return {
    events,
    loading,
    error,
    total,
    page,
    pageSize,
    totalPages,
    refetch: fetchEvents,
  };
}

export interface UseEventResult {
  event: Ref<Event | null>;
  loading: Ref<boolean>;
  error: Ref<string | null>;
  refetch: () => Promise<void>;
}

export function useEvent(slug: Ref<string | null> | string | null): UseEventResult {
  const slugRef = typeof slug === 'string' ? ref(slug) : slug || ref(null);
  
  const event = ref<Event | null>(null);
  const loading = ref(!!slugRef.value);
  const error = ref<string | null>(null);

  const fetchEvent = async () => {
    if (!slugRef.value) {
      event.value = null;
      loading.value = false;
      return;
    }

    try {
      loading.value = true;
      error.value = null;

      // Backend returns: { data: {...}, success: boolean, message?, errors? }
      const response = await eventsClient.getEventBySlug(slugRef.value);
      
      if (response.success && response.data) {
        event.value = response.data;
      } else {
        throw new Error(response.message || 'Failed to fetch event');
      }
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : 'Failed to fetch event';
      error.value = errorMessage;
      console.error('Error fetching event:', err);
      event.value = null;
    } finally {
      loading.value = false;
    }
  };

  // Watch for slug changes
  watch(slugRef, fetchEvent, { immediate: true });

  return {
    event,
    loading,
    error,
    refetch: fetchEvent,
  };
}

export interface UseFeaturedEventResult {
  featuredEvent: Ref<FeaturedEvent | null>;
  event: Ref<Event | null>;
  loading: Ref<boolean>;
  error: Ref<string | null>;
  refetch: () => Promise<void>;
}

export function useFeaturedEvent(): UseFeaturedEventResult {
  const featuredEvent = ref<FeaturedEvent | null>(null);
  const event = ref<Event | null>(null);
  const loading = ref(true);
  const error = ref<string | null>(null);

  const fetchFeaturedEvent = async () => {
    try {
      loading.value = true;
      error.value = null;

      // Backend returns: { data: {...}, success: boolean, message?, errors? }
      const response = await eventsClient.getFeaturedEvents();
      
      if (response.success && response.data) {
        featuredEvent.value = response.data;
        // Optionally fetch the full event details if needed
        if (response.data.event_id) {
          const eventResponse = await eventsClient.getEventById(response.data.event_id);
          if (eventResponse.success && eventResponse.data) {
            event.value = eventResponse.data;
          }
        }
      } else {
        // No featured event is not an error, just no data
        featuredEvent.value = null;
        event.value = null;
      }
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : 'Failed to fetch featured event';
      error.value = errorMessage;
      console.error('Error fetching featured event:', err);
      featuredEvent.value = null;
      event.value = null;
    } finally {
      loading.value = false;
    }
  };

  onMounted(fetchFeaturedEvent);

  return {
    featuredEvent,
    event,
    loading,
    error,
    refetch: fetchFeaturedEvent,
  };
}

// Legacy compatibility function
export function useFeaturedEvents(): { events: Ref<Event[]>; loading: Ref<boolean>; error: Ref<string | null>; refetch: () => Promise<void> } {
  const { event, loading, error, refetch } = useFeaturedEvent();
  const events = computed(() => event.value ? [event.value] : []);
  
  return {
    events,
    loading,
    error,
    refetch,
  };
}

export interface UseSearchEventsResult {
  results: Ref<Event[]>;
  loading: Ref<boolean>;
  error: Ref<string | null>;
  total: Ref<number>;
  page: Ref<number>;
  pageSize: Ref<number>;
  totalPages: Ref<number>;
  search: (query: string, options?: Partial<GetEventsParams>) => Promise<void>;
}

export function useSearchEvents(): UseSearchEventsResult {
  const results = ref<Event[]>([]);
  const loading = ref(false);
  const error = ref<string | null>(null);
  const total = ref(0);
  const page = ref(1);
  const pageSize = ref(10);
  const totalPages = ref(0);

  const search = async (query: string, options: Partial<GetEventsParams> = {}) => {
    if (!query.trim()) {
      results.value = [];
      total.value = 0;
      totalPages.value = 0;
      return;
    }

    try {
      loading.value = true;
      error.value = null;

      // Backend returns: { data: [...], pagination: {...}, success: boolean, message?, errors? }
      const response = await eventsClient.searchEvents({
        q: query,
        page: options.page || 1,
        pageSize: options.pageSize || 10,
        category_id: options.category_id,
        event_type: options.event_type,
        publishing_status: options.publishing_status,
        event_date_from: options.event_date_from,
        event_date_to: options.event_date_to,
        ...options,
      });

      if (response.success && response.data) {
        results.value = response.data;
        total.value = response.pagination.total;
        page.value = response.pagination.page;
        pageSize.value = response.pagination.pageSize;
        totalPages.value = response.pagination.totalPages;
      } else {
        throw new Error(response.message || 'Failed to search events');
      }
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : 'Failed to search events';
      error.value = errorMessage;
      console.error('Error searching events:', err);
      results.value = [];
      total.value = 0;
      totalPages.value = 0;
    } finally {
      loading.value = false;
    }
  };

  return {
    results,
    loading,
    error,
    total,
    page,
    pageSize,
    totalPages,
    search,
  };
}