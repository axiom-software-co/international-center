import { describe, it, expect, vi, beforeEach } from 'vitest';
import { ref, nextTick } from 'vue';
import { useMediaInquiry, useMediaInquirySubmission } from './useMediaInquiry';
import { MediaInquiryRestClient } from '../rest/MediaInquiryRestClient';
import { RestError } from '../rest/BaseRestClient';
import type { MediaInquirySubmission, InquirySubmissionResponse } from '../inquiries/types';

// Mock the MediaInquiryRestClient
vi.mock('../rest/MediaInquiryRestClient', () => ({
  MediaInquiryRestClient: vi.fn().mockImplementation(() => ({
    submitMediaInquiry: vi.fn(),
    getMediaInquiry: vi.fn(),
  }))
}));

describe('useMediaInquiry composables', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  describe('useMediaInquirySubmission', () => {
    const mockStandardSubmission: MediaInquirySubmission = {
      contact_name: 'Sarah Reporter',
      email: 'sarah.reporter@newsnetwork.com',
      outlet: 'Medical News Network',
      title: 'Senior Medical Reporter',
      phone: '+1-555-987-6543',
      media_type: 'digital',
      subject: 'Request for interview regarding new treatment protocol and FDA approval process',
      urgency: 'medium'
    };

    const mockUrgentSubmission: MediaInquirySubmission = {
      contact_name: 'Tom Journalist',
      email: 'tom.journalist@tv.com',
      outlet: 'Health TV',
      title: 'Health Correspondent',
      phone: '+1-555-111-2222',
      media_type: 'television',
      deadline: '2024-03-16',
      urgency: 'high',
      subject: 'Breaking: New FDA approval for innovative treatment - urgent interview needed'
    };

    const mockSuccessResponse: InquirySubmissionResponse = {
      media_inquiry: {
        inquiry_id: '789e0123-e89b-12d3-a456-426614174002',
        contact_name: 'Sarah Reporter',
        email: 'sarah.reporter@newsnetwork.com',
        outlet: 'Medical News Network',
        title: 'Senior Medical Reporter',
        phone: '+1-555-987-6543',
        media_type: 'digital',
        subject: mockStandardSubmission.subject,
        urgency: 'medium',
        status: 'new',
        priority: 'medium',
        source: 'website',
        created_at: '2024-03-15T10:00:00Z',
        updated_at: '2024-03-15T10:00:00Z',
        created_by: 'system',
        updated_by: 'system',
        is_deleted: false
      },
      correlation_id: 'corr-789-012-345',
      success: true,
      message: 'Media inquiry submitted successfully'
    };

    it('should initialize with correct default values', () => {
      const { isSubmitting, error, response, isSuccess, isError } = useMediaInquirySubmission();

      expect(isSubmitting.value).toBe(false);
      expect(error.value).toBe(null);
      expect(response.value).toBe(null);
      expect(isSuccess.value).toBe(false);
      expect(isError.value).toBe(false);
    });

    it('should submit media inquiry successfully', async () => {
      const mockClient = new MediaInquiryRestClient();
      (mockClient.submitMediaInquiry as any).mockResolvedValueOnce(mockSuccessResponse);

      const { submitInquiry, isSubmitting, response, isSuccess, error } = useMediaInquirySubmission();

      expect(isSubmitting.value).toBe(false);

      const submissionPromise = submitInquiry(mockStandardSubmission);
      
      // Should be submitting during the request
      expect(isSubmitting.value).toBe(true);

      await submissionPromise;
      await nextTick();

      // Should complete successfully
      expect(isSubmitting.value).toBe(false);
      expect(isSuccess.value).toBe(true);
      expect(error.value).toBe(null);
      expect(response.value).toEqual(mockSuccessResponse);
      expect(response.value?.media_inquiry?.outlet).toBe('Medical News Network');
      expect(mockClient.submitMediaInquiry).toHaveBeenCalledWith(mockStandardSubmission);
    });

    it('should handle urgent media inquiry with deadline correctly', async () => {
      const urgentResponse = {
        ...mockSuccessResponse,
        media_inquiry: {
          ...mockSuccessResponse.media_inquiry!,
          contact_name: 'Tom Journalist',
          email: 'tom.journalist@tv.com',
          outlet: 'Health TV',
          title: 'Health Correspondent',
          phone: '+1-555-111-2222',
          media_type: 'television',
          deadline: '2024-03-16',
          urgency: 'high',
          subject: mockUrgentSubmission.subject
        }
      };

      const mockClient = new MediaInquiryRestClient();
      (mockClient.submitMediaInquiry as any).mockResolvedValueOnce(urgentResponse);

      const { submitInquiry, response, isSuccess } = useMediaInquirySubmission();

      await submitInquiry(mockUrgentSubmission);
      await nextTick();

      expect(isSuccess.value).toBe(true);
      expect(response.value?.media_inquiry?.urgency).toBe('high');
      expect(response.value?.media_inquiry?.deadline).toBe('2024-03-16');
      expect(response.value?.media_inquiry?.media_type).toBe('television');
    });

    it('should handle different media types correctly', async () => {
      const printInquiry = {
        ...mockStandardSubmission,
        media_type: 'print' as const,
        outlet: 'Medical Journal',
        subject: 'Feature article on breakthrough medical research findings'
      };

      const printResponse = {
        ...mockSuccessResponse,
        media_inquiry: {
          ...mockSuccessResponse.media_inquiry!,
          media_type: 'print',
          outlet: 'Medical Journal',
          subject: printInquiry.subject
        }
      };

      const mockClient = new MediaInquiryRestClient();
      (mockClient.submitMediaInquiry as any).mockResolvedValueOnce(printResponse);

      const { submitInquiry, response, isSuccess } = useMediaInquirySubmission();

      await submitInquiry(printInquiry);
      await nextTick();

      expect(isSuccess.value).toBe(true);
      expect(response.value?.media_inquiry?.media_type).toBe('print');
      expect(response.value?.media_inquiry?.outlet).toBe('Medical Journal');
    });

    it('should handle podcast media type correctly', async () => {
      const podcastInquiry = {
        ...mockStandardSubmission,
        media_type: 'podcast' as const,
        outlet: 'Healthcare Podcast',
        subject: 'Podcast interview about patient care innovations and treatment outcomes'
      };

      const podcastResponse = {
        ...mockSuccessResponse,
        media_inquiry: {
          ...mockSuccessResponse.media_inquiry!,
          media_type: 'podcast',
          outlet: 'Healthcare Podcast',
          subject: podcastInquiry.subject
        }
      };

      const mockClient = new MediaInquiryRestClient();
      (mockClient.submitMediaInquiry as any).mockResolvedValueOnce(podcastResponse);

      const { submitInquiry, response, isSuccess } = useMediaInquirySubmission();

      await submitInquiry(podcastInquiry);
      await nextTick();

      expect(isSuccess.value).toBe(true);
      expect(response.value?.media_inquiry?.media_type).toBe('podcast');
      expect(response.value?.media_inquiry?.outlet).toBe('Healthcare Podcast');
    });

    it('should handle radio media type correctly', async () => {
      const radioInquiry = {
        ...mockStandardSubmission,
        media_type: 'radio' as const,
        outlet: 'Health Radio Network',
        subject: 'Radio interview about patient care improvements and treatment accessibility'
      };

      const radioResponse = {
        ...mockSuccessResponse,
        media_inquiry: {
          ...mockSuccessResponse.media_inquiry!,
          media_type: 'radio',
          outlet: 'Health Radio Network',
          subject: radioInquiry.subject
        }
      };

      const mockClient = new MediaInquiryRestClient();
      (mockClient.submitMediaInquiry as any).mockResolvedValueOnce(radioResponse);

      const { submitInquiry, response, isSuccess } = useMediaInquirySubmission();

      await submitInquiry(radioInquiry);
      await nextTick();

      expect(isSuccess.value).toBe(true);
      expect(response.value?.media_inquiry?.media_type).toBe('radio');
    });

    it('should handle different urgency levels correctly', async () => {
      const lowUrgencyInquiry = {
        ...mockStandardSubmission,
        urgency: 'low' as const,
        subject: 'General inquiry about future research developments - no rush'
      };

      const lowUrgencyResponse = {
        ...mockSuccessResponse,
        media_inquiry: {
          ...mockSuccessResponse.media_inquiry!,
          urgency: 'low',
          subject: lowUrgencyInquiry.subject
        }
      };

      const mockClient = new MediaInquiryRestClient();
      (mockClient.submitMediaInquiry as any).mockResolvedValueOnce(lowUrgencyResponse);

      const { submitInquiry, response, isSuccess } = useMediaInquirySubmission();

      await submitInquiry(lowUrgencyInquiry);
      await nextTick();

      expect(isSuccess.value).toBe(true);
      expect(response.value?.media_inquiry?.urgency).toBe('low');
    });

    it('should handle submission errors', async () => {
      const errorMessage = 'Network connection failed';
      const mockClient = new MediaInquiryRestClient();
      (mockClient.submitMediaInquiry as any).mockRejectedValueOnce(new Error(errorMessage));

      const { submitInquiry, isSubmitting, error, isError, isSuccess } = useMediaInquirySubmission();

      const submissionPromise = submitInquiry(mockStandardSubmission);
      
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
          'Phone number is required for media inquiries',
          'Subject must be at least 20 characters',
          'Outlet is required'
        ],
        correlation_id: 'corr-validation-error',
        success: false,
        message: 'Please correct the following errors'
      };

      const mockClient = new MediaInquiryRestClient();
      (mockClient.submitMediaInquiry as any).mockResolvedValueOnce(validationErrorResponse);

      const { submitInquiry, error, isError, response, isSuccess } = useMediaInquirySubmission();

      await submitInquiry(mockStandardSubmission);
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
        retry_after: 180
      };

      const mockClient = new MediaInquiryRestClient();
      (mockClient.submitMediaInquiry as any).mockResolvedValueOnce(rateLimitResponse);

      const { submitInquiry, error, isError, response } = useMediaInquirySubmission();

      await submitInquiry(mockStandardSubmission);
      await nextTick();

      expect(isError.value).toBe(true);
      expect(error.value).toBe('Rate limit exceeded');
      expect(response.value?.retry_after).toBe(180);
    });

    it('should reset state correctly', async () => {
      const mockClient = new MediaInquiryRestClient();
      (mockClient.submitMediaInquiry as any).mockResolvedValueOnce(mockSuccessResponse);

      const { submitInquiry, error, response, reset } = useMediaInquirySubmission();

      // First submission
      await submitInquiry(mockStandardSubmission);
      await nextTick();

      expect(response.value).toEqual(mockSuccessResponse);

      // Reset state
      reset();

      expect(error.value).toBe(null);
      expect(response.value).toBe(null);
    });

    it('should handle minimal submission without optional fields', async () => {
      const minimalSubmission = {
        contact_name: 'Basic Reporter',
        email: 'basic@news.com',
        outlet: 'Local News',
        title: 'Reporter',
        phone: '+1-555-999-8888',
        subject: 'General inquiry about your medical research programs and patient services',
        urgency: 'medium' as const
      };

      const minimalResponse = {
        ...mockSuccessResponse,
        media_inquiry: {
          ...mockSuccessResponse.media_inquiry!,
          contact_name: 'Basic Reporter',
          email: 'basic@news.com',
          outlet: 'Local News',
          title: 'Reporter',
          phone: '+1-555-999-8888',
          subject: minimalSubmission.subject,
          media_type: undefined,
          deadline: undefined
        }
      };

      const mockClient = new MediaInquiryRestClient();
      (mockClient.submitMediaInquiry as any).mockResolvedValueOnce(minimalResponse);

      const { submitInquiry, response, isSuccess } = useMediaInquirySubmission();

      await submitInquiry(minimalSubmission);
      await nextTick();

      expect(isSuccess.value).toBe(true);
      expect(response.value?.media_inquiry?.outlet).toBe('Local News');
    });

    it('should handle medical journal media type correctly', async () => {
      const journalInquiry = {
        ...mockStandardSubmission,
        media_type: 'medical-journal' as const,
        outlet: 'Journal of Medical Innovation',
        subject: 'Research publication opportunity for groundbreaking clinical study'
      };

      const journalResponse = {
        ...mockSuccessResponse,
        media_inquiry: {
          ...mockSuccessResponse.media_inquiry!,
          media_type: 'medical-journal',
          outlet: 'Journal of Medical Innovation',
          subject: journalInquiry.subject
        }
      };

      const mockClient = new MediaInquiryRestClient();
      (mockClient.submitMediaInquiry as any).mockResolvedValueOnce(journalResponse);

      const { submitInquiry, response, isSuccess } = useMediaInquirySubmission();

      await submitInquiry(journalInquiry);
      await nextTick();

      expect(isSuccess.value).toBe(true);
      expect(response.value?.media_inquiry?.media_type).toBe('medical-journal');
      expect(response.value?.media_inquiry?.outlet).toBe('Journal of Medical Innovation');
    });
  });

  describe('useMediaInquiry', () => {
    const mockMediaInquiry = {
      inquiry_id: '789e0123-e89b-12d3-a456-426614174002',
      contact_name: 'Sarah Reporter',
      email: 'sarah.reporter@newsnetwork.com',
      outlet: 'Medical News Network',
      title: 'Senior Medical Reporter',
      phone: '+1-555-987-6543',
      media_type: 'digital' as const,
      subject: 'Request for interview regarding new treatment protocol',
      urgency: 'medium' as const,
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
      const { inquiry, loading, error } = useMediaInquiry(ref(null));

      expect(inquiry.value).toBe(null);
      expect(loading.value).toBe(false);
      expect(error.value).toBe(null);
    });

    it('should fetch media inquiry by ID', async () => {
      const mockResponse = {
        media_inquiry: mockMediaInquiry,
        correlation_id: 'corr-get-789'
      };

      const mockClient = new MediaInquiryRestClient();
      (mockClient.getMediaInquiry as any).mockResolvedValueOnce(mockResponse);

      const inquiryId = ref('789e0123-e89b-12d3-a456-426614174002');
      const { inquiry, loading, error } = useMediaInquiry(inquiryId);

      await nextTick();

      expect(inquiry.value).toEqual(mockMediaInquiry);
      expect(loading.value).toBe(false);
      expect(error.value).toBe(null);
      expect(mockClient.getMediaInquiry).toHaveBeenCalledWith('789e0123-e89b-12d3-a456-426614174002');
    });

    it('should handle inquiry not found', async () => {
      const notFoundResponse = {
        error: 'Media inquiry not found',
        correlation_id: 'corr-not-found',
        success: false
      };

      const mockClient = new MediaInquiryRestClient();
      (mockClient.getMediaInquiry as any).mockResolvedValueOnce(notFoundResponse);

      const inquiryId = ref('non-existent-id');
      const { inquiry, loading, error } = useMediaInquiry(inquiryId);

      await nextTick();

      expect(inquiry.value).toBe(null);
      expect(loading.value).toBe(false);
      expect(error.value).toBe('Media inquiry not found');
    });

    it('should handle fetch errors', async () => {
      const errorMessage = 'Failed to fetch media inquiry';
      const mockClient = new MediaInquiryRestClient();
      (mockClient.getMediaInquiry as any).mockRejectedValueOnce(new Error(errorMessage));

      const inquiryId = ref('789e0123-e89b-12d3-a456-426614174002');
      const { inquiry, error } = useMediaInquiry(inquiryId);

      await nextTick();

      expect(inquiry.value).toBe(null);
      expect(error.value).toBe(errorMessage);
    });

    it('should refetch when inquiry ID changes', async () => {
      const mockResponse1 = {
        media_inquiry: { ...mockMediaInquiry, inquiry_id: 'id-1' },
        correlation_id: 'corr-1'
      };

      const mockResponse2 = {
        media_inquiry: { ...mockMediaInquiry, inquiry_id: 'id-2', outlet: 'New Outlet' },
        correlation_id: 'corr-2'
      };

      const mockClient = new MediaInquiryRestClient();
      (mockClient.getMediaInquiry as any).mockResolvedValueOnce(mockResponse1);

      const inquiryId = ref('id-1');
      const { inquiry } = useMediaInquiry(inquiryId);

      await nextTick();
      expect(inquiry.value?.inquiry_id).toBe('id-1');

      // Change ID and mock second response
      (mockClient.getMediaInquiry as any).mockResolvedValueOnce(mockResponse2);
      inquiryId.value = 'id-2';

      await nextTick();
      expect(inquiry.value?.inquiry_id).toBe('id-2');
      expect(inquiry.value?.outlet).toBe('New Outlet');
      expect(mockClient.getMediaInquiry).toHaveBeenCalledTimes(2);
    });

    it('should handle loading states correctly', async () => {
      const mockResponse = {
        media_inquiry: mockMediaInquiry,
        correlation_id: 'corr-loading-test'
      };

      const mockClient = new MediaInquiryRestClient();
      let resolvePromise: (value: any) => void;
      const pendingPromise = new Promise(resolve => {
        resolvePromise = resolve;
      });
      (mockClient.getMediaInquiry as any).mockReturnValueOnce(pendingPromise);

      const inquiryId = ref('789e0123-e89b-12d3-a456-426614174002');
      const { inquiry, loading, error } = useMediaInquiry(inquiryId);

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
      expect(inquiry.value).toEqual(mockMediaInquiry);
      expect(error.value).toBe(null);
    });
  });

  describe('error handling edge cases', () => {
    it('should handle network timeouts', async () => {
      const timeoutError = new Error('Request timeout');
      const mockClient = new MediaInquiryRestClient();
      (mockClient.submitMediaInquiry as any).mockRejectedValueOnce(timeoutError);

      const { submitInquiry, error, isError } = useMediaInquirySubmission();

      const mockSubmissionData: MediaInquirySubmission = {
        contact_name: 'Test Reporter',
        email: 'test@example.com',
        outlet: 'Test News',
        title: 'Reporter',
        phone: '+1-555-000-0000',
        subject: 'This is a test message for timeout scenario testing.',
        urgency: 'medium'
      };

      await submitInquiry(mockSubmissionData);
      await nextTick();

      expect(isError.value).toBe(true);
      expect(error.value).toBe('Request timeout');
    });

    it('should handle malformed responses gracefully', async () => {
      const mockClient = new MediaInquiryRestClient();
      (mockClient.submitMediaInquiry as any).mockResolvedValueOnce(null);

      const { submitInquiry, error, isError } = useMediaInquirySubmission();

      const mockSubmissionData: MediaInquirySubmission = {
        contact_name: 'Test Reporter',
        email: 'test@example.com',
        outlet: 'Test News',
        title: 'Reporter',
        phone: '+1-555-000-0000',
        subject: 'This is a test message for malformed response scenario.',
        urgency: 'medium'
      };

      await submitInquiry(mockSubmissionData);
      await nextTick();

      expect(isError.value).toBe(true);
      expect(error.value).toContain('response');
    });
  });

  describe('media-specific business logic', () => {
    it('should handle same-day deadline urgency correctly', async () => {
      const today = new Date().toISOString().split('T')[0];
      const sameDayInquiry = {
        ...mockUrgentSubmission,
        deadline: today,
        urgency: 'high' as const,
        subject: 'Same-day deadline: Breaking news requires immediate response'
      };

      const sameDayResponse = {
        ...mockSuccessResponse,
        media_inquiry: {
          ...mockSuccessResponse.media_inquiry!,
          deadline: today,
          urgency: 'high',
          subject: sameDayInquiry.subject
        }
      };

      const mockClient = new MediaInquiryRestClient();
      (mockClient.submitMediaInquiry as any).mockResolvedValueOnce(sameDayResponse);

      const { submitInquiry, response, isSuccess } = useMediaInquirySubmission();

      await submitInquiry(sameDayInquiry);
      await nextTick();

      expect(isSuccess.value).toBe(true);
      expect(response.value?.media_inquiry?.deadline).toBe(today);
      expect(response.value?.media_inquiry?.urgency).toBe('high');
    });

    it('should handle future deadline correctly', async () => {
      const futureDate = '2024-03-20';
      const futureDeadlineInquiry = {
        ...mockStandardSubmission,
        deadline: futureDate,
        urgency: 'low' as const,
        subject: 'Feature story with flexible timeline - advance planning'
      };

      const futureDeadlineResponse = {
        ...mockSuccessResponse,
        media_inquiry: {
          ...mockSuccessResponse.media_inquiry!,
          deadline: futureDate,
          urgency: 'low',
          subject: futureDeadlineInquiry.subject
        }
      };

      const mockClient = new MediaInquiryRestClient();
      (mockClient.submitMediaInquiry as any).mockResolvedValueOnce(futureDeadlineResponse);

      const { submitInquiry, response, isSuccess } = useMediaInquirySubmission();

      await submitInquiry(futureDeadlineInquiry);
      await nextTick();

      expect(isSuccess.value).toBe(true);
      expect(response.value?.media_inquiry?.deadline).toBe(futureDate);
      expect(response.value?.media_inquiry?.urgency).toBe('low');
    });

    it('should handle television media type with urgent deadline', async () => {
      const tvInquiry = {
        ...mockUrgentSubmission,
        media_type: 'television' as const,
        urgency: 'high' as const,
        subject: 'Live TV interview needed for breaking medical news story'
      };

      const tvResponse = {
        ...mockSuccessResponse,
        media_inquiry: {
          ...mockSuccessResponse.media_inquiry!,
          media_type: 'television',
          urgency: 'high',
          subject: tvInquiry.subject
        }
      };

      const mockClient = new MediaInquiryRestClient();
      (mockClient.submitMediaInquiry as any).mockResolvedValueOnce(tvResponse);

      const { submitInquiry, response, isSuccess } = useMediaInquirySubmission();

      await submitInquiry(tvInquiry);
      await nextTick();

      expect(isSuccess.value).toBe(true);
      expect(response.value?.media_inquiry?.media_type).toBe('television');
      expect(response.value?.media_inquiry?.urgency).toBe('high');
      expect(response.value?.media_inquiry?.subject).toContain('Live TV interview');
    });
  });
});