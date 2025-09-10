import { defineStore } from 'pinia';
import { apiClient } from '../lib/api-client';
import type { 
  Event, 
  EventCategory,
  GetEventsRequest
} from '@international-center/public-api-client';
import type { 
  EventsStoreState, 
  EventsStoreActions, 
  EventsStoreGetters,
  CacheOptions 
} from './interfaces';
import {
  createBaseState,
  createBaseGetters,
  createCacheActions,
  createStateActions,
  createDomainStateSetters,
  createDomainGetters,
  createGroupingGetter,
  withCachedApiAction,
  withApiAction,
  handleEmptySearch,
  CACHE_TIMEOUT
} from './base';

export const useEventsStore = defineStore('events', {
  state: (): EventsStoreState => ({
    events: [],
    event: null,
    categories: [],
    featuredEvents: [],
    searchResults: [],
    ...createBaseState(),
  }),

  getters: {
    ...createBaseGetters(),
    ...createDomainGetters<Event>('events'),

    eventsByType: createGroupingGetter<Event>('events', 'event_type'),

    upcomingEvents(): Event[] {
      const today = new Date().toISOString().split('T')[0]; // YYYY-MM-DD format
      return this.events
        .filter(event => event.event_date >= today)
        .sort((a, b) => new Date(a.event_date).getTime() - new Date(b.event_date).getTime());
    },

    pastEvents(): Event[] {
      const today = new Date().toISOString().split('T')[0]; // YYYY-MM-DD format
      return this.events
        .filter(event => event.event_date < today)
        .sort((a, b) => new Date(b.event_date).getTime() - new Date(a.event_date).getTime());
    },
  } satisfies EventsStoreGetters,

  actions: {
    // Base functionality
    ...createStateActions(),
    ...createCacheActions('events'),
    ...createDomainStateSetters<Event, EventCategory>('events', 'categories', 'featuredEvents'),

    // Domain-specific state setters

    // API Actions
    async fetchEvents(params?: GetEventsRequest, options: CacheOptions = {}): Promise<void> {
      await withCachedApiAction(
        this,
        params,
        options,
        () => apiClient.getEvents({
          page: params?.page || 1,
          limit: params?.limit || 20,
          search: params?.search,
          categoryId: params?.categoryId
        }),
        (response) => this.setEvents(
          response.data || [], 
          response.pagination?.total_items || 0, 
          params?.page || 1, 
          params?.limit || 20
        ),
        (items, count) => this.setEvents(items, count, 1, 20),
        'Failed to fetch events via contract client'
      );
    },

    async fetchEvent(slug: string): Promise<Event | null> {
      const result = await withApiAction(
        this,
        () => apiClient.getEventById(slug), // Using ID for now - slug lookup would need API extension
        'Failed to fetch event via contract client'
      );
      this.event = result?.data || null;
      return this.event;
    },

    async fetchFeaturedEvents(limit?: number): Promise<void> {
      const result = await withApiAction(
        this,
        () => apiClient.getFeaturedEvents(),
        'Failed to fetch featured events via contract client'
      );
      this.setFeaturedEvents(result?.data?.slice(0, limit) || []);
    },

    async searchEvents(params: { q: string, page?: number, limit?: number }): Promise<void> {
      // Handle empty search queries
      if (handleEmptySearch(params.q, this.setSearchResults)) {
        return;
      }

      const result = await withApiAction(
        this,
        () => apiClient.getEvents({
          page: params.page || 1,
          limit: params.limit || 20,
          search: params.q
        }),
        'Failed to search events via contract client'
      );
      this.setSearchResults(result?.data || [], result?.pagination?.total_items || 0);
    },

    async fetchEventCategories(): Promise<void> {
      const result = await withApiAction(
        this,
        () => apiClient.getEventCategories(),
        'Failed to fetch event categories via contract client'
      );
      this.setCategories(result?.data || []);
    },
  } satisfies EventsStoreActions,
});