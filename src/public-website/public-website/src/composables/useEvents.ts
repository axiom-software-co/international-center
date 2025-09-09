// Events Composables - Vue 3 Composition API with Store Delegation

import { ref, computed, watch, onMounted, isRef, unref, type Ref } from 'vue';
import { storeToRefs } from 'pinia';
import { useEventsStore } from '../stores/events';
import type { Event, EventCategory, GetEventsParams, SearchEventsParams } from '../lib/clients/events/types';

// Domain-specific type aliases
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

// Main events list composable - delegates to store
export const useEvents = (options: UseEventsOptions = {}): UseEventsResult => {
  const { enabled = true, immediate = true, ...params } = options;
  const store = useEventsStore();
  const { events, loading, error, total, totalPages } = storeToRefs(store);
  
  // Local refs for pagination
  const page = ref(params.page || 1);
  const pageSize = ref(params.pageSize || 10);

  const fetchEvents = async () => {
    if (!enabled) return;
    await store.fetchEvents(params);
  };

  // Watch for parameter changes
  watch(() => params, fetchEvents, { deep: true });
  
  // Call immediately if enabled and immediate is true
  if (enabled && immediate) {
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
};

export interface UseEventResult {
  event: Ref<Event | null>;
  loading: Ref<boolean>;
  error: Ref<string | null>;
  refetch: () => Promise<void>;
}

// Single event composable - delegates to store
export const useEvent = (slug: Ref<string | null> | string | null): UseEventResult => {
  const slugRef = isRef(slug) ? slug : ref(slug);
  const store = useEventsStore();
  const { event, loading, error } = storeToRefs(store);

  const fetchEvent = async () => {
    const currentSlug = unref(slugRef);
    if (!currentSlug) {
      store.event = null;
      return;
    }

    await store.fetchEvent(currentSlug);
  };

  // Watch for slug changes and call immediately
  watch(slugRef, (newSlug) => {
    if (newSlug) {
      fetchEvent();
    } else {
      store.event = null;
    }
  }, { immediate: true });

  return {
    event,
    loading,
    error,
    refetch: fetchEvent,
  };
};

export interface UseFeaturedEventsResult {
  events: Ref<Event[]>;
  loading: Ref<boolean>;
  error: Ref<string | null>;
  refetch: () => Promise<void>;
}

// Featured events composable - delegates to store
export const useFeaturedEvents = (limit?: Ref<number | undefined> | number | undefined): UseFeaturedEventsResult => {
  const limitRef = isRef(limit) ? limit : ref(limit);
  const store = useEventsStore();
  const { featuredEvents: events, loading, error } = storeToRefs(store);

  const fetchFeaturedEvents = async () => {
    await store.fetchFeaturedEvents(unref(limitRef));
  };

  // Trigger initial fetch immediately
  fetchFeaturedEvents();
  
  // Watch for limit changes
  watch(limitRef, fetchFeaturedEvents);

  return {
    events,
    loading,
    error,
    refetch: fetchFeaturedEvents,
  };
};

// Legacy compatibility - single featured event
export function useFeaturedEvent(): { event: Ref<Event | null>; loading: Ref<boolean>; error: Ref<string | null>; refetch: () => Promise<void> } {
  const { events, loading, error, refetch } = useFeaturedEvents(1);
  const event = computed(() => events.value.length > 0 ? events.value[0] : null);
  
  return {
    event,
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
  search: (query: string, options?: Partial<SearchEventsParams>) => Promise<void>;
}

// Search events composable - delegates to store
export const useSearchEvents = (): UseSearchEventsResult => {
  const store = useEventsStore();
  const { searchResults: results, loading, error, searchTotal: total } = storeToRefs(store);
  
  // Local refs for search-specific pagination
  const page = ref(1);
  const pageSize = ref(10);

  const totalPages = computed(() => Math.ceil(total.value / pageSize.value) || 0);

  const search = async (query: string, options: Partial<SearchEventsParams> = {}) => {
    const searchParams = {
      q: query,
      page: options.page || 1,
      pageSize: options.pageSize || 10,
      ...options,
    };

    // Update local pagination refs
    page.value = searchParams.page;
    pageSize.value = searchParams.pageSize;

    await store.searchEvents(searchParams);
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
};

export interface UseEventCategoriesResult {
  categories: Ref<EventCategory[]>;
  loading: Ref<boolean>;
  error: Ref<string | null>;
  refetch: () => Promise<void>;
}

// Event categories composable - delegates to store
export const useEventCategories = (): UseEventCategoriesResult => {
  const store = useEventsStore();
  const { categories, loading, error } = storeToRefs(store);

  const fetchCategories = async () => {
    await store.fetchEventCategories();
  };

  // Trigger initial fetch immediately
  fetchCategories();

  return {
    categories,
    loading,
    error,
    refetch: fetchCategories,
  };
};
