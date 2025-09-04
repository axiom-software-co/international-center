// Events Composables - Vue 3 Composition API composables for events data
// Provides clean interface for Vue components to interact with events domain

import { ref, computed, onMounted, watch, type Ref } from 'vue';
import { eventsClient } from '../lib/clients';
import type { Event, GetEventsParams } from '../lib/clients';

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

      // Backend returns: { events: [...], count: number, correlation_id: string }
      const response = await eventsClient.getEvents(params);

      if (response.events) {
        events.value = response.events;
        total.value = response.count || response.events.length;
        // Calculate pagination from current params since backend doesn't return pagination info
        page.value = params.page || 1;
        pageSize.value = params.pageSize || 10;
        totalPages.value = Math.ceil(total.value / pageSize.value);
      } else {
        throw new Error('Failed to fetch events');
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

      // Backend returns: { event: {...}, correlation_id: string }
      const response = await eventsClient.getEventBySlug(slugRef.value);
      
      if (response.event) {
        event.value = response.event;
      } else {
        throw new Error('Failed to fetch event');
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

export interface UseFeaturedEventsResult {
  events: Ref<Event[]>;
  loading: Ref<boolean>;
  error: Ref<string | null>;
  refetch: () => Promise<void>;
}

export function useFeaturedEvents(limit?: Ref<number> | number): UseFeaturedEventsResult {
  const limitRef = typeof limit === 'number' ? ref(limit) : limit;
  
  const events = ref<Event[]>([]);
  const loading = ref(true);
  const error = ref<string | null>(null);

  const fetchFeaturedEvents = async () => {
    try {
      loading.value = true;
      error.value = null;

      // Backend returns: { events: [...], count: number, correlation_id: string }
      const response = await eventsClient.getFeaturedEvents(limitRef?.value);
      
      if (response.events) {
        events.value = response.events;
      } else {
        throw new Error('Failed to fetch featured events');
      }
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : 'Failed to fetch featured events';
      error.value = errorMessage;
      console.error('Error fetching featured events:', err);
      events.value = [];
    } finally {
      loading.value = false;
    }
  };

  // Watch for limit changes
  if (limitRef) {
    watch(limitRef, fetchFeaturedEvents, { immediate: true });
  } else {
    onMounted(fetchFeaturedEvents);
  }

  return {
    events,
    loading,
    error,
    refetch: fetchFeaturedEvents,
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

      // Backend returns: { events: [...], count: number, correlation_id: string }
      const response = await eventsClient.searchEvents({
        q: query,
        page: options.page || 1,
        pageSize: options.pageSize || 10,
        category: options.category,
        ...options,
      });

      if (response.events) {
        results.value = response.events;
        total.value = response.count || response.events.length;
        page.value = options.page || 1;
        pageSize.value = options.pageSize || 10;
        totalPages.value = Math.ceil(total.value / pageSize.value);
      } else {
        throw new Error('Failed to search events');
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