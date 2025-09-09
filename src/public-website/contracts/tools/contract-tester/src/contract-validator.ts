/**
 * Contract validation utilities for testing API compliance
 */

import axios, { AxiosResponse, AxiosError } from 'axios';
import type { OpenAPIV3_1 } from 'openapi-types';

export interface ContractTestConfig {
  baseUrl: string;
  timeout?: number;
  headers?: Record<string, string>;
  retries?: number;
}

export interface ValidationError {
  path: string;
  message: string;
  expected: any;
  actual: any;
}

export interface ContractTestResult {
  passed: boolean;
  endpoint: string;
  method: string;
  statusCode: number;
  errors: ValidationError[];
  responseTime: number;
}

export class ContractValidator {
  private config: ContractTestConfig;
  private spec: OpenAPIV3_1.Document;

  constructor(spec: OpenAPIV3_1.Document, config: ContractTestConfig) {
    this.spec = spec;
    this.config = {
      timeout: 15000,
      retries: 3,
      ...config
    };
  }

  async testEndpoint(
    path: string,
    method: string,
    requestData?: any,
    expectedStatus: number = 200,
    headers?: Record<string, string>
  ): Promise<ContractTestResult> {
    const startTime = Date.now();
    const result: ContractTestResult = {
      passed: false,
      endpoint: path,
      method: method.toUpperCase(),
      statusCode: 0,
      errors: [],
      responseTime: 0
    };

    try {
      const response = await this.makeRequest(path, method, requestData, headers);
      result.statusCode = response.status;
      result.responseTime = Date.now() - startTime;

      // Validate response status
      if (response.status !== expectedStatus) {
        result.errors.push({
          path: 'status',
          message: `Expected status ${expectedStatus}, got ${response.status}`,
          expected: expectedStatus,
          actual: response.status
        });
      }

      // Validate response schema
      const schemaErrors = this.validateResponseSchema(path, method, response.status, response.data);
      result.errors.push(...schemaErrors);

      // Validate response headers
      const headerErrors = this.validateResponseHeaders(path, method, response.status, response.headers);
      result.errors.push(...headerErrors);

      result.passed = result.errors.length === 0;
    } catch (error) {
      result.responseTime = Date.now() - startTime;
      
      if (error instanceof AxiosError) {
        result.statusCode = error.response?.status || 0;
        result.errors.push({
          path: 'request',
          message: error.message,
          expected: 'successful request',
          actual: error.code || 'unknown error'
        });
      } else {
        result.errors.push({
          path: 'request',
          message: 'Unexpected error occurred',
          expected: 'successful request',
          actual: error
        });
      }
    }

    return result;
  }

  async testHealthEndpoint(): Promise<ContractTestResult> {
    return this.testEndpoint('/health', 'GET');
  }

  async testAllEndpoints(sampleData?: Record<string, any>): Promise<ContractTestResult[]> {
    const results: ContractTestResult[] = [];
    
    if (!this.spec.paths) {
      return results;
    }

    for (const [path, pathItem] of Object.entries(this.spec.paths)) {
      if (!pathItem || typeof pathItem !== 'object') continue;

      for (const [method, operation] of Object.entries(pathItem)) {
        if (!operation || typeof operation !== 'object' || !('responses' in operation)) continue;

        // Skip parameters and other non-operation properties
        if (!['get', 'post', 'put', 'delete', 'patch'].includes(method)) continue;

        try {
          const requestData = this.generateRequestData(operation, sampleData);
          const expectedStatus = this.getExpectedSuccessStatus(operation);
          
          const result = await this.testEndpoint(path, method, requestData, expectedStatus);
          results.push(result);
        } catch (error) {
          results.push({
            passed: false,
            endpoint: path,
            method: method.toUpperCase(),
            statusCode: 0,
            errors: [{
              path: 'test_setup',
              message: `Failed to test endpoint: ${error}`,
              expected: 'successful test',
              actual: error
            }],
            responseTime: 0
          });
        }
      }
    }

    return results;
  }

  private async makeRequest(
    path: string,
    method: string,
    data?: any,
    additionalHeaders?: Record<string, string>
  ): Promise<AxiosResponse> {
    const url = `${this.config.baseUrl}${path}`;
    const headers = {
      'Content-Type': 'application/json',
      'Accept': 'application/json',
      ...this.config.headers,
      ...additionalHeaders
    };

    const config = {
      url,
      method: method.toLowerCase(),
      headers,
      timeout: this.config.timeout,
      data: ['post', 'put', 'patch'].includes(method.toLowerCase()) ? data : undefined,
      params: ['get', 'delete'].includes(method.toLowerCase()) ? data : undefined
    };

    return axios.request(config);
  }

  private validateResponseSchema(
    path: string,
    method: string,
    statusCode: number,
    responseData: any
  ): ValidationError[] {
    const errors: ValidationError[] = [];
    
    try {
      const operation = this.getOperation(path, method);
      if (!operation || !operation.responses) {
        return errors;
      }

      const responseSpec = operation.responses[statusCode] || operation.responses['default'];
      if (!responseSpec || typeof responseSpec !== 'object') {
        return errors;
      }

      // Basic validation - in a full implementation, you'd use a JSON Schema validator
      if ('content' in responseSpec && responseSpec.content) {
        const contentSpec = responseSpec.content['application/json'];
        if (contentSpec && 'schema' in contentSpec) {
          // Validate that response has expected structure
          if (typeof responseData !== 'object') {
            errors.push({
              path: 'response.body',
              message: 'Response should be a JSON object',
              expected: 'object',
              actual: typeof responseData
            });
          }
        }
      }
    } catch (error) {
      errors.push({
        path: 'schema_validation',
        message: `Schema validation error: ${error}`,
        expected: 'valid schema',
        actual: error
      });
    }

    return errors;
  }

  private validateResponseHeaders(
    path: string,
    method: string,
    statusCode: number,
    responseHeaders: Record<string, string>
  ): ValidationError[] {
    const errors: ValidationError[] = [];

    // Check for required headers
    const requiredHeaders = ['content-type'];
    
    for (const header of requiredHeaders) {
      if (!responseHeaders[header] && !responseHeaders[header.toLowerCase()]) {
        errors.push({
          path: `headers.${header}`,
          message: `Missing required header: ${header}`,
          expected: header,
          actual: 'missing'
        });
      }
    }

    return errors;
  }

  private getOperation(path: string, method: string): OpenAPIV3_1.OperationObject | undefined {
    if (!this.spec.paths || !this.spec.paths[path]) {
      return undefined;
    }

    const pathItem = this.spec.paths[path];
    if (!pathItem || typeof pathItem !== 'object') {
      return undefined;
    }

    const operation = pathItem[method.toLowerCase() as keyof OpenAPIV3_1.PathItemObject];
    return operation as OpenAPIV3_1.OperationObject | undefined;
  }

  private generateRequestData(operation: OpenAPIV3_1.OperationObject, sampleData?: Record<string, any>): any {
    // Basic request data generation - would be more sophisticated in practice
    if (operation.requestBody && typeof operation.requestBody === 'object' && 'content' in operation.requestBody) {
      const content = operation.requestBody.content?.['application/json'];
      if (content && 'schema' in content) {
        return sampleData?.requestBody || {};
      }
    }
    
    return undefined;
  }

  private getExpectedSuccessStatus(operation: OpenAPIV3_1.OperationObject): number {
    if (!operation.responses) return 200;
    
    // Look for success status codes (2xx)
    const successCodes = Object.keys(operation.responses).filter(code => 
      code.startsWith('2') && code !== 'default'
    );
    
    if (successCodes.length > 0) {
      return parseInt(successCodes[0]);
    }
    
    return 200;
  }
}

export function generateTestReport(results: ContractTestResult[]): string {
  const passed = results.filter(r => r.passed).length;
  const failed = results.length - passed;
  
  let report = `Contract Test Report\n`;
  report += `==================\n`;
  report += `Total Tests: ${results.length}\n`;
  report += `Passed: ${passed}\n`;
  report += `Failed: ${failed}\n`;
  report += `Success Rate: ${((passed / results.length) * 100).toFixed(2)}%\n\n`;
  
  if (failed > 0) {
    report += `Failed Tests:\n`;
    report += `-------------\n`;
    
    results.filter(r => !r.passed).forEach(result => {
      report += `${result.method} ${result.endpoint} (Status: ${result.statusCode})\n`;
      result.errors.forEach(error => {
        report += `  - ${error.path}: ${error.message}\n`;
      });
      report += `\n`;
    });
  }
  
  return report;
}