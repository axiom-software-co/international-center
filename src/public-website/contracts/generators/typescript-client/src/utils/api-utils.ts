/**
 * API utility functions for error handling, request configuration, and common operations
 */

import type { ApiError, ApiResponse, RequestOptions, ValidationError, ApiValidationError } from '../types/common';

/**
 * Handle API errors and convert them to a standardized format
 */
export function handleApiError(error: any): ApiResponse<any> {
  if (error.response) {
    // HTTP error response
    const apiError: ApiError = {
      code: error.response.status.toString(),
      message: error.response.statusText || 'Request failed',
      details: error.response.data,
      correlation_id: error.response.headers?.['x-correlation-id'] || 'unknown',
      timestamp: new Date().toISOString()
    };
    return { data: null, error: apiError };
  } else if (error.request) {
    // Network error
    const apiError: ApiError = {
      code: 'NETWORK_ERROR',
      message: 'Network request failed',
      details: error.message,
      correlation_id: 'unknown',
      timestamp: new Date().toISOString()
    };
    return { data: null, error: apiError };
  } else {
    // Other error
    const apiError: ApiError = {
      code: 'UNKNOWN_ERROR',
      message: error.message || 'An unexpected error occurred',
      details: error,
      correlation_id: 'unknown',
      timestamp: new Date().toISOString()
    };
    return { data: null, error: apiError };
  }
}

/**
 * Build request options with common configuration
 */
export function buildRequestOptions(
  init: RequestInit | undefined, 
  config: { timeout?: number; defaultHeaders?: Record<string, string> }
): RequestInit {
  const headers = {
    'Content-Type': 'application/json',
    'Accept': 'application/json',
    ...config.defaultHeaders,
    ...init?.headers
  };

  return {
    ...init,
    headers,
    signal: config.timeout ? AbortSignal.timeout(config.timeout) : init?.signal
  };
}

/**
 * Extract validation errors from API response
 */
export function extractValidationErrors(error: ApiError): ValidationError[] {
  if ('validation_errors' in error) {
    return (error as ApiValidationError).validation_errors;
  }
  
  if (error.details?.validation_errors) {
    return error.details.validation_errors;
  }
  
  return [];
}

/**
 * Check if error is a validation error
 */
export function isValidationError(error: ApiError): boolean {
  return error.code === '400' || error.code === 'VALIDATION_ERROR';
}

/**
 * Check if error is an authentication error
 */
export function isAuthError(error: ApiError): boolean {
  return error.code === '401' || error.code === 'UNAUTHORIZED';
}

/**
 * Check if error is a permission error
 */
export function isPermissionError(error: ApiError): boolean {
  return error.code === '403' || error.code === 'FORBIDDEN';
}

/**
 * Check if error is a not found error
 */
export function isNotFoundError(error: ApiError): boolean {
  return error.code === '404' || error.code === 'NOT_FOUND';
}

/**
 * Check if error is a server error
 */
export function isServerError(error: ApiError): boolean {
  const code = parseInt(error.code);
  return code >= 500 && code < 600;
}

/**
 * Format error message for user display
 */
export function formatErrorMessage(error: ApiError): string {
  switch (error.code) {
    case '400':
      return 'Please check your input and try again.';
    case '401':
      return 'Please log in to continue.';
    case '403':
      return 'You do not have permission to perform this action.';
    case '404':
      return 'The requested resource was not found.';
    case '429':
      return 'Too many requests. Please try again later.';
    case '500':
      return 'A server error occurred. Please try again later.';
    case '503':
      return 'The service is temporarily unavailable. Please try again later.';
    case 'NETWORK_ERROR':
      return 'Network connection failed. Please check your internet connection.';
    default:
      return error.message || 'An unexpected error occurred.';
  }
}

/**
 * Create a retry function with exponential backoff
 */
export function createRetryFunction<T>(
  fn: () => Promise<T>, 
  maxRetries: number = 3, 
  baseDelay: number = 1000
): () => Promise<T> {
  return async (): Promise<T> => {
    let lastError: Error;
    
    for (let attempt = 0; attempt <= maxRetries; attempt++) {
      try {
        return await fn();
      } catch (error) {
        lastError = error as Error;
        
        if (attempt === maxRetries) {
          throw lastError;
        }
        
        // Exponential backoff with jitter
        const delay = baseDelay * Math.pow(2, attempt) + Math.random() * 1000;
        await new Promise(resolve => setTimeout(resolve, delay));
      }
    }
    
    throw lastError!;
  };
}

/**
 * Build query parameters from an object
 */
export function buildQueryParams(params: Record<string, any>): string {
  const searchParams = new URLSearchParams();
  
  Object.entries(params).forEach(([key, value]) => {
    if (value !== undefined && value !== null) {
      if (Array.isArray(value)) {
        value.forEach(v => searchParams.append(key, v.toString()));
      } else {
        searchParams.append(key, value.toString());
      }
    }
  });
  
  return searchParams.toString();
}

/**
 * Parse correlation ID from response headers
 */
export function getCorrelationId(response: Response): string {
  return response.headers.get('x-correlation-id') || 
         response.headers.get('X-Correlation-ID') || 
         'unknown';
}

/**
 * Create a request interceptor for logging
 */
export function createRequestLogger(prefix: string = 'API') {
  return (url: RequestInfo | URL, options: RequestInit) => {
    console.log(`${prefix} Request:`, {
      url: url.toString(),
      method: options.method || 'GET',
      headers: options.headers,
      timestamp: new Date().toISOString()
    });
  };
}

/**
 * Create a response interceptor for logging
 */
export function createResponseLogger(prefix: string = 'API') {
  return (response: Response) => {
    console.log(`${prefix} Response:`, {
      url: response.url,
      status: response.status,
      statusText: response.statusText,
      headers: Object.fromEntries(response.headers.entries()),
      timestamp: new Date().toISOString()
    });
  };
}