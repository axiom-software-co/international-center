<template>
  <ErrorBoundary
    :error-title="errorTitle"
    :fallback-route="fallbackRoute"
    @error="handleGlobalError"
  >
    <div class="optimized-data-list">
      <!-- Loading State -->
      <LoadingState
        v-if="loading"
        :variant="loadingVariant"
        :items="loadingItems"
        :message="loadingMessage"
      />
      
      <!-- Error State -->
      <div v-else-if="error" class="error-container p-6 text-center">
        <div class="max-w-md mx-auto">
          <h3 class="text-lg font-medium text-red-600 mb-2">
            {{ errorTitle || 'Error Loading Data' }}
          </h3>
          <p class="text-gray-600 mb-4">
            {{ errorMessage }}
          </p>
          <div class="flex justify-center gap-3">
            <button
              @click="retryLoad"
              :disabled="isRetrying"
              class="px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700 disabled:opacity-50 disabled:cursor-not-allowed flex items-center gap-2"
            >
              <svg
                v-if="isRetrying"
                class="animate-spin h-4 w-4"
                fill="none"
                viewBox="0 0 24 24"
              >
                <circle
                  class="opacity-25"
                  cx="12"
                  cy="12"
                  r="10"
                  stroke="currentColor"
                  stroke-width="4"
                ></circle>
                <path
                  class="opacity-75"
                  fill="currentColor"
                  d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"
                ></path>
              </svg>
              {{ isRetrying ? 'Retrying...' : 'Retry' }}
            </button>
          </div>
        </div>
      </div>
      
      <!-- Empty State -->
      <div v-else-if="items.length === 0" class="empty-state p-6 text-center">
        <div class="max-w-md mx-auto">
          <div class="mb-4">
            <svg
              class="mx-auto h-12 w-12 text-gray-400"
              fill="none"
              stroke="currentColor"
              viewBox="0 0 24 24"
            >
              <path
                stroke-linecap="round"
                stroke-linejoin="round"
                stroke-width="2"
                d="M20 13V6a2 2 0 00-2-2H6a2 2 0 00-2 2v7m16 0v5a2 2 0 01-2 2H6a2 2 0 01-2-2v-5m16 0h-2M4 13h2m0 0V9a2 2 0 012-2h2m0 0V6a2 2 0 012-2h2a2 2 0 012 2v1M9 7h6"
              />
            </svg>
          </div>
          <h3 class="text-lg font-medium text-gray-900 mb-2">
            {{ emptyTitle }}
          </h3>
          <p class="text-gray-600">
            {{ emptyMessage }}
          </p>
        </div>
      </div>
      
      <!-- Data List -->
      <div v-else class="data-list">
        <!-- Performance Stats (Development only) -->
        <div
          v-if="isDevelopment && showStats"
          class="performance-stats mb-4 p-3 bg-gray-100 dark:bg-gray-800 rounded text-xs"
        >
          <div class="grid grid-cols-3 gap-4">
            <div>
              <strong>Cache Hit Rate:</strong> {{ (cacheHitRate * 100).toFixed(1) }}%
            </div>
            <div>
              <strong>Avg Response:</strong> {{ averageResponseTime.toFixed(0) }}ms
            </div>
            <div>
              <strong>Error Rate:</strong> {{ (errorRate * 100).toFixed(1) }}%
            </div>
          </div>
        </div>
        
        <!-- Virtual Scrolling Container (for large lists) -->
        <div
          v-if="enableVirtualScroll && items.length > virtualScrollThreshold"
          ref="virtualScrollContainer"
          :style="{ height: `${virtualScrollHeight}px` }"
          class="overflow-auto"
          @scroll="handleScroll"
        >
          <div :style="{ height: `${totalHeight}px`, position: 'relative' }">
            <div :style="{ transform: `translateY(${offsetY}px)` }">
              <slot
                :items="visibleItems"
                :loading="loading"
                :error="error"
                :retry="retryLoad"
              />
            </div>
          </div>
        </div>
        
        <!-- Regular Rendering -->
        <div v-else>
          <slot
            :items="items"
            :loading="loading"
            :error="error"
            :retry="retryLoad"
          />
        </div>
        
        <!-- Load More (if infinite scroll is disabled) -->
        <div
          v-if="hasMore && !infiniteScroll"
          class="load-more-container mt-6 text-center"
        >
          <button
            @click="loadMore"
            :disabled="loadingMore"
            class="px-4 py-2 border border-gray-300 rounded-md text-gray-700 hover:bg-gray-50 disabled:opacity-50 disabled:cursor-not-allowed"
          >
            {{ loadingMore ? 'Loading...' : 'Load More' }}
          </button>
        </div>
      </div>
    </div>
  </ErrorBoundary>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted, watch, nextTick } from 'vue';
import ErrorBoundary from './ErrorBoundary.vue';
import LoadingState from './LoadingState.vue';
import { usePerformanceMonitor } from '../../composables/performance/usePerformanceMonitor';
import { useErrorHandler } from '../../composables/performance/useErrorHandler';
import { useSmartCache } from '../../composables/performance/useSmartCache';

interface OptimizedDataListProps {
  // Data props
  items: any[];
  loading: boolean;
  error: Error | null;
  hasMore?: boolean;
  loadingMore?: boolean;
  
  // Loading props
  loadingVariant?: 'card' | 'list' | 'table' | 'article' | 'form' | 'spinner';
  loadingItems?: number;
  loadingMessage?: string;
  
  // Error props
  errorTitle?: string;
  fallbackRoute?: string;
  
  // Empty state props
  emptyTitle?: string;
  emptyMessage?: string;
  
  // Performance props
  enableVirtualScroll?: boolean;
  virtualScrollThreshold?: number;
  virtualScrollHeight?: number;
  itemHeight?: number;
  enableInfiniteScroll?: boolean;
  infiniteScrollThreshold?: number;
  
  // Debug props
  showStats?: boolean;
}

interface OptimizedDataListEmits {
  (e: 'retry'): void;
  (e: 'load-more'): void;
  (e: 'error', error: Error): void;
}

const props = withDefaults(defineProps<OptimizedDataListProps>(), {
  hasMore: false,
  loadingMore: false,
  loadingVariant: 'card',
  loadingItems: 3,
  loadingMessage: '',
  errorTitle: '',
  fallbackRoute: '/',
  emptyTitle: 'No items found',
  emptyMessage: 'There are no items to display at the moment.',
  enableVirtualScroll: false,
  virtualScrollThreshold: 100,
  virtualScrollHeight: 400,
  itemHeight: 80,
  enableInfiniteScroll: false,
  infiniteScrollThreshold: 200,
  showStats: false,
});

const emit = defineEmits<OptimizedDataListEmits>();

// Performance monitoring
const performanceMonitor = usePerformanceMonitor({
  trackApiCalls: true,
  trackUserInteractions: true,
  sampleRate: 0.1,
});

const errorHandler = useErrorHandler({
  enablePerformanceTracking: true,
  performanceMonitor,
});

// Environment detection
const isDevelopment = computed(() => {
  return import.meta.env?.DEV || process.env.NODE_ENV === 'development';
});

// Error state
const errorMessage = computed(() => {
  return props.error ? errorHandler.getDisplayMessage(props.error) : '';
});

const isRetrying = computed(() => errorHandler.errorState.isRetrying);

// Performance metrics
const cacheHitRate = ref(0);
const averageResponseTime = ref(0);
const errorRate = ref(0);

// Update performance metrics periodically
if (props.showStats && isDevelopment.value) {
  const updateMetrics = () => {
    const summary = performanceMonitor.getMetricsSummary();
    
    // Mock cache hit rate (would come from actual cache in real implementation)
    cacheHitRate.value = Math.random() * 0.3 + 0.7; // 70-100%
    
    const apiCallStats = summary['api_call'];
    averageResponseTime.value = apiCallStats?.avg || 0;
    
    const errorStats = summary['api_call_error'];
    const successStats = summary['api_call'];
    const totalErrors = errorStats?.count || 0;
    const totalSuccess = successStats?.count || 0;
    const total = totalErrors + totalSuccess;
    errorRate.value = total > 0 ? totalErrors / total : 0;
  };

  let metricsInterval: number;
  onMounted(() => {
    metricsInterval = setInterval(updateMetrics, 2000);
  });
  onUnmounted(() => {
    clearInterval(metricsInterval);
  });
}

// Virtual scrolling
const virtualScrollContainer = ref<HTMLElement>();
const offsetY = ref(0);
const visibleItems = ref<any[]>([]);

const totalHeight = computed(() => props.items.length * props.itemHeight);

const calculateVisibleItems = () => {
  if (!props.enableVirtualScroll || !virtualScrollContainer.value) {
    visibleItems.value = props.items;
    return;
  }

  const scrollTop = virtualScrollContainer.value.scrollTop;
  const containerHeight = props.virtualScrollHeight;
  const startIndex = Math.floor(scrollTop / props.itemHeight);
  const endIndex = Math.min(
    startIndex + Math.ceil(containerHeight / props.itemHeight) + 1,
    props.items.length
  );

  offsetY.value = startIndex * props.itemHeight;
  visibleItems.value = props.items.slice(startIndex, endIndex);
};

const handleScroll = () => {
  calculateVisibleItems();
  
  // Infinite scroll
  if (props.enableInfiniteScroll && props.hasMore && !props.loadingMore) {
    const container = virtualScrollContainer.value;
    if (!container) return;

    const scrollPosition = container.scrollTop + container.clientHeight;
    const threshold = container.scrollHeight - props.infiniteScrollThreshold;
    
    if (scrollPosition >= threshold) {
      emit('load-more');
    }
  }
};

// Watch for items changes to recalculate visible items
watch(() => props.items, calculateVisibleItems, { immediate: true });

// Actions
const retryLoad = () => {
  performanceMonitor.recordMetric('user_retry', 1, {
    error: props.error?.message,
    component: 'OptimizedDataList',
  });
  emit('retry');
};

const loadMore = () => {
  performanceMonitor.recordMetric('load_more', 1, {
    currentItems: props.items.length,
    component: 'OptimizedDataList',
  });
  emit('load-more');
};

const handleGlobalError = (error: Error) => {
  emit('error', error);
};

// Intersection Observer for regular infinite scroll (non-virtual)
let intersectionObserver: IntersectionObserver | null = null;

onMounted(() => {
  if (props.enableInfiniteScroll && !props.enableVirtualScroll) {
    intersectionObserver = new IntersectionObserver(
      (entries) => {
        if (entries[0].isIntersecting && props.hasMore && !props.loadingMore) {
          loadMore();
        }
      },
      { threshold: 0.1 }
    );
  }
});

onUnmounted(() => {
  if (intersectionObserver) {
    intersectionObserver.disconnect();
  }
});
</script>

<style scoped>
.optimized-data-list {
  width: 100%;
}

.performance-stats {
  font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', monospace;
  border: 1px dashed #ccc;
}

.virtual-scroll-container {
  position: relative;
  overflow: auto;
}

/* Smooth scrolling performance optimization */
.virtual-scroll-container {
  scroll-behavior: smooth;
  -webkit-overflow-scrolling: touch;
}

/* Optimize rendering performance */
.data-list {
  contain: layout style paint;
}

/* Reduce reflow/repaint on hover */
.load-more-container button {
  transition: background-color 0.15s ease;
  will-change: background-color;
}
</style>