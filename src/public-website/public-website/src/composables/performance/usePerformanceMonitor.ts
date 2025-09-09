// Performance Monitoring Composable
// Tracks page load times, API response times, and user interactions

import { ref, onMounted, onUnmounted, nextTick } from 'vue';

export interface PerformanceMetric {
  name: string;
  value: number;
  timestamp: number;
  context?: Record<string, any>;
}

export interface UsePerformanceMonitorOptions {
  trackPageLoad?: boolean;
  trackApiCalls?: boolean;
  trackUserInteractions?: boolean;
  sampleRate?: number; // 0-1, percentage of events to track
}

export interface UsePerformanceMonitorResult {
  metrics: PerformanceMetric[];
  startTimer: (name: string, context?: Record<string, any>) => () => void;
  recordMetric: (name: string, value: number, context?: Record<string, any>) => void;
  getMetricsSummary: () => Record<string, { avg: number; count: number; min: number; max: number }>;
  clearMetrics: () => void;
}

export function usePerformanceMonitor(
  options: UsePerformanceMonitorOptions = {}
): UsePerformanceMonitorResult {
  
  const {
    trackPageLoad = true,
    trackApiCalls = true,
    trackUserInteractions = false,
    sampleRate = 0.1, // Track 10% of events by default
  } = options;

  const metrics = ref<PerformanceMetric[]>([]);
  const activeTimers = new Map<string, number>();

  // Check if we should track this event based on sample rate
  const shouldTrack = (): boolean => {
    return Math.random() < sampleRate;
  };

  // Record a performance metric
  const recordMetric = (name: string, value: number, context?: Record<string, any>): void => {
    if (!shouldTrack()) return;

    const metric: PerformanceMetric = {
      name,
      value,
      timestamp: Date.now(),
      context,
    };

    metrics.value.push(metric);

    // Keep only the last 1000 metrics to prevent memory bloat
    if (metrics.value.length > 1000) {
      metrics.value.shift();
    }

    // Log to console in development
    if (import.meta.env?.DEV) {
      console.log(`[Performance] ${name}: ${value}ms`, context);
    }
  };

  // Start a timer and return a function to end it
  const startTimer = (name: string, context?: Record<string, any>) => {
    const startTime = performance.now();
    const timerId = `${name}_${startTime}_${Math.random()}`;
    activeTimers.set(timerId, startTime);

    return () => {
      const endTime = performance.now();
      const startTime = activeTimers.get(timerId);
      
      if (startTime !== undefined) {
        const duration = endTime - startTime;
        recordMetric(name, duration, context);
        activeTimers.delete(timerId);
      }
    };
  };

  // Get metrics summary
  const getMetricsSummary = (): Record<string, { avg: number; count: number; min: number; max: number }> => {
    const summary: Record<string, { avg: number; count: number; min: number; max: number }> = {};

    metrics.value.forEach(metric => {
      if (!summary[metric.name]) {
        summary[metric.name] = {
          avg: 0,
          count: 0,
          min: Infinity,
          max: -Infinity,
        };
      }

      const stat = summary[metric.name];
      stat.count += 1;
      stat.min = Math.min(stat.min, metric.value);
      stat.max = Math.max(stat.max, metric.value);
      stat.avg = (stat.avg * (stat.count - 1) + metric.value) / stat.count;
    });

    return summary;
  };

  // Clear all metrics
  const clearMetrics = (): void => {
    metrics.value = [];
    activeTimers.clear();
  };

  // Track page load performance
  if (trackPageLoad) {
    onMounted(async () => {
      await nextTick();
      
      // Track initial page load using Navigation Timing API
      if (typeof window !== 'undefined' && window.performance) {
        const perfData = performance.getEntriesByType('navigation')[0] as PerformanceNavigationTiming;
        
        if (perfData) {
          recordMetric('page_load_total', perfData.loadEventEnd - perfData.navigationStart, {
            type: 'page_load',
            url: window.location.pathname,
          });

          recordMetric('dom_content_loaded', perfData.domContentLoadedEventEnd - perfData.navigationStart, {
            type: 'dom_ready',
            url: window.location.pathname,
          });

          recordMetric('first_contentful_paint', perfData.domContentLoadedEventStart - perfData.navigationStart, {
            type: 'rendering',
            url: window.location.pathname,
          });
        }
      }
    });
  }

  // Track API calls by intercepting fetch
  if (trackApiCalls && typeof window !== 'undefined') {
    const originalFetch = window.fetch;
    
    window.fetch = async (...args): Promise<Response> => {
      const startTime = performance.now();
      const url = typeof args[0] === 'string' ? args[0] : args[0].url;
      
      try {
        const response = await originalFetch.apply(window, args);
        const endTime = performance.now();
        
        recordMetric('api_call', endTime - startTime, {
          url,
          status: response.status,
          method: args[1]?.method || 'GET',
          type: 'api_success',
        });
        
        return response;
      } catch (error) {
        const endTime = performance.now();
        
        recordMetric('api_call_error', endTime - startTime, {
          url,
          method: args[1]?.method || 'GET',
          type: 'api_error',
          error: (error as Error).message,
        });
        
        throw error;
      }
    };

    // Restore original fetch on unmount
    onUnmounted(() => {
      window.fetch = originalFetch;
    });
  }

  // Track user interactions
  if (trackUserInteractions && typeof window !== 'undefined') {
    const trackInteraction = (event: Event) => {
      recordMetric('user_interaction', performance.now(), {
        type: event.type,
        target: (event.target as Element)?.tagName?.toLowerCase(),
        timestamp: Date.now(),
      });
    };

    const interactionEvents = ['click', 'keydown', 'scroll', 'touchstart'];
    
    onMounted(() => {
      interactionEvents.forEach(eventType => {
        document.addEventListener(eventType, trackInteraction, { passive: true });
      });
    });

    onUnmounted(() => {
      interactionEvents.forEach(eventType => {
        document.removeEventListener(eventType, trackInteraction);
      });
    });
  }

  // Track Core Web Vitals if available
  if (trackPageLoad && typeof window !== 'undefined') {
    onMounted(() => {
      // Track Largest Contentful Paint (LCP)
      if ('PerformanceObserver' in window) {
        try {
          const lcpObserver = new PerformanceObserver((entryList) => {
            const entries = entryList.getEntries();
            const lastEntry = entries[entries.length - 1];
            
            recordMetric('largest_contentful_paint', lastEntry.startTime, {
              type: 'core_web_vital',
              metric: 'LCP',
            });
          });
          
          lcpObserver.observe({ entryTypes: ['largest-contentful-paint'] });
          
          // Track Cumulative Layout Shift (CLS)
          let clsValue = 0;
          const clsObserver = new PerformanceObserver((entryList) => {
            for (const entry of entryList.getEntries()) {
              if (!(entry as any).hadRecentInput) {
                clsValue += (entry as any).value;
              }
            }
            
            recordMetric('cumulative_layout_shift', clsValue, {
              type: 'core_web_vital',
              metric: 'CLS',
            });
          });
          
          clsObserver.observe({ entryTypes: ['layout-shift'] });
          
          // Clean up observers
          onUnmounted(() => {
            lcpObserver.disconnect();
            clsObserver.disconnect();
          });
        } catch (error) {
          console.warn('Performance observation not fully supported:', error);
        }
      }
    });
  }

  return {
    metrics: metrics.value,
    startTimer,
    recordMetric,
    getMetricsSummary,
    clearMetrics,
  };
}

// Helper function to create a performance-aware async operation wrapper
export function withPerformanceTracking<T extends (...args: any[]) => Promise<any>>(
  operation: T,
  operationName: string,
  performanceMonitor: UsePerformanceMonitorResult
): T {
  return (async (...args: Parameters<T>): Promise<ReturnType<T>> => {
    const endTimer = performanceMonitor.startTimer(operationName, {
      args: args.length,
      timestamp: Date.now(),
    });

    try {
      const result = await operation(...args);
      endTimer();
      return result;
    } catch (error) {
      endTimer();
      performanceMonitor.recordMetric(`${operationName}_error`, 0, {
        error: (error as Error).message,
        type: 'operation_error',
      });
      throw error;
    }
  }) as T;
}