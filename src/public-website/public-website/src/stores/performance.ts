// Performance-Enhanced Store Utilities
// Integrates caching, error handling, and performance monitoring into stores

import { ref, computed } from 'vue';
import type { Ref } from 'vue';
import { useSmartCache, withSmartCache, type UseSmartCacheResult, type CacheOptions } from '../composables/performance/useSmartCache';
import { useErrorHandler, type UseErrorHandlerResult, type ErrorContext } from '../composables/performance/useErrorHandler';
import { usePerformanceMonitor, type UsePerformanceMonitorResult, withPerformanceTracking } from '../composables/performance/usePerformanceMonitor';

export interface PerformanceStoreOptions {
  enableCaching?: boolean;
  enableErrorHandling?: boolean;  
  enablePerformanceMonitoring?: boolean;
  cacheOptions?: {
    defaultTtl?: number;
    maxSize?: number;
    enablePersistence?: boolean;
  };
}

export interface PerformanceStoreResult {
  // Performance utilities
  cache?: UseSmartCacheResult;
  errorHandler?: UseErrorHandlerResult;
  performanceMonitor?: UsePerformanceMonitorResult;
  
  // Enhanced loading state
  isLoading: Ref<boolean>;
  loadingStates: Ref<Record<string, boolean>>;
  setLoading: (key: string, loading: boolean) => void;
  
  // Performance metrics
  metrics: {
    cacheHitRate: Ref<number>;
    averageResponseTime: Ref<number>;
    errorRate: Ref<number>;
  };
  
  // Enhanced API action wrapper
  performAction: <T>(
    actionName: string,
    action: () => Promise<T>,
    options?: {
      cacheKey?: string;
      cacheOptions?: CacheOptions;
      errorContext?: ErrorContext;
      enableRetry?: boolean;
    }
  ) => Promise<T>;
}

export function usePerformanceStore(
  storeName: string,
  options: PerformanceStoreOptions = {}
): PerformanceStoreResult {
  
  const {
    enableCaching = true,
    enableErrorHandling = true,
    enablePerformanceMonitoring = true,
    cacheOptions = {},
  } = options;

  // Initialize performance utilities
  const cache = enableCaching ? useSmartCache({
    defaultTtl: cacheOptions.defaultTtl || 5 * 60 * 1000, // 5 minutes
    maxSize: cacheOptions.maxSize || 100,
    enablePersistence: cacheOptions.enablePersistence || false,
    storageKey: `${storeName}_cache`,
  }) : undefined;

  const performanceMonitor = enablePerformanceMonitoring ? usePerformanceMonitor({
    trackPageLoad: true,
    trackApiCalls: true,
    sampleRate: 0.1,
  }) : undefined;

  const errorHandler = enableErrorHandling ? useErrorHandler({
    enableLogging: true,
    enablePerformanceTracking: enablePerformanceMonitoring,
    performanceMonitor,
  }) : undefined;

  // Enhanced loading state management
  const isLoading = ref(false);
  const loadingStates = ref<Record<string, boolean>>({});

  const setLoading = (key: string, loading: boolean): void => {
    loadingStates.value[key] = loading;
    
    // Update global loading state
    isLoading.value = Object.values(loadingStates.value).some(state => state);
  };

  // Performance metrics
  const metrics = {
    cacheHitRate: computed(() => cache?.stats.hitRate || 0),
    averageResponseTime: computed(() => {
      if (!performanceMonitor) return 0;
      
      const summary = performanceMonitor.getMetricsSummary();
      const apiCallStats = summary['api_call'];
      return apiCallStats?.avg || 0;
    }),
    errorRate: computed(() => {
      if (!performanceMonitor) return 0;
      
      const summary = performanceMonitor.getMetricsSummary();
      const errorStats = summary['api_call_error'];
      const successStats = summary['api_call'];
      
      const totalErrors = errorStats?.count || 0;
      const totalSuccess = successStats?.count || 0;
      const total = totalErrors + totalSuccess;
      
      return total > 0 ? totalErrors / total : 0;
    }),
  };

  // Enhanced action performer with integrated caching, monitoring, and error handling
  const performAction = async <T>(
    actionName: string,
    action: () => Promise<T>,
    options: {
      cacheKey?: string;
      cacheOptions?: CacheOptions;
      errorContext?: ErrorContext;
      enableRetry?: boolean;
    } = {}
  ): Promise<T> => {
    
    const {
      cacheKey,
      cacheOptions: actionCacheOptions,
      errorContext,
      enableRetry = true,
    } = options;

    // Set loading state
    setLoading(actionName, true);

    try {
      // Create enhanced action with all performance features
      let enhancedAction = action;

      // Add performance tracking
      if (performanceMonitor) {
        enhancedAction = withPerformanceTracking(
          enhancedAction,
          `${storeName}_${actionName}`,
          performanceMonitor
        );
      }

      // Add caching
      if (cache && cacheKey) {
        enhancedAction = withSmartCache(
          enhancedAction,
          () => cacheKey,
          cache,
          actionCacheOptions
        );
      }

      // Add error handling with retry
      if (errorHandler && enableRetry) {
        enhancedAction = errorHandler.withErrorHandling(
          enhancedAction,
          {
            component: storeName,
            action: actionName,
            ...errorContext,
          },
          {
            maxAttempts: 3,
            baseDelay: 1000,
            shouldRetry: (error: Error, attempt: number) => {
              // Retry on network errors and 5xx server errors
              return attempt < 3 && (
                error.name.includes('Network') ||
                error.name.includes('Timeout') ||
                error.message.includes('5')
              );
            },
          }
        );
      }

      // Execute the enhanced action
      const result = await enhancedAction();

      return result;

    } catch (error) {
      // Handle error if no error handler is configured
      if (!errorHandler) {
        console.error(`Error in ${storeName}.${actionName}:`, error);
      }
      throw error;
    } finally {
      // Always clear loading state
      setLoading(actionName, false);
    }
  };

  return {
    cache,
    errorHandler,
    performanceMonitor,
    isLoading,
    loadingStates,
    setLoading,
    metrics,
    performAction,
  };
}

// Helper function to create a performance-enhanced store action
export function createPerformantAction<TParams, TResult>(
  performanceStore: PerformanceStoreResult,
  actionName: string,
  apiCall: (params: TParams) => Promise<TResult>,
  options: {
    generateCacheKey?: (params: TParams) => string;
    cacheOptions?: CacheOptions;
    errorContext?: ErrorContext;
  } = {}
) {
  const {
    generateCacheKey,
    cacheOptions,
    errorContext,
  } = options;

  return async (params: TParams): Promise<TResult> => {
    const cacheKey = generateCacheKey ? generateCacheKey(params) : undefined;
    
    return performanceStore.performAction(
      actionName,
      () => apiCall(params),
      {
        cacheKey,
        cacheOptions,
        errorContext,
      }
    );
  };
}

// Predefined cache key generators for common patterns
export const cacheKeyGenerators = {
  list: (params: Record<string, any>) => {
    const searchParams = new URLSearchParams();
    Object.entries(params).forEach(([key, value]) => {
      if (value !== undefined && value !== null) {
        searchParams.append(key, String(value));
      }
    });
    return `list_${searchParams.toString()}`;
  },
  
  single: (id: string | number) => `single_${id}`,
  
  search: (query: string, filters?: Record<string, any>) => {
    const searchParams = new URLSearchParams({ q: query });
    if (filters) {
      Object.entries(filters).forEach(([key, value]) => {
        if (value !== undefined && value !== null) {
          searchParams.append(key, String(value));
        }
      });
    }
    return `search_${searchParams.toString()}`;
  },
};

// Default cache options for different operation types
export const defaultCacheOptions: Record<string, CacheOptions> = {
  list: {
    ttl: 5 * 60 * 1000, // 5 minutes
    tags: ['list'],
  },
  
  single: {
    ttl: 10 * 60 * 1000, // 10 minutes
    tags: ['single'],
  },
  
  search: {
    ttl: 2 * 60 * 1000, // 2 minutes
    tags: ['search'],
  },
  
  featured: {
    ttl: 15 * 60 * 1000, // 15 minutes
    tags: ['featured'],
  },
};