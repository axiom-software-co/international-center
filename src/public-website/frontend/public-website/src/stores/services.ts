import { defineStore } from 'pinia';
import { servicesClient } from '../lib/clients';
import type { 
  Service, 
  ServiceCategory, 
  GetServicesParams, 
  SearchServicesParams 
} from '../lib/clients/services/types';
import type { 
  ServicesStoreState, 
  ServicesStoreActions, 
  ServicesStoreGetters,
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

export const useServicesStore = defineStore('services', {
  state: (): ServicesStoreState => ({
    services: [],
    service: null,
    categories: [],
    featuredServices: [],
    searchResults: [],
    ...createBaseState(),
  }),

  getters: {
    ...createBaseGetters(),
    ...createDomainGetters<Service>('services'),

    servicesByDeliveryMode: createGroupingGetter<Service>('services', 'delivery_mode'),
  } satisfies ServicesStoreGetters,

  actions: {
    // Base functionality
    ...createStateActions(),
    ...createCacheActions('services'),

    ...createDomainStateSetters<Service, ServiceCategory>('services', 'categories', 'featuredServices'),

    // Domain-specific state setters

    // API Actions
    async fetchServices(params?: GetServicesParams, options: CacheOptions = {}): Promise<void> {
      await withCachedApiAction(
        this,
        params,
        options,
        () => servicesClient.getServices(params),
        (response) => this.setServices(
          response.services, 
          response.count, 
          params?.page || 1, 
          params?.pageSize || 10
        ),
        (items, count) => this.setServices(items, count, 1, 10),
        'Failed to fetch services'
      );
    },

    async fetchService(slug: string): Promise<Service | null> {
      const result = await withApiAction(
        this,
        () => servicesClient.getServiceBySlug(slug),
        'Failed to fetch service'
      );
      this.service = result?.service || null;
      return this.service;
    },

    async fetchFeaturedServices(limit?: number): Promise<void> {
      const result = await withApiAction(
        this,
        () => servicesClient.getFeaturedServices(limit),
        'Failed to fetch featured services'
      );
      this.setFeaturedServices(result?.services || []);
    },

    async searchServices(params: SearchServicesParams): Promise<void> {
      // Handle empty search queries
      if (handleEmptySearch(params.q, this.setSearchResults)) {
        return;
      }

      const result = await withApiAction(
        this,
        () => servicesClient.searchServices(params),
        'Failed to search services'
      );
      this.setSearchResults(result?.services || [], result?.count || 0);
    },

    async fetchServiceCategories(): Promise<void> {
      const result = await withApiAction(
        this,
        () => servicesClient.getServiceCategories(),
        'Failed to fetch service categories'
      );
      this.setCategories(result?.categories || []);
    },
  } satisfies ServicesStoreActions,
});