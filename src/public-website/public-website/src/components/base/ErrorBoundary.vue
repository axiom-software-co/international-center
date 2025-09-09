<template>
  <div v-if="hasError" class="error-boundary">
    <div class="max-w-md mx-auto text-center py-8">
      <div class="mb-4">
        <svg class="mx-auto h-12 w-12 text-red-500" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-2.5L13.732 4c-.77-.833-1.732-.833-2.5 0L4.268 18.5c-.77.833.192 2.5 1.732 2.5z" />
        </svg>
      </div>
      
      <h3 class="text-lg font-medium text-gray-900 dark:text-white mb-2">
        {{ errorTitle }}
      </h3>
      
      <p class="text-gray-600 dark:text-gray-400 mb-4">
        {{ errorMessage }}
      </p>
      
      <div class="flex flex-col sm:flex-row gap-3 justify-center">
        <button
          @click="retry"
          class="inline-flex items-center px-4 py-2 border border-transparent text-sm font-medium rounded-md shadow-sm text-white bg-blue-600 hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500"
        >
          <svg class="-ml-1 mr-2 h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
          </svg>
          Try Again
        </button>
        
        <button
          v-if="showFallback"
          @click="goToFallback"
          class="inline-flex items-center px-4 py-2 border border-gray-300 text-sm font-medium rounded-md text-gray-700 bg-white hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500"
        >
          {{ fallbackText }}
        </button>
      </div>
      
      <!-- Development error details -->
      <details v-if="isDevelopment && errorDetails" class="mt-6 text-left">
        <summary class="cursor-pointer text-sm text-gray-500 hover:text-gray-700">
          Show Error Details
        </summary>
        <pre class="mt-2 p-3 bg-gray-100 dark:bg-gray-800 rounded text-xs text-gray-800 dark:text-gray-200 overflow-auto">{{ errorDetails }}</pre>
      </details>
    </div>
  </div>
  
  <slot v-else />
</template>

<script setup lang="ts">
import { ref, onErrorCaptured, computed } from 'vue';

interface ErrorBoundaryProps {
  fallbackRoute?: string;
  fallbackText?: string;
  errorTitle?: string;
  onError?: (error: Error, instance: any, info: string) => void;
}

const props = withDefaults(defineProps<ErrorBoundaryProps>(), {
  fallbackRoute: '/',
  fallbackText: 'Go Home',
  errorTitle: 'Something went wrong',
});

// Error state
const hasError = ref(false);
const errorMessage = ref('');
const errorDetails = ref<string | null>(null);
const retryKey = ref(0);

// Environment detection
const isDevelopment = computed(() => {
  return import.meta.env?.DEV || process.env.NODE_ENV === 'development';
});

// Show fallback option
const showFallback = computed(() => {
  return props.fallbackRoute && props.fallbackRoute !== '/';
});

// Error capture
onErrorCaptured((error: Error, instance: any, info: string) => {
  hasError.value = true;
  errorMessage.value = error.message || 'An unexpected error occurred';
  
  if (isDevelopment.value) {
    errorDetails.value = `${error.stack || error.message}\n\nComponent Info: ${info}`;
  }
  
  // Call custom error handler if provided
  if (props.onError) {
    props.onError(error, instance, info);
  }
  
  // Log error for monitoring (in production, you might want to send this to a service)
  console.error('Error Boundary caught an error:', {
    error: error.message,
    stack: error.stack,
    info,
    timestamp: new Date().toISOString(),
  });
  
  // Prevent the error from propagating further
  return false;
});

// Actions
const retry = () => {
  hasError.value = false;
  errorMessage.value = '';
  errorDetails.value = null;
  retryKey.value += 1;
};

const goToFallback = () => {
  if (props.fallbackRoute) {
    window.location.href = props.fallbackRoute;
  }
};

// Global error handler for unhandled promise rejections
if (typeof window !== 'undefined') {
  window.addEventListener('unhandledrejection', (event) => {
    hasError.value = true;
    errorMessage.value = event.reason?.message || 'An unhandled promise rejection occurred';
    
    if (isDevelopment.value) {
      errorDetails.value = event.reason?.stack || String(event.reason);
    }
    
    console.error('Unhandled Promise Rejection:', event.reason);
    
    // Prevent default browser behavior
    event.preventDefault();
  });
}
</script>

<style scoped>
.error-boundary {
  min-height: 200px;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 2rem;
}

details pre {
  max-height: 200px;
  overflow-y: auto;
  white-space: pre-wrap;
  word-break: break-word;
}
</style>