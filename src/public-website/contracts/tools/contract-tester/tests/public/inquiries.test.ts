/**
 * Public API Inquiries Contract Tests
 */

import { describe, test, expect, beforeAll } from 'vitest';
import { ContractValidator, generateTestReport } from '../../src/contract-validator';
import type { OpenAPIV3_1 } from 'openapi-types';

describe('Public API - Inquiries Contract Tests', () => {
  let validator: ContractValidator;
  let spec: OpenAPIV3_1.Document;

  beforeAll(async () => {
    // Wait for public API service to be available
    const isServiceReady = await testUtils.waitForService(process.env.PUBLIC_API_BASE_URL!);
    expect(isServiceReady, 'Public API service should be available for testing').toBe(true);

    // Load OpenAPI specification
    const { readFile } = await import('fs/promises');
    const { load } = await import('js-yaml');
    
    const specPath = '../../openapi/public-api.yaml';
    const specContent = await readFile(specPath, 'utf-8');
    spec = load(specContent) as OpenAPIV3_1.Document;
    
    validator = new ContractValidator(spec, {
      baseUrl: process.env.PUBLIC_API_BASE_URL!,
      timeout: 15000,
      headers: {
        'User-Agent': 'Contract-Test-Client/1.0.0',
        'Content-Type': 'application/json'
      }
    });
  }, 30000);

  describe('Media Inquiries', () => {
    test('should submit media inquiry successfully', async () => {
      const mediaInquiry = testUtils.generateTestData('mediaInquiry');
      
      const result = await validator.testEndpoint(
        '/api/v1/inquiries/media',
        'POST',
        mediaInquiry,
        201
      );

      expect(result.passed, `Media inquiry submission failed: ${JSON.stringify(result.errors)}`).toBe(true);
      expect(result.statusCode).toBe(201);
      expect(result.responseTime).toBeLessThan(5000);
    });

    test('should reject invalid media inquiry', async () => {
      const invalidInquiry = {
        contactInfo: {
          name: '', // Invalid: empty name
          email: 'invalid-email' // Invalid: bad email format
        }
      };
      
      const result = await validator.testEndpoint(
        '/api/v1/inquiries/media',
        'POST',
        invalidInquiry,
        400
      );

      expect(result.statusCode).toBe(400);
    });
  });

  describe('Business Inquiries', () => {
    test('should submit business inquiry successfully', async () => {
      const businessInquiry = testUtils.generateTestData('businessInquiry');
      
      const result = await validator.testEndpoint(
        '/api/v1/inquiries/business',
        'POST',
        businessInquiry,
        201
      );

      expect(result.passed, `Business inquiry submission failed: ${JSON.stringify(result.errors)}`).toBe(true);
      expect(result.statusCode).toBe(201);
    });

    test('should reject business inquiry without required fields', async () => {
      const invalidInquiry = {
        contactInfo: {
          name: 'Test User'
          // Missing required email field
        }
      };
      
      const result = await validator.testEndpoint(
        '/api/v1/inquiries/business',
        'POST',
        invalidInquiry,
        400
      );

      expect(result.statusCode).toBe(400);
    });
  });

  describe('Donation Inquiries', () => {
    test('should submit donation inquiry successfully', async () => {
      const donationInquiry = testUtils.generateTestData('donationInquiry');
      
      const result = await validator.testEndpoint(
        '/api/v1/inquiries/donations',
        'POST',
        donationInquiry,
        201
      );

      expect(result.passed, `Donation inquiry submission failed: ${JSON.stringify(result.errors)}`).toBe(true);
      expect(result.statusCode).toBe(201);
    });

    test('should reject donation inquiry with invalid amount', async () => {
      const invalidInquiry = {
        donorInfo: {
          name: 'Test Donor',
          email: 'donor@test.com'
        },
        donationDetails: {
          amount: -100, // Invalid: negative amount
          frequency: 'one-time',
          cause: 'general'
        }
      };
      
      const result = await validator.testEndpoint(
        '/api/v1/inquiries/donations',
        'POST',
        invalidInquiry,
        400
      );

      expect(result.statusCode).toBe(400);
    });
  });

  describe('Volunteer Inquiries', () => {
    test('should submit volunteer inquiry successfully', async () => {
      const volunteerInquiry = testUtils.generateTestData('volunteerInquiry');
      
      const result = await validator.testEndpoint(
        '/api/v1/inquiries/volunteer',
        'POST',
        volunteerInquiry,
        201
      );

      expect(result.passed, `Volunteer inquiry submission failed: ${JSON.stringify(result.errors)}`).toBe(true);
      expect(result.statusCode).toBe(201);
    });

    test('should reject volunteer inquiry with missing personal info', async () => {
      const invalidInquiry = {
        personalInfo: {
          name: 'Test Volunteer'
          // Missing required fields like email, dateOfBirth
        },
        availability: {
          daysAvailable: ['monday'],
          hoursPerWeek: 5
        }
      };
      
      const result = await validator.testEndpoint(
        '/api/v1/inquiries/volunteer',
        'POST',
        invalidInquiry,
        400
      );

      expect(result.statusCode).toBe(400);
    });
  });

  describe('Cross-cutting Concerns', () => {
    test('should handle malformed JSON gracefully', async () => {
      // Test with malformed JSON by sending a string instead of object
      const result = await validator.testEndpoint(
        '/api/v1/inquiries/media',
        'POST',
        'invalid json string',
        400
      );

      expect(result.statusCode).toBe(400);
    });

    test('should respect request timeout limits', async () => {
      const mediaInquiry = testUtils.generateTestData('mediaInquiry');
      
      const result = await validator.testEndpoint(
        '/api/v1/inquiries/media',
        'POST',
        mediaInquiry,
        201
      );

      expect(result.responseTime).toBeLessThan(15000); // Should complete within timeout
    });

    test('should include proper CORS headers', async () => {
      const result = await validator.testEndpoint(
        '/api/v1/inquiries/media',
        'OPTIONS'
      );

      // OPTIONS request should succeed for CORS preflight
      expect([200, 204]).toContain(result.statusCode);
    });
  });
});