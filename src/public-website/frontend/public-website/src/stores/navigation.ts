// Navigation Store - Centralized navigation state management
// Caches navigation data and provides cross-domain navigation utilities

import { defineStore } from 'pinia';
import { ref, computed } from 'vue';
import { servicesClient, newsClient, researchClient, eventsClient } from '../lib/clients';
import type { NavigationServiceCategory, FooterServiceCategory } from '../lib/navigation-data';
import { transformToNavigationCategories, transformToFooterCategories } from '../lib/navigation-data';
import { DOMAIN_CONFIGS, type DomainConfig } from '../lib/navigation/domains';

interface NavigationCache {
  data: any;
  timestamp: number;
  expiry: number;
}

interface NavigationState {
  // Services navigation data
  serviceCategories: NavigationServiceCategory[];
  footerCategories: FooterServiceCategory[];
  
  // Domain categories for cross-domain filtering
  eventsCategories: any[];
  newsCategories: any[];
  researchCategories: any[];
  
  // Loading states
  servicesLoading: boolean;
  categoriesLoading: boolean;
  
  // Cache management
  lastUpdated: number;
  cacheExpiry: number; // 15 minutes
}

export const useNavigationStore = defineStore('navigation', () => {
  // State
  const state = ref<NavigationState>({
    serviceCategories: [],
    footerCategories: [],
    eventsCategories: [],
    newsCategories: [],
    researchCategories: [],
    servicesLoading: false,
    categoriesLoading: false,
    lastUpdated: 0,
    cacheExpiry: 15 * 60 * 1000, // 15 minutes
  });

  // Cache for API responses
  const cache = new Map<string, NavigationCache>();

  // Getters
  const isServicesDataFresh = computed(() => {
    const now = Date.now();
    return (now - state.value.lastUpdated) < state.value.cacheExpiry;
  });

  const availableDomains = computed(() => {
    return Object.keys(DOMAIN_CONFIGS).map(domain => ({
      domain,
      config: DOMAIN_CONFIGS[domain],
      hasCategories: getCategoriesForDomain(domain).length > 0,
    }));
  });

  // Cache utilities
  function setCache(key: string, data: any, ttl: number = 15 * 60 * 1000) {
    cache.set(key, {
      data,
      timestamp: Date.now(),
      expiry: ttl,
    });
  }

  function getCache(key: string): any | null {
    const cached = cache.get(key);
    if (!cached) return null;

    const now = Date.now();
    if (now > cached.timestamp + cached.expiry) {
      cache.delete(key);
      return null;
    }

    return cached.data;
  }

  function clearCache(pattern?: string) {
    if (pattern) {
      Array.from(cache.keys())
        .filter(key => key.includes(pattern))
        .forEach(key => cache.delete(key));
    } else {
      cache.clear();
    }
  }

  // Domain-specific category getters
  function getCategoriesForDomain(domain: string) {
    switch (domain) {
      case 'events':
        return state.value.eventsCategories;
      case 'news':
        return state.value.newsCategories;
      case 'research':
        return state.value.researchCategories;
      case 'services':
        return state.value.serviceCategories;
      default:
        return [];
    }
  }

  // Actions
  async function loadServicesNavigation(forceRefresh = false): Promise<void> {
    const cacheKey = 'services-navigation';
    
    if (!forceRefresh && isServicesDataFresh.value) {
      const cached = getCache(cacheKey);
      if (cached) {
        state.value.serviceCategories = cached.serviceCategories;
        state.value.footerCategories = cached.footerCategories;
        return;
      }
    }

    try {
      state.value.servicesLoading = true;

      const [servicesResponse, categoriesResponse] = await Promise.all([
        servicesClient.getServices({ pageSize: 100 }),
        servicesClient.getServiceCategories(),
      ]);

      const services = servicesResponse.services || servicesResponse.data || [];
      const categories = categoriesResponse.categories || categoriesResponse.data || [];

      const navigationCategories = transformToNavigationCategories(services, categories);
      const footerCategories = transformToFooterCategories(services, categories);

      // Update state
      state.value.serviceCategories = navigationCategories;
      state.value.footerCategories = footerCategories;
      state.value.lastUpdated = Date.now();

      // Cache results
      setCache(cacheKey, {
        serviceCategories: navigationCategories,
        footerCategories: footerCategories,
      });

    } catch (error) {
      console.error('Failed to load services navigation:', error);
      // Keep existing data on error
    } finally {
      state.value.servicesLoading = false;
    }
  }

  async function loadDomainCategories(domain: string, forceRefresh = false): Promise<void> {
    const cacheKey = `${domain}-categories`;
    
    if (!forceRefresh) {
      const cached = getCache(cacheKey);
      if (cached) {
        switch (domain) {
          case 'events':
            state.value.eventsCategories = cached;
            break;
          case 'news':
            state.value.newsCategories = cached;
            break;
          case 'research':
            state.value.researchCategories = cached;
            break;
        }
        return;
      }
    }

    try {
      state.value.categoriesLoading = true;
      let categories: any[] = [];

      switch (domain) {
        case 'events':
          const eventsResponse = await eventsClient.getEventCategories();
          categories = eventsResponse.categories || eventsResponse.data || [];
          state.value.eventsCategories = categories;
          break;

        case 'news':
          const newsResponse = await newsClient.getNewsCategories();
          categories = newsResponse.categories || newsResponse.data || [];
          state.value.newsCategories = categories;
          break;

        case 'research':
          const researchResponse = await researchClient.getResearchCategories();
          categories = researchResponse.categories || researchResponse.data || [];
          state.value.researchCategories = categories;
          break;

        default:
          console.warn(`Unknown domain: ${domain}`);
          return;
      }

      // Cache results
      setCache(cacheKey, categories);

    } catch (error) {
      console.error(`Failed to load ${domain} categories:`, error);
    } finally {
      state.value.categoriesLoading = false;
    }
  }

  async function loadAllNavigationData(forceRefresh = false): Promise<void> {
    await Promise.all([
      loadServicesNavigation(forceRefresh),
      loadDomainCategories('events', forceRefresh),
      loadDomainCategories('news', forceRefresh),
      loadDomainCategories('research', forceRefresh),
    ]);
  }

  // Domain utilities
  function getDomainConfig(domain: string): DomainConfig | null {
    return DOMAIN_CONFIGS[domain] || null;
  }

  function generateBreadcrumbProps(domain: string, itemName: string, title: string, category?: string) {
    const config = getDomainConfig(domain);
    if (!config) {
      throw new Error(`Unknown domain: ${domain}`);
    }

    return {
      parentPath: config.parentPath,
      parentLabel: config.parentLabel,
      itemName,
      title,
      category,
    };
  }

  function generateItemUrl(domain: string, slug: string): string {
    const config = getDomainConfig(domain);
    if (!config) {
      throw new Error(`Unknown domain: ${domain}`);
    }
    return `${config.baseUrl}/${slug}`;
  }

  // Initialize on store creation
  loadAllNavigationData();

  return {
    // State
    state: state.value,
    
    // Getters
    isServicesDataFresh,
    availableDomains,
    getCategoriesForDomain,
    
    // Actions
    loadServicesNavigation,
    loadDomainCategories,
    loadAllNavigationData,
    clearCache,
    
    // Utilities
    getDomainConfig,
    generateBreadcrumbProps,
    generateItemUrl,
  };
});