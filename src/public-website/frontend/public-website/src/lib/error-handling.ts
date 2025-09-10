// Contract-compliant error handling using generated TypeScript client types
import type { InlineObject1Error } from '@international-center/public-api-client'

// Error types matching contract specifications
export interface ContractError {
  code: string
  message: string
  correlation_id: string
  timestamp: string
  details?: Record<string, unknown>
}

// Contract-compliant error response wrapper
export interface ApiErrorResponse {
  error: ContractError
  status: number
  isContractError: boolean
}

// Error classification for consistent handling
export enum ErrorType {
  VALIDATION = 'validation',
  NOT_FOUND = 'not_found',
  UNAUTHORIZED = 'unauthorized',
  FORBIDDEN = 'forbidden',
  TIMEOUT = 'timeout',
  RATE_LIMIT = 'rate_limit',
  SERVER_ERROR = 'server_error',
  NETWORK_ERROR = 'network_error',
  CONTRACT_VIOLATION = 'contract_violation'
}

// Contract error parser and handler
export class ContractErrorHandler {
  
  // Parse error response from contract client
  static parseContractError(error: unknown): ApiErrorResponse {
    if (this.isContractErrorResponse(error)) {
      return {
        error: error.error,
        status: error.status || 500,
        isContractError: true
      }
    }

    if (error instanceof Response) {
      return {
        error: {
          code: `HTTP_${error.status}`,
          message: error.statusText || 'HTTP error',
          correlation_id: error.headers.get('X-Correlation-ID') || 'unknown',
          timestamp: new Date().toISOString()
        },
        status: error.status,
        isContractError: false
      }
    }

    if (error instanceof Error) {
      return {
        error: {
          code: 'CLIENT_ERROR',
          message: error.message,
          correlation_id: 'unknown',
          timestamp: new Date().toISOString(),
          details: { stack: error.stack }
        },
        status: 0,
        isContractError: false
      }
    }

    return {
      error: {
        code: 'UNKNOWN_ERROR',
        message: 'An unknown error occurred',
        correlation_id: 'unknown',
        timestamp: new Date().toISOString()
      },
      status: 500,
      isContractError: false
    }
  }

  // Check if error is a contract-compliant error response
  static isContractErrorResponse(error: unknown): error is { error: InlineObject1Error } {
    return (
      typeof error === 'object' && 
      error !== null && 
      'error' in error &&
      typeof (error as any).error === 'object' &&
      'code' in (error as any).error &&
      'message' in (error as any).error &&
      'correlationId' in (error as any).error
    )
  }

  // Classify error type for handling strategy
  static classifyError(apiError: ApiErrorResponse): ErrorType {
    const code = apiError.error.code.toUpperCase()
    const status = apiError.status

    if (code.includes('VALIDATION') || status === 400) {
      return ErrorType.VALIDATION
    }
    if (code.includes('NOT_FOUND') || status === 404) {
      return ErrorType.NOT_FOUND
    }
    if (code.includes('UNAUTHORIZED') || status === 401) {
      return ErrorType.UNAUTHORIZED
    }
    if (code.includes('FORBIDDEN') || status === 403) {
      return ErrorType.FORBIDDEN
    }
    if (code.includes('TIMEOUT') || status === 408) {
      return ErrorType.TIMEOUT
    }
    if (code.includes('RATE_LIMIT') || status === 429) {
      return ErrorType.RATE_LIMIT
    }
    if (status >= 500) {
      return ErrorType.SERVER_ERROR
    }
    if (status === 0) {
      return ErrorType.NETWORK_ERROR
    }

    return ErrorType.SERVER_ERROR
  }

  // Generate user-friendly error message
  static getUserFriendlyMessage(apiError: ApiErrorResponse, context?: string): string {
    const errorType = this.classifyError(apiError)
    const baseContext = context || 'operation'

    switch (errorType) {
      case ErrorType.VALIDATION:
        return `Please check your input and try again. ${apiError.error.message}`
      
      case ErrorType.NOT_FOUND:
        return `The requested ${baseContext} could not be found.`
      
      case ErrorType.UNAUTHORIZED:
        return `Please sign in to access this ${baseContext}.`
      
      case ErrorType.FORBIDDEN:
        return `You don't have permission to access this ${baseContext}.`
      
      case ErrorType.TIMEOUT:
        return `The ${baseContext} is taking too long to respond. Please try again.`
      
      case ErrorType.RATE_LIMIT:
        return `Too many requests. Please wait a moment before trying again.`
      
      case ErrorType.NETWORK_ERROR:
        return `Network connection error. Please check your internet connection.`
      
      case ErrorType.CONTRACT_VIOLATION:
        return `The server response doesn't match the expected format. Please contact support.`
      
      case ErrorType.SERVER_ERROR:
      default:
        return `A server error occurred. Please try again later.`
    }
  }

  // Enhanced error logging with contract context
  static logError(apiError: ApiErrorResponse, context?: string, additionalInfo?: Record<string, unknown>) {
    const errorType = this.classifyError(apiError)
    
    console.error(`🚨 [ContractError] ${errorType}:`, {
      code: apiError.error.code,
      message: apiError.error.message,
      correlation_id: apiError.error.correlation_id,
      timestamp: apiError.error.timestamp,
      status: apiError.status,
      context: context || 'unknown',
      is_contract_error: apiError.isContractError,
      details: apiError.error.details,
      additional_info: additionalInfo
    })
  }

  // Composable for Vue components
  static createErrorComposable(context: string) {
    return {
      handleApiError: (error: unknown, additionalInfo?: Record<string, unknown>) => {
        const apiError = this.parseContractError(error)
        this.logError(apiError, context, additionalInfo)
        return this.getUserFriendlyMessage(apiError, context)
      },
      
      isRetryableError: (error: unknown): boolean => {
        const apiError = this.parseContractError(error)
        const errorType = this.classifyError(apiError)
        return [ErrorType.TIMEOUT, ErrorType.NETWORK_ERROR, ErrorType.SERVER_ERROR].includes(errorType)
      },
      
      shouldShowToUser: (error: unknown): boolean => {
        const apiError = this.parseContractError(error)
        const errorType = this.classifyError(apiError)
        return ![ErrorType.CONTRACT_VIOLATION].includes(errorType)
      }
    }
  }
}

// Convenience exports
export const handleApiError = ContractErrorHandler.handleApiError
export const isContractError = ContractErrorHandler.isContractErrorResponse
export const classifyError = ContractErrorHandler.classifyError
export const createErrorHandler = ContractErrorHandler.createErrorComposable