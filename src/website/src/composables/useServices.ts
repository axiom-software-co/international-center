// Services Composables - Vue 3 Composition API with Direct Client Integration

import { type Ref, ref, isRef, unref, watch, onMounted, computed, nextTick } from 'vue';
import { servicesClient } from '../lib/clients';
import type { Service, ServiceCategory, GetServicesParams, SearchServicesParams } from '../lib/clients/services/types';

// Domain-specific type aliases
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

// Explicit main list composable
export function useServices(options: UseServicesOptions = {}): UseServicesResult {
  const { enabled = true, immediate = true, ...params } = options;
  
  const services = ref<Service[]>([]);
  const loading = ref(false);
  const error = ref<string | null>(null);
  const total = ref(0);
  const page = ref(params.page || 1);
  const pageSize = ref(params.pageSize || 10);
  
  const totalPages = computed(() => Math.ceil(total.value / pageSize.value) || 0);

  const fetchServices = async () => {
    if (!enabled) return;
    
    loading.value = true;
    error.value = null;
    
    try {
      const response = await servicesClient.getServices(params);
      services.value = response.services;
      total.value = response.count;
    } catch (err) {
      error.value = err instanceof Error ? err.message : 'Failed to fetch services';
      services.value = [];
      total.value = 0;
    } finally {
      loading.value = false;
    }
  };

  // Watch for parameter changes
  watch(() => params, fetchServices, { deep: true });
  
  // Call immediately if enabled and immediate is true
  if (enabled && immediate) {
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

// Explicit single item composable
export function useService(slug: Ref<string | null> | string | null): UseServiceResult {
  const slugRef = isRef(slug) ? slug : ref(slug);
  const service = ref<Service | null>(null);
  const loading = ref(false);
  const error = ref<string | null>(null);

  const fetchService = async () => {
    const currentSlug = unref(slugRef);
    if (!currentSlug) {
      service.value = null;
      return;
    }

    loading.value = true;
    error.value = null;
    
    try {
      const response = await servicesClient.getServiceBySlug(currentSlug);
      service.value = response.service;
    } catch (err) {
      error.value = err instanceof Error ? err.message : 'Failed to fetch service';
      service.value = null;
    } finally {
      loading.value = false;
    }
  };

  // Watch for slug changes and call immediately
  watch(slugRef, (newSlug) => {
    if (newSlug) {
      fetchService();
    } else {
      service.value = null;
    }
  }, { immediate: true });

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

// Explicit featured services composable
export function useFeaturedServices(limit?: Ref<number | undefined> | number | undefined): UseFeaturedServicesResult {
  const limitRef = isRef(limit) ? limit : ref(limit);
  const services = ref<Service[]>([]);
  const loading = ref(false);
  const error = ref<string | null>(null);

  const fetchFeaturedServices = async () => {
    loading.value = true;
    error.value = null;
    
    try {
      const response = await servicesClient.getFeaturedServices(unref(limitRef));
      services.value = response.services;
    } catch (err) {
      error.value = err instanceof Error ? err.message : 'Failed to fetch featured services';
      services.value = [];
    } finally {
      loading.value = false;
    }
  };

  // Watch for limit changes and fetch immediately
  watch(limitRef, fetchFeaturedServices, { immediate: true });

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

// Explicit categories composable
export function useServiceCategories(): UseServiceCategoriesResult {
  const categories = ref<ServiceCategory[]>([]);
  const loading = ref(false);
  const error = ref<string | null>(null);

  const fetchCategories = async () => {
    loading.value = true;
    error.value = null;
    
    try {
      const response = await servicesClient.getServiceCategories();
      categories.value = response.categories;
    } catch (err) {
      error.value = err instanceof Error ? err.message : 'Failed to fetch service categories';
      categories.value = [];
    } finally {
      loading.value = false;
    }
  };

  // Trigger initial fetch immediately
  onMounted(() => {
    fetchCategories();
  });

  return {
    categories,
    loading,
    error,
    refetch: fetchCategories,
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
  search: (query: string, options?: Partial<SearchServicesParams>) => Promise<void>;
}

// Explicit search composable
export function useSearchServices(): UseSearchServicesResult {
  const results = ref<Service[]>([]);
  const loading = ref(false);
  const error = ref<string | null>(null);
  const total = ref(0);
  const page = ref(1);
  const pageSize = ref(10);

  const totalPages = computed(() => Math.ceil(total.value / pageSize.value) || 0);

  const search = async (query: string, options: Partial<SearchServicesParams> = {}) => {
    const searchParams = {
      q: query,
      page: options.page || 1,
      pageSize: options.pageSize || 10,
      ...options,
    };

    // Update local pagination refs
    page.value = searchParams.page;
    pageSize.value = searchParams.pageSize;

    loading.value = true;
    error.value = null;
    
    try {
      const response = await servicesClient.searchServices(searchParams);
      results.value = response.services;
      total.value = response.count;
    } catch (err) {
      error.value = err instanceof Error ? err.message : 'Failed to search services';
      results.value = [];
      total.value = 0;
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