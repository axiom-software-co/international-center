import { defineStore } from 'pinia';
import { eventsClient } from '../lib/clients';
import type { 
  Event, 
  EventCategory, 
  GetEventsParams, 
  SearchEventsParams 
} from '../lib/clients/events/types';
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
    async fetchEvents(params?: GetEventsParams, options: CacheOptions = {}): Promise<void> {
      await withCachedApiAction(
        this,
        params,
        options,
        () => eventsClient.getEvents(params),
        (response) => this.setEvents(
          response.events, 
          response.count, 
          params?.page || 1, 
          params?.pageSize || 10
        ),
        (items, count) => this.setEvents(items, count, 1, 10),
        'Failed to fetch events'
      );
    },

    async fetchEvent(slug: string): Promise<Event | null> {
      const result = await withApiAction(
        this,
        () => eventsClient.getEventBySlug(slug),
        'Failed to fetch event'
      );
      this.event = result?.event || null;
      return this.event;
    },

    async fetchFeaturedEvents(limit?: number): Promise<void> {
      const result = await withApiAction(
        this,
        () => eventsClient.getFeaturedEvents(limit),
        'Failed to fetch featured events'
      );
      this.setFeaturedEvents(result?.events || []);
    },

    async searchEvents(params: SearchEventsParams): Promise<void> {
      // Handle empty search queries
      if (handleEmptySearch(params.q, this.setSearchResults)) {
        return;
      }

      const result = await withApiAction(
        this,
        () => eventsClient.searchEvents(params),
        'Failed to search events'
      );
      this.setSearchResults(result?.events || [], result?.count || 0);
    },

    async fetchEventCategories(): Promise<void> {
      const result = await withApiAction(
        this,
        () => eventsClient.getEventCategories(),
        'Failed to fetch event categories'
      );
      this.setCategories(result?.categories || []);
    },
  } satisfies EventsStoreActions,
});