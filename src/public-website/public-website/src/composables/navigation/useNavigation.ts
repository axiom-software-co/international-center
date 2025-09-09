// Navigation Composable - Centralized navigation state and utilities
// Provides consistent navigation functionality across the application

import { computed, onMounted } from 'vue';
import { useNavigationStore } from '../../stores/navigation';
import type { NavigationServiceCategory, FooterServiceCategory } from '../../lib/navigation-data';
import { generateItemUrl, generateCategoryUrl, generateSearchUrl } from '../../lib/navigation/domains';

export interface UseNavigationResult {
  // Service navigation data
  serviceCategories: NavigationServiceCategory[];
  footerCategories: FooterServiceCategory[];
  
  // Domain categories for cross-domain navigation
  eventsCategories: any[];
  newsCategories: any[];
  researchCategories: any[];
  
  // Loading states
  isLoading: boolean;
  
  // Utilities
  refreshNavigation: (forceRefresh?: boolean) => Promise<void>;
  generateItemUrl: (domain: string, slug: string) => string;
  generateCategoryUrl: (domain: string, categorySlug: string) => string;
  generateSearchUrl: (domain: string, query?: string) => string;
  getCategoriesForDomain: (domain: string) => any[];
}

/**
 * Main navigation composable - provides all navigation data and utilities
 */
export function useNavigation(): UseNavigationResult {
  const navigationStore = useNavigationStore();

  // Computed properties for reactive state
  const serviceCategories = computed(() => navigationStore.state.serviceCategories);
  const footerCategories = computed(() => navigationStore.state.footerCategories);
  const eventsCategories = computed(() => navigationStore.state.eventsCategories);
  const newsCategories = computed(() => navigationStore.state.newsCategories);
  const researchCategories = computed(() => navigationStore.state.researchCategories);

  const isLoading = computed(() => {
    return navigationStore.state.servicesLoading || navigationStore.state.categoriesLoading;
  });

  // Actions
  const refreshNavigation = async (forceRefresh = false) => {
    await navigationStore.loadAllNavigationData(forceRefresh);
  };

  // Initialize navigation data on mount
  onMounted(() => {
    // Only load if data is stale
    if (!navigationStore.isServicesDataFresh) {
      refreshNavigation();
    }
  });

  return {
    // Data
    serviceCategories: serviceCategories.value,
    footerCategories: footerCategories.value,
    eventsCategories: eventsCategories.value,
    newsCategories: newsCategories.value,
    researchCategories: researchCategories.value,
    
    // State
    isLoading: isLoading.value,
    
    // Actions
    refreshNavigation,
    
    // Utilities
    generateItemUrl,
    generateCategoryUrl,
    generateSearchUrl,
    getCategoriesForDomain: navigationStore.getCategoriesForDomain,
  };
}

/**
 * Services-specific navigation composable
 */
export function useServicesNavigation() {
  const navigationStore = useNavigationStore();

  const serviceCategories = computed(() => navigationStore.state.serviceCategories);
  const footerCategories = computed(() => navigationStore.state.footerCategories);
  const isLoading = computed(() => navigationStore.state.servicesLoading);

  const refreshServices = async (forceRefresh = false) => {
    await navigationStore.loadServicesNavigation(forceRefresh);
  };

  onMounted(() => {
    if (!navigationStore.isServicesDataFresh) {
      refreshServices();
    }
  });

  return {
    serviceCategories: serviceCategories.value,
    footerCategories: footerCategories.value,
    isLoading: isLoading.value,
    refreshServices,
  };
}

/**
 * Domain categories navigation composable
 */
export function useDomainCategories(domain: string) {
  const navigationStore = useNavigationStore();

  const categories = computed(() => navigationStore.getCategoriesForDomain(domain));
  const isLoading = computed(() => navigationStore.state.categoriesLoading);

  const refreshCategories = async (forceRefresh = false) => {
    await navigationStore.loadDomainCategories(domain, forceRefresh);
  };

  onMounted(() => {
    const hasCategories = categories.value.length > 0;
    if (!hasCategories) {
      refreshCategories();
    }
  });

  return {
    categories: categories.value,
    isLoading: isLoading.value,
    refreshCategories,
  };
}