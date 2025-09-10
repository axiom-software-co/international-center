import { defineStore } from 'pinia';
import { apiClient } from '../lib/api-client';
import type { 
  Service, 
  ServiceCategory,
  GetServicesRequest
} from '@international-center/public-api-client';
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
    async fetchServices(params?: GetServicesRequest, options: CacheOptions = {}): Promise<void> {
      await withCachedApiAction(
        this,
        params,
        options,
        () => apiClient.getServices({
          page: params?.page || 1,
          limit: params?.limit || 20,
          search: params?.search,
          categoryId: params?.categoryId
        }),
        (response) => this.setServices(
          response.data || [], 
          response.pagination?.total_items || 0, 
          params?.page || 1, 
          params?.limit || 20
        ),
        (items, count) => this.setServices(items, count, 1, 20),
        'Failed to fetch services via contract client'
      );
    },

    async fetchService(slug: string): Promise<Service | null> {
      const result = await withApiAction(
        this,
        () => apiClient.getServiceById(slug), // Using ID for now - slug lookup would need API extension
        'Failed to fetch service via contract client'
      );
      this.service = result?.data || null;
      return this.service;
    },

    async fetchFeaturedServices(limit?: number): Promise<void> {
      const result = await withApiAction(
        this,
        () => apiClient.getFeaturedServices(),
        'Failed to fetch featured services via contract client'
      );
      this.setFeaturedServices(result?.data?.slice(0, limit) || []);
    },

    async searchServices(params: { q: string, page?: number, limit?: number }): Promise<void> {
      // Handle empty search queries
      if (handleEmptySearch(params.q, this.setSearchResults)) {
        return;
      }

      const result = await withApiAction(
        this,
        () => apiClient.getServices({
          page: params.page || 1,
          limit: params.limit || 20,
          search: params.q
        }),
        'Failed to search services via contract client'
      );
      this.setSearchResults(result?.data || [], result?.pagination?.total_items || 0);
    },

    async fetchServiceCategories(): Promise<void> {
      const result = await withApiAction(
        this,
        () => apiClient.getServiceCategories(),
        'Failed to fetch service categories via contract client'
      );
      this.setCategories(result?.data || []);
    },
  } satisfies ServicesStoreActions,
});