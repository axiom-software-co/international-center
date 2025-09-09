import { describe, it, expect, vi, beforeEach } from 'vitest';
import { ref, nextTick } from 'vue';
import { useBusinessInquiry, useBusinessInquirySubmission } from './useBusinessInquiry';
import { BusinessInquiryRestClient } from '../rest/BusinessInquiryRestClient';
import { RestError } from '../rest/BaseRestClient';
import type { BusinessInquirySubmission, InquirySubmissionResponse } from '../inquiries/types';

// Mock the BusinessInquiryRestClient with hoisted functions
const {
  mockSubmitBusinessInquiry,
  mockGetBusinessInquiry,
  MockedBusinessInquiryRestClient
} = vi.hoisted(() => {
  const mockSubmit = vi.fn();
  const mockGet = vi.fn();
  
  return {
    mockSubmitBusinessInquiry: mockSubmit,
    mockGetBusinessInquiry: mockGet,
    MockedBusinessInquiryRestClient: vi.fn().mockImplementation(() => ({
      submitBusinessInquiry: mockSubmit,
      getBusinessInquiry: mockGet,
    }))
  };
});

vi.mock('../rest/BusinessInquiryRestClient', () => ({
  BusinessInquiryRestClient: MockedBusinessInquiryRestClient
}));

describe('useBusinessInquiry composables', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  describe('useBusinessInquirySubmission', () => {
    const mockSubmissionData: BusinessInquirySubmission = {
      contact_name: 'John Smith',
      email: 'john.smith@company.com',
      phone: '+1-555-123-4567',
      organization_name: 'Acme Corporation',
      title: 'Director of Partnerships',
      industry: 'Technology',
      inquiry_type: 'partnership',
      message: 'We are interested in exploring partnership opportunities with your organization to advance medical research and technology development.'
    };

    const mockSuccessResponse: InquirySubmissionResponse = {
      business_inquiry: {
        inquiry_id: '123e4567-e89b-12d3-a456-426614174000',
        contact_name: 'John Smith',
        email: 'john.smith@company.com',
        phone: '+1-555-123-4567',
        organization_name: 'Acme Corporation',
        title: 'Director of Partnerships',
        industry: 'Technology',
        inquiry_type: 'partnership',
        message: mockSubmissionData.message,
        status: 'new',
        priority: 'medium',
        source: 'website',
        created_at: '2024-03-15T10:00:00Z',
        updated_at: '2024-03-15T10:00:00Z',
        created_by: 'system',
        updated_by: 'system',
        is_deleted: false
      },
      correlation_id: 'corr-123-456-789',
      success: true,
      message: 'Business inquiry submitted successfully'
    };

    it('should initialize with correct default values', () => {
      const { isSubmitting, error, response, isSuccess, isError } = useBusinessInquirySubmission();

      expect(isSubmitting.value).toBe(false);
      expect(error.value).toBe(null);
      expect(response.value).toBe(null);
      expect(isSuccess.value).toBe(false);
      expect(isError.value).toBe(false);
    });

    it('should submit business inquiry successfully', async () => {
      mockSubmitBusinessInquiry.mockResolvedValueOnce(mockSuccessResponse);

      const { submitInquiry, isSubmitting, response, isSuccess, error } = useBusinessInquirySubmission();

      expect(isSubmitting.value).toBe(false);

      const submissionPromise = submitInquiry(mockSubmissionData);
      
      // Should be submitting during the request
      expect(isSubmitting.value).toBe(true);

      await submissionPromise;
      await nextTick();

      // Should complete successfully
      expect(isSubmitting.value).toBe(false);
      expect(isSuccess.value).toBe(true);
      expect(error.value).toBe(null);
      expect(response.value).toEqual(mockSuccessResponse);
      expect(response.value?.business_inquiry?.inquiry_type).toBe('partnership');
      expect(mockSubmitBusinessInquiry).toHaveBeenCalledWith(mockSubmissionData);
    });

    it('should handle different inquiry types correctly', async () => {
      const researchInquiry = {
        ...mockSubmissionData,
        inquiry_type: 'research' as const,
        message: 'We would like to collaborate on medical research projects and clinical trials.'
      };

      const researchResponse = {
        ...mockSuccessResponse,
        business_inquiry: {
          ...mockSuccessResponse.business_inquiry!,
          inquiry_type: 'research',
          message: researchInquiry.message
        }
      };

      mockSubmitBusinessInquiry.mockResolvedValueOnce(researchResponse);

      const { submitInquiry, response, isSuccess } = useBusinessInquirySubmission();

      await submitInquiry(researchInquiry);
      await nextTick();

      expect(isSuccess.value).toBe(true);
      expect(response.value?.business_inquiry?.inquiry_type).toBe('research');
      expect(response.value?.business_inquiry?.message).toContain('medical research');
    });

    it('should handle licensing inquiries correctly', async () => {
      const licensingInquiry = {
        ...mockSubmissionData,
        inquiry_type: 'licensing' as const,
        message: 'We are interested in licensing your patented medical technology for commercial applications.'
      };

      const licensingResponse = {
        ...mockSuccessResponse,
        business_inquiry: {
          ...mockSuccessResponse.business_inquiry!,
          inquiry_type: 'licensing',
          message: licensingInquiry.message
        }
      };

      mockSubmitBusinessInquiry.mockResolvedValueOnce(licensingResponse);

      const { submitInquiry, response, isSuccess } = useBusinessInquirySubmission();

      await submitInquiry(licensingInquiry);
      await nextTick();

      expect(isSuccess.value).toBe(true);
      expect(response.value?.business_inquiry?.inquiry_type).toBe('licensing');
      expect(response.value?.business_inquiry?.message).toContain('licensing');
    });

    it('should handle submission errors', async () => {
      const errorMessage = 'Network connection failed';
      mockSubmitBusinessInquiry.mockRejectedValueOnce(new Error(errorMessage));

      const { submitInquiry, isSubmitting, error, isError, isSuccess } = useBusinessInquirySubmission();

      const submissionPromise = submitInquiry(mockSubmissionData);
      
      expect(isSubmitting.value).toBe(true);

      await submissionPromise;
      await nextTick();

      expect(isSubmitting.value).toBe(false);
      expect(isError.value).toBe(true);
      expect(isSuccess.value).toBe(false);
      expect(error.value).toBe(errorMessage);
    });

    it('should handle validation errors from backend', async () => {
      const validationErrorResponse = {
        error: 'Validation failed',
        validation_errors: [
          'Organization name is required',
          'Message must be at least 20 characters'
        ],
        correlation_id: 'corr-validation-error',
        success: false,
        message: 'Please correct the following errors'
      };

      mockSubmitBusinessInquiry.mockResolvedValueOnce(validationErrorResponse);

      const { submitInquiry, error, isError, response, isSuccess } = useBusinessInquirySubmission();

      await submitInquiry(mockSubmissionData);
      await nextTick();

      expect(isError.value).toBe(true);
      expect(isSuccess.value).toBe(false);
      expect(error.value).toBe('Validation failed');
      expect(response.value).toEqual(validationErrorResponse);
    });

    it('should handle rate limiting responses', async () => {
      const rateLimitResponse = {
        error: 'Rate limit exceeded',
        correlation_id: 'corr-rate-limit',
        success: false,
        message: 'Too many requests. Please try again later.',
        retry_after: 60
      };

      mockSubmitBusinessInquiry.mockResolvedValueOnce(rateLimitResponse);

      const { submitInquiry, error, isError, response } = useBusinessInquirySubmission();

      await submitInquiry(mockSubmissionData);
      await nextTick();

      expect(isError.value).toBe(true);
      expect(error.value).toBe('Rate limit exceeded');
      expect(response.value?.retry_after).toBe(60);
    });

    it('should reset state when submitting new inquiry', async () => {
      mockSubmitBusinessInquiry.mockResolvedValueOnce(mockSuccessResponse);

      const { submitInquiry, error, response, reset } = useBusinessInquirySubmission();

      // First submission
      await submitInquiry(mockSubmissionData);
      await nextTick();

      expect(response.value).toEqual(mockSuccessResponse);

      // Reset state
      reset();

      expect(error.value).toBe(null);
      expect(response.value).toBe(null);
    });

    it('should handle submission with optional fields', async () => {
      const submissionWithOptionalFields = {
        contact_name: 'Jane Doe',
        email: 'jane.doe@biotech.com',
        organization_name: 'BioTech Solutions',
        title: 'CEO',
        inquiry_type: 'technology' as const,
        message: 'We have developed innovative biotechnology that could benefit your research programs.'
      };

      const responseWithOptionals = {
        ...mockSuccessResponse,
        business_inquiry: {
          ...mockSuccessResponse.business_inquiry!,
          contact_name: 'Jane Doe',
          email: 'jane.doe@biotech.com',
          organization_name: 'BioTech Solutions',
          title: 'CEO',
          inquiry_type: 'technology',
          message: submissionWithOptionalFields.message,
          phone: undefined,
          industry: undefined
        }
      };

      mockSubmitBusinessInquiry.mockResolvedValueOnce(responseWithOptionals);

      const { submitInquiry, response, isSuccess } = useBusinessInquirySubmission();

      await submitInquiry(submissionWithOptionalFields);
      await nextTick();

      expect(isSuccess.value).toBe(true);
      expect(response.value?.business_inquiry?.inquiry_type).toBe('technology');
      expect(response.value?.business_inquiry?.organization_name).toBe('BioTech Solutions');
    });
  });

  describe('useBusinessInquiry', () => {
    const mockBusinessInquiry = {
      inquiry_id: '123e4567-e89b-12d3-a456-426614174000',
      contact_name: 'John Smith',
      email: 'john.smith@company.com',
      organization_name: 'Acme Corporation',
      title: 'Director of Partnerships',
      inquiry_type: 'partnership' as const,
      message: 'We are interested in exploring partnership opportunities.',
      status: 'new' as const,
      priority: 'medium' as const,
      source: 'website',
      created_at: '2024-03-15T10:00:00Z',
      updated_at: '2024-03-15T10:00:00Z',
      created_by: 'system',
      updated_by: 'system',
      is_deleted: false
    };

    it('should initialize with correct default values when no ID provided', () => {
      const { inquiry, loading, error } = useBusinessInquiry(ref(null));

      expect(inquiry.value).toBe(null);
      expect(loading.value).toBe(false);
      expect(error.value).toBe(null);
    });

    it('should fetch business inquiry by ID', async () => {
      const mockResponse = {
        business_inquiry: mockBusinessInquiry,
        correlation_id: 'corr-get-123'
      };

      mockGetBusinessInquiry.mockResolvedValueOnce(mockResponse);

      const inquiryId = ref('123e4567-e89b-12d3-a456-426614174000');
      const { inquiry, loading, error } = useBusinessInquiry(inquiryId);

      await nextTick();

      expect(inquiry.value).toEqual(mockBusinessInquiry);
      expect(loading.value).toBe(false);
      expect(error.value).toBe(null);
      expect(mockGetBusinessInquiry).toHaveBeenCalledWith('123e4567-e89b-12d3-a456-426614174000');
    });

    it('should handle inquiry not found', async () => {
      const notFoundResponse = {
        error: 'Business inquiry not found',
        correlation_id: 'corr-not-found',
        success: false
      };

      mockGetBusinessInquiry.mockResolvedValueOnce(notFoundResponse);

      const inquiryId = ref('non-existent-id');
      const { inquiry, loading, error } = useBusinessInquiry(inquiryId);

      await nextTick();

      expect(inquiry.value).toBe(null);
      expect(loading.value).toBe(false);
      expect(error.value).toBe('Business inquiry not found');
    });

    it('should handle fetch errors', async () => {
      const errorMessage = 'Failed to fetch business inquiry';
      mockGetBusinessInquiry.mockRejectedValueOnce(new Error(errorMessage));

      const inquiryId = ref('123e4567-e89b-12d3-a456-426614174000');
      const { inquiry, error } = useBusinessInquiry(inquiryId);

      await nextTick();

      expect(inquiry.value).toBe(null);
      expect(error.value).toBe(errorMessage);
    });

    it('should refetch when inquiry ID changes', async () => {
      const mockResponse1 = {
        business_inquiry: { ...mockBusinessInquiry, inquiry_id: 'id-1' },
        correlation_id: 'corr-1'
      };

      const mockResponse2 = {
        business_inquiry: { ...mockBusinessInquiry, inquiry_id: 'id-2', organization_name: 'New Company' },
        correlation_id: 'corr-2'
      };

      mockGetBusinessInquiry.mockResolvedValueOnce(mockResponse1);

      const inquiryId = ref('id-1');
      const { inquiry } = useBusinessInquiry(inquiryId);

      await nextTick();
      expect(inquiry.value?.inquiry_id).toBe('id-1');

      // Change ID and mock second response
      mockGetBusinessInquiry.mockResolvedValueOnce(mockResponse2);
      inquiryId.value = 'id-2';

      await nextTick();
      expect(inquiry.value?.inquiry_id).toBe('id-2');
      expect(inquiry.value?.organization_name).toBe('New Company');
      expect(mockGetBusinessInquiry).toHaveBeenCalledTimes(2);
    });

    it('should handle loading states correctly', async () => {
      const mockResponse = {
        business_inquiry: mockBusinessInquiry,
        correlation_id: 'corr-loading-test'
      };

      let resolvePromise: (value: any) => void;
      const pendingPromise = new Promise(resolve => {
        resolvePromise = resolve;
      });
      mockGetBusinessInquiry.mockReturnValueOnce(pendingPromise);

      const inquiryId = ref('123e4567-e89b-12d3-a456-426614174000');
      const { inquiry, loading, error } = useBusinessInquiry(inquiryId);

      await nextTick();

      // Should be loading
      expect(loading.value).toBe(true);
      expect(inquiry.value).toBe(null);
      expect(error.value).toBe(null);

      // Resolve the promise
      resolvePromise!(mockResponse);
      await nextTick();

      // Should complete loading
      expect(loading.value).toBe(false);
      expect(inquiry.value).toEqual(mockBusinessInquiry);
      expect(error.value).toBe(null);
    });
  });

  describe('error handling edge cases', () => {
    it('should handle network timeouts', async () => {
      const timeoutError = new Error('Request timeout');
      mockSubmitBusinessInquiry.mockRejectedValueOnce(timeoutError);

      const { submitInquiry, error, isError } = useBusinessInquirySubmission();

      const mockSubmissionData: BusinessInquirySubmission = {
        contact_name: 'Test User',
        email: 'test@example.com',
        organization_name: 'Test Corp',
        title: 'Manager',
        inquiry_type: 'other',
        message: 'This is a test message for timeout scenario.'
      };

      await submitInquiry(mockSubmissionData);
      await nextTick();

      expect(isError.value).toBe(true);
      expect(error.value).toBe('Request timeout');
    });

    it('should handle malformed responses gracefully', async () => {
      mockSubmitBusinessInquiry.mockResolvedValueOnce(null);

      const { submitInquiry, error, isError } = useBusinessInquirySubmission();

      const mockSubmissionData: BusinessInquirySubmission = {
        contact_name: 'Test User',
        email: 'test@example.com',
        organization_name: 'Test Corp',
        title: 'Manager',
        inquiry_type: 'other',
        message: 'This is a test message for malformed response.'
      };

      await submitInquiry(mockSubmissionData);
      await nextTick();

      expect(isError.value).toBe(true);
      expect(error.value).toContain('response');
    });
  });
});