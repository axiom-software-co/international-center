// Enhanced Error Handler Composable
// Provides centralized error handling with retry logic, logging, and user-friendly messaging

import { ref, reactive } from 'vue';
import type { UsePerformanceMonitorResult } from './usePerformanceMonitor';

export interface ErrorContext {
  component?: string;
  action?: string;
  userId?: string;
  url?: string;
  timestamp?: number;
  userAgent?: string;
  stack?: string;
  metadata?: Record<string, any>;
}

export interface RetryOptions {
  maxAttempts?: number;
  baseDelay?: number; // Base delay in milliseconds
  maxDelay?: number; // Maximum delay in milliseconds  
  backoffFactor?: number; // Exponential backoff multiplier
  shouldRetry?: (error: Error, attempt: number) => boolean;
}

export interface UseErrorHandlerOptions {
  enableLogging?: boolean;
  enablePerformanceTracking?: boolean;
  performanceMonitor?: UsePerformanceMonitorResult;
  globalErrorHandler?: (error: Error, context: ErrorContext) => void;
}

export interface ErrorState {
  hasError: boolean;
  error: Error | null;
  context: ErrorContext | null;
  isRetrying: boolean;
  retryCount: number;
}

export interface UseErrorHandlerResult {
  // Error state
  errorState: ErrorState;
  
  // Error handling functions
  handleError: (error: Error, context?: ErrorContext) => void;
  clearError: () => void;
  retry: (retryFn: () => Promise<any>) => Promise<any>;
  
  // Async operation wrappers
  withErrorHandling: <T extends (...args: any[]) => Promise<any>>(
    fn: T,
    context?: ErrorContext,
    retryOptions?: RetryOptions
  ) => T;
  
  // User-friendly error messages
  getDisplayMessage: (error: Error) => string;
  
  // Error recovery
  recoverFromError: (recoveryFn: () => Promise<void>) => Promise<void>;
}

// Common error types and their user-friendly messages
const ERROR_MESSAGES = {
  NetworkError: 'Unable to connect to the server. Please check your internet connection.',
  TimeoutError: 'The request took too long to complete. Please try again.',
  ValidationError: 'Please check your input and try again.',
  AuthenticationError: 'You need to log in to perform this action.',
  AuthorizationError: 'You do not have permission to perform this action.',
  NotFoundError: 'The requested resource was not found.',
  RateLimitError: 'Too many requests. Please wait a moment and try again.',
  ServerError: 'An unexpected server error occurred. Please try again later.',
  ClientError: 'There was a problem with your request. Please try again.',
  UnknownError: 'An unexpected error occurred. Please try again.',
} as const;

export function useErrorHandler(options: UseErrorHandlerOptions = {}): UseErrorHandlerResult {
  const {
    enableLogging = true,
    enablePerformanceTracking = false,
    performanceMonitor,
    globalErrorHandler,
  } = options;

  // Error state
  const errorState = reactive<ErrorState>({
    hasError: false,
    error: null,
    context: null,
    isRetrying: false,
    retryCount: 0,
  });

  // Handle an error
  const handleError = (error: Error, context: ErrorContext = {}): void => {
    const enrichedContext: ErrorContext = {
      timestamp: Date.now(),
      url: typeof window !== 'undefined' ? window.location.href : undefined,
      userAgent: typeof navigator !== 'undefined' ? navigator.userAgent : undefined,
      stack: error.stack,
      ...context,
    };

    errorState.hasError = true;
    errorState.error = error;
    errorState.context = enrichedContext;

    // Log error
    if (enableLogging) {
      console.error('Error handled:', {
        message: error.message,
        name: error.name,
        context: enrichedContext,
      });
    }

    // Track error performance impact
    if (enablePerformanceTracking && performanceMonitor) {
      performanceMonitor.recordMetric('error_occurred', 1, {
        errorType: error.name,
        errorMessage: error.message,
        context: enrichedContext.component,
        action: enrichedContext.action,
      });
    }

    // Call global error handler
    if (globalErrorHandler) {
      try {
        globalErrorHandler(error, enrichedContext);
      } catch (handlerError) {
        console.error('Error in global error handler:', handlerError);
      }
    }
  };

  // Clear error state
  const clearError = (): void => {
    errorState.hasError = false;
    errorState.error = null;
    errorState.context = null;
    errorState.isRetrying = false;
    errorState.retryCount = 0;
  };

  // Retry with exponential backoff
  const retry = async (retryFn: () => Promise<any>): Promise<any> => {
    errorState.isRetrying = true;
    errorState.retryCount++;

    try {
      const result = await retryFn();
      clearError();
      return result;
    } catch (error) {
      errorState.isRetrying = false;
      throw error;
    }
  };

  // Get user-friendly error message
  const getDisplayMessage = (error: Error): string => {
    // Check for specific error types
    if (error.name === 'TypeError' && error.message.includes('fetch')) {
      return ERROR_MESSAGES.NetworkError;
    }
    
    if (error.name === 'AbortError' || error.message.includes('timeout')) {
      return ERROR_MESSAGES.TimeoutError;
    }
    
    if (error.name === 'ValidationError' || error.message.includes('validation')) {
      return ERROR_MESSAGES.ValidationError;
    }
    
    if (error.message.includes('401') || error.message.includes('Unauthorized')) {
      return ERROR_MESSAGES.AuthenticationError;
    }
    
    if (error.message.includes('403') || error.message.includes('Forbidden')) {
      return ERROR_MESSAGES.AuthorizationError;
    }
    
    if (error.message.includes('404') || error.message.includes('Not Found')) {
      return ERROR_MESSAGES.NotFoundError;
    }
    
    if (error.message.includes('429') || error.message.includes('Rate limit')) {
      return ERROR_MESSAGES.RateLimitError;
    }
    
    if (error.message.includes('5')) { // 5xx errors
      return ERROR_MESSAGES.ServerError;
    }
    
    if (error.message.includes('4')) { // 4xx errors
      return ERROR_MESSAGES.ClientError;
    }

    // Default fallback
    return ERROR_MESSAGES.UnknownError;
  };

  // Wrap async function with error handling and retry logic
  const withErrorHandling = <T extends (...args: any[]) => Promise<any>>(
    fn: T,
    context: ErrorContext = {},
    retryOptions: RetryOptions = {}
  ): T => {
    const {
      maxAttempts = 3,
      baseDelay = 1000,
      maxDelay = 10000,
      backoffFactor = 2,
      shouldRetry = (error: Error, attempt: number) => {
        // Retry on network errors, timeouts, and 5xx server errors
        return attempt < maxAttempts && (
          error.name === 'NetworkError' ||
          error.name === 'TimeoutError' ||
          error.message.includes('5') ||
          error.message.includes('timeout') ||
          error.message.includes('fetch')
        );
      },
    } = retryOptions;

    return (async (...args: Parameters<T>): Promise<ReturnType<T>> => {
      let lastError: Error;
      
      for (let attempt = 1; attempt <= maxAttempts; attempt++) {
        try {
          clearError(); // Clear any previous errors
          const result = await fn(...args);
          
          if (enablePerformanceTracking && performanceMonitor) {
            performanceMonitor.recordMetric('operation_success', attempt, {
              context: context.component,
              action: context.action,
              attempts: attempt,
            });
          }
          
          return result;
        } catch (error) {
          lastError = error as Error;
          
          // Handle the error
          handleError(lastError, {
            ...context,
            action: context.action || 'async_operation',
            metadata: {
              attempt,
              maxAttempts,
              args: args.length,
            },
          });
          
          // Check if we should retry
          if (shouldRetry(lastError, attempt) && attempt < maxAttempts) {
            // Calculate delay with exponential backoff and jitter
            const delay = Math.min(
              baseDelay * Math.pow(backoffFactor, attempt - 1),
              maxDelay
            );
            const jitter = delay * 0.1 * Math.random(); // Add up to 10% jitter
            
            errorState.isRetrying = true;
            errorState.retryCount = attempt;
            
            await new Promise(resolve => setTimeout(resolve, delay + jitter));
          } else {
            // Max attempts reached or shouldn't retry
            break;
          }
        }
      }
      
      // If we get here, all attempts failed
      throw lastError!;
    }) as T;
  };

  // Recover from error with custom recovery function
  const recoverFromError = async (recoveryFn: () => Promise<void>): Promise<void> => {
    try {
      await recoveryFn();
      clearError();
    } catch (recoveryError) {
      handleError(recoveryError as Error, {
        action: 'error_recovery',
        metadata: {
          originalError: errorState.error?.message,
        },
      });
      throw recoveryError;
    }
  };

  return {
    errorState,
    handleError,
    clearError,
    retry,
    withErrorHandling,
    getDisplayMessage,
    recoverFromError,
  };
}

// Helper function to determine if an error is retryable
export function isRetryableError(error: Error): boolean {
  const retryablePatterns = [
    /network/i,
    /timeout/i,
    /fetch/i,
    /5\d\d/,  // 5xx status codes
    /rate.?limit/i,
  ];
  
  return retryablePatterns.some(pattern => 
    pattern.test(error.message) || pattern.test(error.name)
  );
}

// Helper to create error context from Vue component
export function createErrorContext(
  component: string,
  action: string,
  metadata?: Record<string, any>
): ErrorContext {
  return {
    component,
    action,
    timestamp: Date.now(),
    metadata,
  };
}