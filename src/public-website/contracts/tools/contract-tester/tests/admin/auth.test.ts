/**
 * Admin API Authentication Contract Tests
 */

import { describe, test, expect, beforeAll } from 'vitest';
import { ContractValidator, generateTestReport } from '../../src/contract-validator';
import type { OpenAPIV3_1 } from 'openapi-types';

describe('Admin API - Authentication Contract Tests', () => {
  let validator: ContractValidator;
  let spec: OpenAPIV3_1.Document;
  let authToken: string;

  beforeAll(async () => {
    // Wait for admin API service to be available
    const isServiceReady = await testUtils.waitForService(process.env.ADMIN_API_BASE_URL!);
    expect(isServiceReady, 'Admin API service should be available for testing').toBe(true);

    // Load OpenAPI specification
    const { readFile } = await import('fs/promises');
    const { load } = await import('js-yaml');
    
    const specPath = '../../openapi/admin-api.yaml';
    const specContent = await readFile(specPath, 'utf-8');
    spec = load(specContent) as OpenAPIV3_1.Document;
    
    validator = new ContractValidator(spec, {
      baseUrl: process.env.ADMIN_API_BASE_URL!,
      timeout: 15000,
      headers: {
        'User-Agent': 'Contract-Test-Client/1.0.0',
        'Content-Type': 'application/json'
      }
    });
  }, 30000);

  describe('Authentication Flow', () => {
    test('should reject login with invalid credentials', async () => {
      const invalidCredentials = {
        email: 'invalid@test.com',
        password: 'wrongpassword'
      };

      const result = await validator.testEndpoint(
        '/api/v1/auth/login',
        'POST',
        invalidCredentials,
        401
      );

      expect(result.statusCode).toBe(401);
      expect(result.responseTime).toBeLessThan(5000);
    });

    test('should reject login with malformed email', async () => {
      const invalidCredentials = {
        email: 'not-an-email',
        password: 'somepassword'
      };

      const result = await validator.testEndpoint(
        '/api/v1/auth/login',
        'POST',
        invalidCredentials,
        400
      );

      expect(result.statusCode).toBe(400);
    });

    test('should reject login with missing password', async () => {
      const invalidCredentials = {
        email: 'admin@test.com'
        // Missing password field
      };

      const result = await validator.testEndpoint(
        '/api/v1/auth/login',
        'POST',
        invalidCredentials,
        400
      );

      expect(result.statusCode).toBe(400);
    });

    test('should handle valid login attempt structure', async () => {
      // Note: For contract testing, we validate the endpoint structure
      // We don't expect this to succeed unless we have test credentials set up
      const testCredentials = {
        email: 'test@example.com',
        password: 'testpassword'
      };

      const result = await validator.testEndpoint(
        '/api/v1/auth/login',
        'POST',
        testCredentials
      );

      // We expect either 200 (successful login) or 401 (invalid credentials)
      // Both are valid responses from a contract perspective
      expect([200, 401]).toContain(result.statusCode);
    });

    test('should reject refresh token without valid token', async () => {
      const result = await validator.testEndpoint(
        '/api/v1/auth/refresh',
        'POST',
        { refreshToken: 'invalid-token' },
        401
      );

      expect(result.statusCode).toBe(401);
    });

    test('should handle logout request structure', async () => {
      // Test logout endpoint structure (should require authentication)
      const result = await validator.testEndpoint(
        '/api/v1/auth/logout',
        'POST'
      );

      // Should return 401 (unauthorized) since we don't have a valid token
      expect(result.statusCode).toBe(401);
    });
  });

  describe('Protected Endpoints Access Control', () => {
    test('should reject access to protected endpoints without token', async () => {
      const result = await validator.testEndpoint('/api/v1/users', 'GET');

      expect(result.statusCode).toBe(401);
    });

    test('should reject access with invalid bearer token', async () => {
      const result = await validator.testEndpoint(
        '/api/v1/users',
        'GET',
        undefined,
        401,
        { 'Authorization': 'Bearer invalid-token' }
      );

      expect(result.statusCode).toBe(401);
    });

    test('should validate bearer token format', async () => {
      const result = await validator.testEndpoint(
        '/api/v1/users',
        'GET',
        undefined,
        401,
        { 'Authorization': 'InvalidFormat token' }
      );

      expect(result.statusCode).toBe(401);
    });
  });

  describe('Security Headers and CORS', () => {
    test('should include security headers in responses', async () => {
      const result = await validator.testEndpoint('/api/v1/auth/login', 'POST', {
        email: 'test@test.com',
        password: 'password'
      });

      // Response should include security headers (validated by ContractValidator)
      expect(result.errors.filter(e => e.path.startsWith('headers')).length).toBe(0);
    });

    test('should handle CORS preflight requests', async () => {
      const result = await validator.testEndpoint('/api/v1/auth/login', 'OPTIONS');

      // OPTIONS request should be handled properly for CORS
      expect([200, 204]).toContain(result.statusCode);
    });

    test('should reject requests with missing content-type for POST', async () => {
      const result = await validator.testEndpoint(
        '/api/v1/auth/login',
        'POST',
        { email: 'test@test.com', password: 'password' },
        400,
        { 'Content-Type': undefined as any }
      );

      // Should handle missing content-type appropriately
      expect([400, 415]).toContain(result.statusCode);
    });
  });

  describe('Rate Limiting', () => {
    test('should handle multiple rapid login attempts', async () => {
      const credentials = {
        email: 'test@test.com',
        password: 'wrongpassword'
      };

      // Make multiple rapid requests to test rate limiting
      const requests = Array.from({ length: 10 }, () =>
        validator.testEndpoint('/api/v1/auth/login', 'POST', credentials)
      );

      const results = await Promise.all(requests);
      
      // Should handle rate limiting gracefully
      const rateLimitedRequests = results.filter(r => r.statusCode === 429);
      const unauthorizedRequests = results.filter(r => r.statusCode === 401);
      
      // Either rate limited or unauthorized responses are acceptable
      expect(rateLimitedRequests.length + unauthorizedRequests.length).toBe(results.length);
    });
  });
});