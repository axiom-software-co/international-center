import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { BusinessInquiryRestClient } from './BusinessInquiryRestClient';
import type { BusinessInquiry, BusinessInquirySubmission, InquirySubmissionResponse } from '../inquiries/types';

// Mock the BaseRestClient
vi.mock('./BaseRestClient');

describe('BusinessInquiryRestClient', () => {
  let client: BusinessInquiryRestClient;
  let mockPost: ReturnType<typeof vi.fn>;
  let mockGet: ReturnType<typeof vi.fn>;

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
    mockPost = vi.fn();
    mockGet = vi.fn();
    
    client = new BusinessInquiryRestClient();
    // Mock the inherited methods from BaseRestClient
    (client as any).post = mockPost;
    (client as any).get = mockGet;
  });

  afterEach(() => {
    vi.clearAllMocks();
  });

  describe('submitBusinessInquiry', () => {
    it('should submit business inquiry with valid data', async () => {
      mockPost.mockResolvedValue(mockSubmissionResponse);

      const result = await client.submitBusinessInquiry(mockSubmissionData);

      expect(mockPost).toHaveBeenCalledWith('/api/inquiries/business', mockSubmissionData);
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

      mockPost.mockResolvedValue(responseWithOptionals);

      const result = await client.submitBusinessInquiry(submissionWithOptionalFields);

      expect(mockPost).toHaveBeenCalledWith('/api/inquiries/business', submissionWithOptionalFields);
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

      mockPost.mockResolvedValue(researchResponse);

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

      mockPost.mockRejectedValue(new Error('Network error'));

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

      mockPost.mockResolvedValue(validationErrorResponse);

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

      mockPost.mockResolvedValue(rateLimitResponse);

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

      mockGet.mockResolvedValue(getResponse);

      const result = await client.getBusinessInquiry('123e4567-e89b-12d3-a456-426614174000');

      expect(mockGet).toHaveBeenCalledWith('/api/inquiries/business/123e4567-e89b-12d3-a456-426614174000');
      expect(result).toEqual(getResponse);
      expect(result.business_inquiry?.inquiry_id).toBe('123e4567-e89b-12d3-a456-426614174000');
    });

    it('should handle inquiry not found', async () => {
      const notFoundResponse = {
        error: 'Inquiry not found',
        correlation_id: 'corr-not-found',
        success: false
      };

      mockGet.mockResolvedValue(notFoundResponse);

      const result = await client.getBusinessInquiry('non-existent-id');

      expect(result.error).toBe('Inquiry not found');
      expect(result.success).toBe(false);
    });
  });

  describe('error handling', () => {
    it('should handle network errors appropriately', async () => {
      const networkError = new Error('Network connection failed');
      mockPost.mockRejectedValue(networkError);

      await expect(client.submitBusinessInquiry(mockSubmissionData))
        .rejects.toThrow('Network connection failed');
    });

    it('should handle timeout errors', async () => {
      const timeoutError = new Error('Request timeout');
      mockPost.mockRejectedValue(timeoutError);

      await expect(client.submitBusinessInquiry(mockSubmissionData))
        .rejects.toThrow('Request timeout');
    });

    it('should handle malformed responses', async () => {
      mockPost.mockResolvedValue(null);

      await expect(client.submitBusinessInquiry(mockSubmissionData))
        .rejects.toThrow();
    });
  });

  describe('request formatting', () => {
    it('should properly format business inquiry submission data', async () => {
      mockPost.mockResolvedValue(mockSubmissionResponse);

      await client.submitBusinessInquiry(mockSubmissionData);

      const calledWith = mockPost.mock.calls[0][1];
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

      mockPost.mockResolvedValue(mockSubmissionResponse);

      await client.submitBusinessInquiry(submissionWithOptionals);

      const calledWith = mockPost.mock.calls[0][1];
      expect(calledWith.phone).toBe('+1-555-123-4567');
      expect(calledWith.industry).toBe('Healthcare');
    });

    it('should not include undefined optional fields', async () => {
      const submissionWithUndefined = {
        ...mockSubmissionData,
        phone: undefined,
        industry: undefined
      };

      mockPost.mockResolvedValue(mockSubmissionResponse);

      await client.submitBusinessInquiry(submissionWithUndefined);

      const calledWith = mockPost.mock.calls[0][1];
      expect(calledWith.phone).toBeUndefined();
      expect(calledWith.industry).toBeUndefined();
    });
  });

  describe('response handling', () => {
    it('should properly parse successful submission response', async () => {
      mockPost.mockResolvedValue(mockSubmissionResponse);

      const result = await client.submitBusinessInquiry(mockSubmissionData);

      expect(result.success).toBe(true);
      expect(result.message).toBe('Business inquiry submitted successfully');
      expect(result.correlation_id).toBe('corr-123-456-789');
      expect(result.business_inquiry).toMatchObject(mockBusinessInquiry);
    });

    it('should handle responses with correlation IDs', async () => {
      mockPost.mockResolvedValue(mockSubmissionResponse);

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

      mockPost.mockResolvedValue({
        ...mockSubmissionResponse,
        business_inquiry: {
          ...mockBusinessInquiry,
          inquiry_type: 'partnership',
          message: partnershipInquiry.message
        }
      });

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

      mockPost.mockResolvedValue({
        ...mockSubmissionResponse,
        business_inquiry: {
          ...mockBusinessInquiry,
          inquiry_type: 'licensing',
          message: licensingInquiry.message
        }
      });

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

      mockPost.mockResolvedValue({
        ...mockSubmissionResponse,
        business_inquiry: {
          ...mockBusinessInquiry,
          inquiry_type: 'technology',
          message: technologyInquiry.message
        }
      });

      const result = await client.submitBusinessInquiry(technologyInquiry);

      expect(result.business_inquiry?.inquiry_type).toBe('technology');
      expect(result.business_inquiry?.message).toContain('innovative technology');
    });
  });
});