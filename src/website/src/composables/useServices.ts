// Services Composables - Vue 3 Composition API composables for services data
// Provides clean interface for Vue components to interact with services domain

import { ref, computed, onMounted, watch, type Ref } from 'vue';
import { servicesClient } from '../lib/clients';
import type { Service, GetServicesParams, ServiceCategory } from '../lib/clients';

export interface UseServicesResult {
  services: Ref<Service[]>;
  loading: Ref<boolean>;
  error: Ref<string | null>;
  total: Ref<number>;
  page: Ref<number>;
  pageSize: Ref<number>;
  totalPages: Ref<number>;
  refetch: () => Promise<void>;
}

export interface UseServicesOptions extends GetServicesParams {
  enabled?: boolean;
  immediate?: boolean;
}

export function useServices(options: UseServicesOptions = {}): UseServicesResult {
  const { enabled = true, immediate = true, ...params } = options;

  const services = ref<Service[]>([]);
  const loading = ref(enabled && immediate);
  const error = ref<string | null>(null);
  const total = ref(0);
  const page = ref(params.page || 1);
  const pageSize = ref(params.pageSize || 10);
  const totalPages = ref(0);

  const fetchServices = async () => {
    if (!enabled) return;

    try {
      loading.value = true;
      error.value = null;

      // Backend returns: { services: [...], count: number, correlation_id: string }
      const response = await servicesClient.getServices(params);

      if (response.services) {
        services.value = response.services;
        total.value = response.count || response.services.length;
        // Calculate pagination from current params since backend doesn't return pagination info
        page.value = params.page || 1;
        pageSize.value = params.pageSize || 10;
        totalPages.value = Math.ceil(total.value / pageSize.value);
      } else {
        throw new Error('Failed to fetch services');
      }
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : 'Failed to fetch services';
      error.value = errorMessage;
      console.error('Error fetching services:', err);
      services.value = [];
      total.value = 0;
      totalPages.value = 0;
    } finally {
      loading.value = false;
    }
  };

  // Watch for parameter changes
  watch(() => params, fetchServices, { deep: true });
  
  if (immediate) {
    onMounted(fetchServices);
  }

  return {
    services,
    loading,
    error,
    total,
    page,
    pageSize,
    totalPages,
    refetch: fetchServices,
  };
}

export interface UseServiceResult {
  service: Ref<Service | null>;
  loading: Ref<boolean>;
  error: Ref<string | null>;
  refetch: () => Promise<void>;
}

export function useService(slug: Ref<string | null> | string | null): UseServiceResult {
  const slugRef = typeof slug === 'string' ? ref(slug) : slug || ref(null);
  
  const service = ref<Service | null>(null);
  const loading = ref(!!slugRef.value);
  const error = ref<string | null>(null);

  const fetchService = async () => {
    if (!slugRef.value) {
      service.value = null;
      loading.value = false;
      return;
    }

    try {
      loading.value = true;
      error.value = null;

      // Backend returns: { service: {...}, correlation_id: string }
      const response = await servicesClient.getServiceBySlug(slugRef.value);
      
      if (response.service) {
        service.value = response.service;
      } else {
        throw new Error('Failed to fetch service');
      }
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : 'Failed to fetch service';
      error.value = errorMessage;
      console.error('Error fetching service:', err);
      service.value = null;
    } finally {
      loading.value = false;
    }
  };

  // Watch for slug changes
  watch(slugRef, fetchService, { immediate: true });

  return {
    service,
    loading,
    error,
    refetch: fetchService,
  };
}

export interface UseFeaturedServicesResult {
  services: Ref<Service[]>;
  loading: Ref<boolean>;
  error: Ref<string | null>;
  refetch: () => Promise<void>;
}

export function useFeaturedServices(limit?: Ref<number> | number): UseFeaturedServicesResult {
  const limitRef = typeof limit === 'number' ? ref(limit) : limit;
  
  const services = ref<Service[]>([]);
  const loading = ref(true);
  const error = ref<string | null>(null);

  const fetchFeaturedServices = async () => {
    try {
      loading.value = true;
      error.value = null;

      // Backend returns: { services: [...], count: number, correlation_id: string }
      const response = await servicesClient.getFeaturedServices(limitRef?.value);
      
      if (response.services) {
        services.value = response.services;
      } else {
        throw new Error('Failed to fetch featured services');
      }
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : 'Failed to fetch featured services';
      error.value = errorMessage;
      console.error('Error fetching featured services:', err);
      services.value = [];
    } finally {
      loading.value = false;
    }
  };

  // Watch for limit changes
  if (limitRef) {
    watch(limitRef, fetchFeaturedServices, { immediate: true });
  } else {
    onMounted(fetchFeaturedServices);
  }

  return {
    services,
    loading,
    error,
    refetch: fetchFeaturedServices,
  };
}

export interface UseServiceCategoriesResult {
  categories: Ref<ServiceCategory[]>;
  loading: Ref<boolean>;
  error: Ref<string | null>;
  refetch: () => Promise<void>;
}

export function useServiceCategories(): UseServiceCategoriesResult {
  const categories = ref<ServiceCategory[]>([]);
  const loading = ref(true);
  const error = ref<string | null>(null);

  const fetchServiceCategories = async () => {
    try {
      loading.value = true;
      error.value = null;

      // Backend returns: { categories: [...], count: number, correlation_id: string }
      const response = await servicesClient.getServiceCategories();
      
      if (response.categories) {
        categories.value = response.categories;
      } else {
        throw new Error('Failed to fetch service categories');
      }
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : 'Failed to fetch service categories';
      error.value = errorMessage;
      console.error('Error fetching service categories:', err);
      categories.value = [];
    } finally {
      loading.value = false;
    }
  };

  onMounted(fetchServiceCategories);

  return {
    categories,
    loading,
    error,
    refetch: fetchServiceCategories,
  };
}

export interface UseSearchServicesResult {
  results: Ref<Service[]>;
  loading: Ref<boolean>;
  error: Ref<string | null>;
  total: Ref<number>;
  page: Ref<number>;
  pageSize: Ref<number>;
  totalPages: Ref<number>;
  search: (query: string, options?: Partial<GetServicesParams>) => Promise<void>;
}

export function useSearchServices(): UseSearchServicesResult {
  const results = ref<Service[]>([]);
  const loading = ref(false);
  const error = ref<string | null>(null);
  const total = ref(0);
  const page = ref(1);
  const pageSize = ref(10);
  const totalPages = ref(0);

  const search = async (query: string, options: Partial<GetServicesParams> = {}) => {
    if (!query.trim()) {
      results.value = [];
      total.value = 0;
      totalPages.value = 0;
      return;
    }

    try {
      loading.value = true;
      error.value = null;

      // Backend returns: { services: [...], count: number, correlation_id: string }
      const response = await servicesClient.searchServices({
        q: query,
        page: options.page || 1,
        pageSize: options.pageSize || 10,
        category: options.category,
        ...options,
      });

      if (response.services) {
        results.value = response.services;
        total.value = response.count || response.services.length;
        page.value = options.page || 1;
        pageSize.value = options.pageSize || 10;
        totalPages.value = Math.ceil(total.value / pageSize.value);
      } else {
        throw new Error('Failed to search services');
      }
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : 'Failed to search services';
      error.value = errorMessage;
      console.error('Error searching services:', err);
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