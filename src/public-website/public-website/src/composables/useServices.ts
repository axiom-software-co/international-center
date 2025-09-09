// Services Composables - Vue 3 Composition API with Store Delegation

import { type Ref, ref, isRef, unref, watch, onMounted, computed } from 'vue';
import { storeToRefs } from 'pinia';
import { useServicesStore } from '../stores/services';
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

// Main list composable - delegates to store
export function useServices(options: UseServicesOptions = {}): UseServicesResult {
  const { enabled = true, immediate = true, ...params } = options;
  const store = useServicesStore();
  const { services, loading, error, total, totalPages } = storeToRefs(store);
  
  // Local refs for pagination
  const page = ref(params.page || 1);
  const pageSize = ref(params.pageSize || 10);

  const fetchServices = async () => {
    if (!enabled) return;
    await store.fetchServices(params);
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

// Single item composable - delegates to store
export function useService(slug: Ref<string | null> | string | null): UseServiceResult {
  const slugRef = isRef(slug) ? slug : ref(slug);
  const store = useServicesStore();
  const { service, loading, error } = storeToRefs(store);

  const fetchService = async () => {
    const currentSlug = unref(slugRef);
    if (!currentSlug) {
      store.service = null;
      return;
    }

    await store.fetchService(currentSlug);
  };

  // Watch for slug changes and call immediately
  watch(slugRef, (newSlug) => {
    if (newSlug) {
      fetchService();
    } else {
      store.service = null;
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

// Featured services composable - delegates to store
export function useFeaturedServices(limit?: Ref<number | undefined> | number | undefined): UseFeaturedServicesResult {
  const limitRef = isRef(limit) ? limit : ref(limit);
  const store = useServicesStore();
  const { featuredServices: services, loading, error } = storeToRefs(store);

  const fetchFeaturedServices = async () => {
    await store.fetchFeaturedServices(unref(limitRef));
  };

  // Trigger initial fetch immediately
  fetchFeaturedServices();
  
  // Watch for limit changes
  watch(limitRef, fetchFeaturedServices);

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

// Categories composable - delegates to store
export function useServiceCategories(): UseServiceCategoriesResult {
  const store = useServicesStore();
  const { categories, loading, error } = storeToRefs(store);

  const fetchCategories = async () => {
    await store.fetchServiceCategories();
  };

  // Trigger initial fetch immediately
  fetchCategories();

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

// Search composable - delegates to store
export function useSearchServices(): UseSearchServicesResult {
  const store = useServicesStore();
  const { searchResults: results, loading, error, searchTotal: total } = storeToRefs(store);
  
  // Local refs for search-specific pagination
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

    await store.searchServices(searchParams);
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