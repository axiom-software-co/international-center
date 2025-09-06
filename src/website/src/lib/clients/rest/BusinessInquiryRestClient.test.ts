import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { BusinessInquiryRestClient } from './BusinessInquiryRestClient';
import type { BusinessInquiry, BusinessInquirySubmission, InquirySubmissionResponse } from '../inquiries/types';
import { mockFetch } from '../../../test/setup';

// Helper function to create mock responses
const createMockResponse = (data: any, status = 200) => {
  return {
    ok: status >= 200 && status < 300,
    status,
    headers: {
      get: vi.fn((header: string) => {
        if (header === 'content-type') return 'application/json';
        return null;
      })
    },
    json: () => Promise.resolve(data),
    text: () => Promise.resolve(JSON.stringify(data))
  } as Response;
};

describe('BusinessInquiryRestClient', () => {
  let client: BusinessInquiryRestClient;

  const mockBusinessInquiry: BusinessInquiry = {
    inquiry_id: '123e4567-e89b-12d3-a456-426614174000',
    contact_name: 'John Smith',
    email: 'john.smith@company.com',
    organization_name: 'Acme Corporation',
    title: 'Director of Partnerships',
    inquiry_type: 'partnership',
    message: 'We are interested in exploring partnership opportunities with your organization.',
    status: 'new',
    priority: 'medium',
    source: 'website',
    created_at: '2024-03-15T10:00:00Z',
    updated_at: '2024-03-15T10:00:00Z',
    created_by: 'system',
    updated_by: 'system',
    is_deleted: false
  };

  const mockSubmissionData: BusinessInquirySubmission = {
    contact_name: 'John Smith',
    email: 'john.smith@company.com',
    phone: '+1-555-123-4567',
    organization_name: 'Acme Corporation',
    title: 'Director of Partnerships',
    industry: 'Technology',
    inquiry_type: 'partnership',
    message: 'We are interested in exploring partnership opportunities with your organization.'
  };

  const mockSubmissionResponse: InquirySubmissionResponse = {
    business_inquiry: mockBusinessInquiry,
    correlation_id: 'corr-123-456-789',
    success: true,
    message: 'Business inquiry submitted successfully'
  };

  beforeEach(() => {
    client = new BusinessInquiryRestClient();
    mockFetch.mockReset();
  });

  afterEach(() => {
    vi.clearAllMocks();
  });

  describe('submitBusinessInquiry', () => {
    it('should submit business inquiry with valid data', async () => {
      mockFetch.mockResolvedValue(createMockResponse(mockSubmissionResponse));

      const result = await client.submitBusinessInquiry(mockSubmissionData);

      expect(mockFetch).toHaveBeenCalledWith('http://localhost:7220/api/inquiries/business', expect.objectContaining({
        method: 'POST',
        headers: expect.objectContaining({
          'Content-Type': 'application/json',
          'Accept': 'application/json',
          'X-Retry-Attempt': '1'
        }),
        body: JSON.stringify(mockSubmissionData),
        signal: expect.any(AbortSignal)
      }));
      expect(result).toEqual(mockSubmissionResponse);
      expect(result.success).toBe(true);
      expect(result.business_inquiry?.inquiry_type).toBe('partnership');
    });

    it('should handle submission with optional fields', async () => {
      const submissionWithOptionalFields = {
        ...mockSubmissionData,
        industry: 'Healthcare Technology'
      };
      const responseWithOptionals = {
        ...mockSubmissionResponse,
        business_inquiry: {
          ...mockBusinessInquiry,
          industry: 'Healthcare Technology'
        }
      };

      mockFetch.mockResolvedValue(createMockResponse(responseWithOptionals));

      const result = await client.submitBusinessInquiry(submissionWithOptionalFields);

      expect(mockFetch).toHaveBeenCalledWith('http://localhost:7220/api/inquiries/business', expect.objectContaining({
        method: 'POST',
        headers: expect.objectContaining({
          'Content-Type': 'application/json',
          'Accept': 'application/json',
          'X-Retry-Attempt': '1'
        }),
        body: JSON.stringify(submissionWithOptionalFields),
        signal: expect.any(AbortSignal)
      }));
      expect(result.business_inquiry?.industry).toBe('Healthcare Technology');
    });

    it('should handle different inquiry types', async () => {
      const researchInquiry = {
        ...mockSubmissionData,
        inquiry_type: 'research' as const,
        message: 'We would like to collaborate on medical research projects.'
      };
      const researchResponse = {
        ...mockSubmissionResponse,
        business_inquiry: {
          ...mockBusinessInquiry,
          inquiry_type: 'research',
          message: 'We would like to collaborate on medical research projects.'
        }
      };

      mockFetch.mockResolvedValue(createMockResponse(researchResponse));

      const result = await client.submitBusinessInquiry(researchInquiry);

      expect(result.business_inquiry?.inquiry_type).toBe('research');
      expect(result.business_inquiry?.message).toContain('medical research');
    });

    it('should handle submission errors', async () => {
      const errorResponse = {
        error: 'Validation failed',
        correlation_id: 'corr-error-123',
        success: false,
        message: 'Invalid organization name'
      };

      mockFetch.mockRejectedValue(new Error('Network error'));

      await expect(client.submitBusinessInquiry(mockSubmissionData))
        .rejects.toThrow('Network error');
    });

    it('should handle validation errors from backend', async () => {
      const validationErrorResponse = {
        error: 'Validation failed',
        validation_errors: [
          'Organization name is required',
          'Message is too short'
        ],
        correlation_id: 'corr-validation-error',
        success: false
      };

      mockFetch.mockResolvedValue(createMockResponse(validationErrorResponse));

      const result = await client.submitBusinessInquiry(mockSubmissionData);

      expect(result.success).toBe(false);
      expect(result.error).toBe('Validation failed');
      expect(result.validation_errors).toContain('Organization name is required');
    });

    it('should handle rate limiting responses', async () => {
      const rateLimitResponse = {
        error: 'Rate limit exceeded',
        correlation_id: 'corr-rate-limit',
        success: false,
        message: 'Too many requests. Please try again later.',
        retry_after: 60
      };

      mockFetch.mockResolvedValue(createMockResponse(rateLimitResponse));

      const result = await client.submitBusinessInquiry(mockSubmissionData);

      expect(result.success).toBe(false);
      expect(result.error).toBe('Rate limit exceeded');
      expect(result.retry_after).toBe(60);
    });
  });

  describe('getBusinessInquiry', () => {
    it('should retrieve business inquiry by ID', async () => {
      const getResponse = {
        business_inquiry: mockBusinessInquiry,
        correlation_id: 'corr-get-123'
      };

      mockFetch.mockResolvedValue(createMockResponse(getResponse));

      const result = await client.getBusinessInquiry('123e4567-e89b-12d3-a456-426614174000');

      expect(mockFetch).toHaveBeenCalledWith('http://localhost:7220/api/inquiries/business/123e4567-e89b-12d3-a456-426614174000', expect.objectContaining({
        method: 'GET',
        headers: expect.objectContaining({
          'Content-Type': 'application/json',
          'Accept': 'application/json',
          'X-Retry-Attempt': '1'
        }),
        signal: expect.any(AbortSignal)
      }));
      expect(result).toEqual(getResponse);
      expect(result.business_inquiry?.inquiry_id).toBe('123e4567-e89b-12d3-a456-426614174000');
    });

    it('should handle inquiry not found', async () => {
      const notFoundResponse = {
        error: 'Inquiry not found',
        correlation_id: 'corr-not-found',
        success: false
      };

      mockFetch.mockResolvedValue(createMockResponse(notFoundResponse));

      const result = await client.getBusinessInquiry('non-existent-id');

      expect(result.error).toBe('Inquiry not found');
      expect(result.success).toBe(false);
    });
  });

  describe('error handling', () => {
    it('should handle network errors appropriately', async () => {
      const networkError = new Error('Network connection failed');
      mockFetch.mockRejectedValue(networkError);

      await expect(client.submitBusinessInquiry(mockSubmissionData))
        .rejects.toThrow('Network connection failed');
    });

    it('should handle timeout errors', async () => {
      const timeoutError = new Error('Request timeout');
      mockFetch.mockRejectedValue(timeoutError);

      await expect(client.submitBusinessInquiry(mockSubmissionData))
        .rejects.toThrow('Request timeout');
    });

    it('should handle malformed responses', async () => {
      mockFetch.mockResolvedValue(null);

      await expect(client.submitBusinessInquiry(mockSubmissionData))
        .rejects.toThrow();
    });
  });

  describe('request formatting', () => {
    it('should properly format business inquiry submission data', async () => {
      mockFetch.mockResolvedValue(createMockResponse(mockSubmissionResponse));

      await client.submitBusinessInquiry(mockSubmissionData);

      const calledWith = JSON.parse(mockFetch.mock.calls[0][1].body);
      expect(calledWith).toMatchObject({
        contact_name: 'John Smith',
        email: 'john.smith@company.com',
        organization_name: 'Acme Corporation',
        title: 'Director of Partnerships',
        inquiry_type: 'partnership',
        message: expect.stringContaining('partnership opportunities')
      });
    });

    it('should include optional fields when provided', async () => {
      const submissionWithOptionals = {
        ...mockSubmissionData,
        phone: '+1-555-123-4567',
        industry: 'Healthcare'
      };

      mockFetch.mockResolvedValue(createMockResponse(mockSubmissionResponse));

      await client.submitBusinessInquiry(submissionWithOptionals);

      const calledWith = JSON.parse(mockFetch.mock.calls[0][1].body);
      expect(calledWith.phone).toBe('+1-555-123-4567');
      expect(calledWith.industry).toBe('Healthcare');
    });

    it('should not include undefined optional fields', async () => {
      const submissionWithUndefined = {
        ...mockSubmissionData,
        phone: undefined,
        industry: undefined
      };

      mockFetch.mockResolvedValue(createMockResponse(mockSubmissionResponse));

      await client.submitBusinessInquiry(submissionWithUndefined);

      const calledWith = JSON.parse(mockFetch.mock.calls[0][1].body);
      expect(calledWith.phone).toBeUndefined();
      expect(calledWith.industry).toBeUndefined();
    });
  });

  describe('response handling', () => {
    it('should properly parse successful submission response', async () => {
      mockFetch.mockResolvedValue(createMockResponse(mockSubmissionResponse));

      const result = await client.submitBusinessInquiry(mockSubmissionData);

      expect(result.success).toBe(true);
      expect(result.message).toBe('Business inquiry submitted successfully');
      expect(result.correlation_id).toBe('corr-123-456-789');
      expect(result.business_inquiry).toMatchObject(mockBusinessInquiry);
    });

    it('should handle responses with correlation IDs', async () => {
      mockFetch.mockResolvedValue(createMockResponse(mockSubmissionResponse));

      const result = await client.submitBusinessInquiry(mockSubmissionData);

      expect(result.correlation_id).toBeTruthy();
      expect(typeof result.correlation_id).toBe('string');
    });
  });

  describe('domain-specific business logic', () => {
    it('should handle partnership inquiry types correctly', async () => {
      const partnershipInquiry = {
        ...mockSubmissionData,
        inquiry_type: 'partnership' as const,
        message: 'We are seeking strategic partnerships for medical device development.'
      };

      mockFetch.mockResolvedValue(createMockResponse({
        ...mockSubmissionResponse,
        business_inquiry: {
          ...mockBusinessInquiry,
          inquiry_type: 'partnership',
          message: partnershipInquiry.message
        }
      }));

      const result = await client.submitBusinessInquiry(partnershipInquiry);

      expect(result.business_inquiry?.inquiry_type).toBe('partnership');
      expect(result.business_inquiry?.message).toContain('strategic partnerships');
    });

    it('should handle licensing inquiry types correctly', async () => {
      const licensingInquiry = {
        ...mockSubmissionData,
        inquiry_type: 'licensing' as const,
        message: 'We are interested in licensing your patented medical technology.'
      };

      mockFetch.mockResolvedValue(createMockResponse({
        ...mockSubmissionResponse,
        business_inquiry: {
          ...mockBusinessInquiry,
          inquiry_type: 'licensing',
          message: licensingInquiry.message
        }
      }));

      const result = await client.submitBusinessInquiry(licensingInquiry);

      expect(result.business_inquiry?.inquiry_type).toBe('licensing');
      expect(result.business_inquiry?.message).toContain('licensing');
    });

    it('should handle technology inquiry types correctly', async () => {
      const technologyInquiry = {
        ...mockSubmissionData,
        inquiry_type: 'technology' as const,
        message: 'We have innovative technology that could benefit your research.'
      };

      mockFetch.mockResolvedValue(createMockResponse({
        ...mockSubmissionResponse,
        business_inquiry: {
          ...mockBusinessInquiry,
          inquiry_type: 'technology',
          message: technologyInquiry.message
        }
      }));

      const result = await client.submitBusinessInquiry(technologyInquiry);

      expect(result.business_inquiry?.inquiry_type).toBe('technology');
      expect(result.business_inquiry?.message).toContain('innovative technology');
    });
  });
});