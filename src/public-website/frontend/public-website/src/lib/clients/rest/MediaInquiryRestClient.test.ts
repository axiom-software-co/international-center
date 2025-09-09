import { describe, it, expect, vi, beforeEach } from 'vitest';
import { MediaInquiryRestClient } from './MediaInquiryRestClient';
import { mockFetch } from '../../../test/setup';
import type { MediaInquiry, MediaInquirySubmission, InquirySubmissionResponse } from '../inquiries/types';

describe('MediaInquiryRestClient', () => {
  let client: MediaInquiryRestClient;

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

  const mockMediaInquiry: MediaInquiry = {
    inquiry_id: '789e0123-e89b-12d3-a456-426614174002',
    contact_name: 'Sarah Reporter',
    email: 'sarah.reporter@newsnetwork.com',
    outlet: 'Medical News Network',
    title: 'Senior Medical Reporter',
    phone: '+1-555-987-6543',
    subject: 'Request for interview regarding new treatment protocol and FDA approval process',
    urgency: 'medium',
    status: 'new',
    priority: 'medium',
    source: 'website',
    created_at: '2024-03-15T10:00:00Z',
    updated_at: '2024-03-15T10:00:00Z',
    created_by: 'system',
    updated_by: 'system',
    is_deleted: false
  };

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

  const mockSubmissionResponse: InquirySubmissionResponse = {
    media_inquiry: mockMediaInquiry,
    correlation_id: 'corr-789-012-345',
    success: true,
    message: 'Media inquiry submitted successfully'
  };

  beforeEach(() => {
    client = new MediaInquiryRestClient();
  });

  describe('submitMediaInquiry', () => {
    it('should submit media inquiry with valid data', async () => {
      mockFetch.mockResolvedValue(createMockResponse(mockSubmissionResponse));

      const result = await client.submitMediaInquiry(mockStandardSubmission);

      expect(mockFetch).toHaveBeenCalledWith(
        'http://localhost:7220/api/v1/inquiries/media',
        expect.objectContaining({
          method: 'POST',
          headers: expect.objectContaining({
            'Content-Type': 'application/json'
          }),
          body: JSON.stringify(mockStandardSubmission)
        })
      );
      expect(result).toEqual(mockSubmissionResponse);
      expect(result.success).toBe(true);
      expect(result.media_inquiry?.outlet).toBe('Medical News Network');
    });

    it('should submit urgent media inquiry with deadline', async () => {
      const urgentResponse = {
        ...mockSubmissionResponse,
        media_inquiry: {
          ...mockMediaInquiry,
          contact_name: 'Tom Journalist',
          email: 'tom.journalist@tv.com',
          outlet: 'Health TV',
          title: 'Health Correspondent',
          phone: '+1-555-111-2222',
          media_type: 'television',
          deadline: '2024-03-16',
          urgency: 'high',
          subject: 'Breaking: New FDA approval for innovative treatment - urgent interview needed'
        }
      };

      mockFetch.mockResolvedValue(createMockResponse(urgentResponse));

      const result = await client.submitMediaInquiry(mockUrgentSubmission);

      expect(mockFetch).toHaveBeenCalledWith(
        'http://localhost:7220/api/v1/inquiries/media',
        expect.objectContaining({
          method: 'POST',
          headers: expect.objectContaining({
            'Content-Type': 'application/json'
          }),
          body: JSON.stringify(mockUrgentSubmission)
        })
      );
      expect(result.media_inquiry?.urgency).toBe('high');
      expect(result.media_inquiry?.deadline).toBe('2024-03-16');
      expect(result.media_inquiry?.media_type).toBe('television');
    });

    it('should handle different media types correctly', async () => {
      const printInquiry = {
        ...mockStandardSubmission,
        media_type: 'print' as const,
        outlet: 'Medical Journal',
        subject: 'Feature article on breakthrough medical research findings'
      };
      const printResponse = {
        ...mockSubmissionResponse,
        media_inquiry: {
          ...mockMediaInquiry,
          media_type: 'print',
          outlet: 'Medical Journal',
          subject: printInquiry.subject
        }
      };

      mockFetch.mockResolvedValue(createMockResponse(printResponse));

      const result = await client.submitMediaInquiry(printInquiry);

      expect(result.media_inquiry?.media_type).toBe('print');
      expect(result.media_inquiry?.outlet).toBe('Medical Journal');
    });

    it('should handle podcast media type correctly', async () => {
      const podcastInquiry = {
        ...mockStandardSubmission,
        media_type: 'podcast' as const,
        outlet: 'Healthcare Podcast',
        subject: 'Podcast interview about patient care innovations and treatment outcomes'
      };
      const podcastResponse = {
        ...mockSubmissionResponse,
        media_inquiry: {
          ...mockMediaInquiry,
          media_type: 'podcast',
          outlet: 'Healthcare Podcast',
          subject: podcastInquiry.subject
        }
      };

      mockFetch.mockResolvedValue(createMockResponse(podcastResponse));

      const result = await client.submitMediaInquiry(podcastInquiry);

      expect(result.media_inquiry?.media_type).toBe('podcast');
      expect(result.media_inquiry?.outlet).toBe('Healthcare Podcast');
    });

    it('should handle medical journal media type correctly', async () => {
      const journalInquiry = {
        ...mockStandardSubmission,
        media_type: 'medical-journal' as const,
        outlet: 'Journal of Medical Innovation',
        subject: 'Research publication opportunity for groundbreaking clinical study'
      };
      const journalResponse = {
        ...mockSubmissionResponse,
        media_inquiry: {
          ...mockMediaInquiry,
          media_type: 'medical-journal',
          outlet: 'Journal of Medical Innovation',
          subject: journalInquiry.subject
        }
      };

      mockFetch.mockResolvedValue(createMockResponse(journalResponse));

      const result = await client.submitMediaInquiry(journalInquiry);

      expect(result.media_inquiry?.media_type).toBe('medical-journal');
      expect(result.media_inquiry?.outlet).toBe('Journal of Medical Innovation');
    });

    it('should handle different urgency levels correctly', async () => {
      const lowUrgencyInquiry = {
        ...mockStandardSubmission,
        urgency: 'low' as const,
        subject: 'General inquiry about future research developments - no rush'
      };
      const lowUrgencyResponse = {
        ...mockSubmissionResponse,
        media_inquiry: {
          ...mockMediaInquiry,
          urgency: 'low',
          subject: lowUrgencyInquiry.subject
        }
      };

      mockFetch.mockResolvedValue(createMockResponse(lowUrgencyResponse));

      const result = await client.submitMediaInquiry(lowUrgencyInquiry);

      expect(result.media_inquiry?.urgency).toBe('low');
    });

    it('should handle inquiries without optional fields', async () => {
      const minimalInquiry = {
        contact_name: 'Basic Reporter',
        email: 'basic@news.com',
        outlet: 'Local News',
        title: 'Reporter',
        phone: '+1-555-999-8888',
        subject: 'General inquiry about your medical research programs and patient services',
        urgency: 'medium' as const
      };

      mockFetch.mockResolvedValue(createMockResponse(mockSubmissionResponse));

      const result = await client.submitMediaInquiry(minimalInquiry);

      expect(mockFetch).toHaveBeenCalledWith(
        'http://localhost:7220/api/v1/inquiries/media',
        expect.objectContaining({
          method: 'POST',
          headers: expect.objectContaining({
            'Content-Type': 'application/json'
          }),
          body: JSON.stringify(minimalInquiry)
        })
      );
      expect(result.success).toBe(true);
    });

    it('should handle submission errors', async () => {
      const errorResponse = {
        error: 'Validation failed',
        correlation_id: 'corr-error-789',
        success: false,
        message: 'Invalid media outlet'
      };

      mockFetch.mockRejectedValue(new Error('Network error'));

      await expect(client.submitMediaInquiry(mockStandardSubmission))
        .rejects.toThrow('Network error');
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
        success: false
      };

      mockFetch.mockResolvedValue(createMockResponse(validationErrorResponse));

      const result = await client.submitMediaInquiry(mockStandardSubmission);

      expect(result.success).toBe(false);
      expect(result.error).toBe('Validation failed');
      expect(result.validation_errors).toContain('Phone number is required for media inquiries');
    });

    it('should handle rate limiting responses', async () => {
      const rateLimitResponse = {
        error: 'Rate limit exceeded',
        correlation_id: 'corr-rate-limit',
        success: false,
        message: 'Too many requests. Please try again later.',
        retry_after: 180
      };

      mockFetch.mockResolvedValue(createMockResponse(rateLimitResponse));

      const result = await client.submitMediaInquiry(mockStandardSubmission);

      expect(result.success).toBe(false);
      expect(result.error).toBe('Rate limit exceeded');
      expect(result.retry_after).toBe(180);
    });
  });

  describe('getMediaInquiry', () => {
    it('should retrieve media inquiry by ID', async () => {
      const getResponse = {
        media_inquiry: mockMediaInquiry,
        correlation_id: 'corr-get-789'
      };

      mockFetch.mockResolvedValue(createMockResponse(getResponse));

      const result = await client.getMediaInquiry('789e0123-e89b-12d3-a456-426614174002');

      expect(mockFetch).toHaveBeenCalledWith(
        'http://localhost:7220/api/v1/inquiries/media/789e0123-e89b-12d3-a456-426614174002',
        {
          method: 'GET',
          headers: {
            'Accept': 'application/json',
            'Content-Type': 'application/json',
            'X-Retry-Attempt': '1'
          },
          signal: expect.any(Object)
        }
      );
      expect(result).toEqual(getResponse);
      expect(result.media_inquiry?.inquiry_id).toBe('789e0123-e89b-12d3-a456-426614174002');
    });

    it('should handle inquiry not found', async () => {
      const notFoundResponse = {
        error: 'Media inquiry not found',
        correlation_id: 'corr-not-found',
        success: false
      };

      mockFetch.mockResolvedValue(createMockResponse(notFoundResponse));

      const result = await client.getMediaInquiry('non-existent-id');

      expect(result.error).toBe('Media inquiry not found');
      expect(result.success).toBe(false);
    });
  });

  describe('error handling', () => {
    it('should handle network errors appropriately', async () => {
      const networkError = new Error('Network connection failed');
      mockFetch.mockRejectedValue(networkError);

      await expect(client.submitMediaInquiry(mockStandardSubmission))
        .rejects.toThrow('Network connection failed');
    });

    it('should handle timeout errors', async () => {
      const timeoutError = new Error('Request timeout');
      mockFetch.mockRejectedValue(timeoutError);

      await expect(client.submitMediaInquiry(mockStandardSubmission))
        .rejects.toThrow('Request timeout');
    });

    it('should handle malformed responses', async () => {
      mockFetch.mockResolvedValue(createMockResponse(null));

      const result = await client.submitMediaInquiry(mockStandardSubmission);
      expect(result).toBe(null);
    });
  });


  describe('response handling', () => {
    it('should properly parse successful submission response', async () => {
      mockFetch.mockResolvedValue(createMockResponse(mockSubmissionResponse));

      const result = await client.submitMediaInquiry(mockStandardSubmission);

      expect(result.success).toBe(true);
      expect(result.message).toBe('Media inquiry submitted successfully');
      expect(result.correlation_id).toBe('corr-789-012-345');
      expect(result.media_inquiry).toMatchObject(mockMediaInquiry);
    });

    it('should handle responses with correlation IDs', async () => {
      mockFetch.mockResolvedValue(createMockResponse(mockSubmissionResponse));

      const result = await client.submitMediaInquiry(mockStandardSubmission);

      expect(result.correlation_id).toBeTruthy();
      expect(typeof result.correlation_id).toBe('string');
    });
  });

  describe('domain-specific media logic', () => {
    it('should handle radio media type correctly', async () => {
      const radioInquiry = {
        ...mockStandardSubmission,
        media_type: 'radio' as const,
        outlet: 'Health Radio Network',
        subject: 'Radio interview about patient care improvements and treatment accessibility'
      };

      mockFetch.mockResolvedValue(createMockResponse({
        ...mockSubmissionResponse,
        media_inquiry: {
          ...mockMediaInquiry,
          media_type: 'radio',
          outlet: 'Health Radio Network',
          subject: radioInquiry.subject
        }
      }));

      const result = await client.submitMediaInquiry(radioInquiry);

      expect(result.media_inquiry?.media_type).toBe('radio');
      expect(result.media_inquiry?.outlet).toBe('Health Radio Network');
    });

    it('should handle same-day deadline urgency', async () => {
      const today = new Date().toISOString().split('T')[0];
      const sameDayInquiry = {
        ...mockUrgentSubmission,
        deadline: today,
        urgency: 'high' as const,
        subject: 'Same-day deadline: Breaking news requires immediate response'
      };

      mockFetch.mockResolvedValue(createMockResponse({
        ...mockSubmissionResponse,
        media_inquiry: {
          ...mockMediaInquiry,
          deadline: today,
          urgency: 'high',
          subject: sameDayInquiry.subject
        }
      }));

      const result = await client.submitMediaInquiry(sameDayInquiry);

      expect(result.media_inquiry?.deadline).toBe(today);
      expect(result.media_inquiry?.urgency).toBe('high');
    });

    it('should handle television media type with urgent deadline', async () => {
      const tvInquiry = {
        ...mockUrgentSubmission,
        media_type: 'television' as const,
        urgency: 'high' as const,
        subject: 'Live TV interview needed for breaking medical news story'
      };

      mockFetch.mockResolvedValue(createMockResponse({
        ...mockSubmissionResponse,
        media_inquiry: {
          ...mockMediaInquiry,
          media_type: 'television',
          urgency: 'high',
          subject: tvInquiry.subject
        }
      }));

      const result = await client.submitMediaInquiry(tvInquiry);

      expect(result.media_inquiry?.media_type).toBe('television');
      expect(result.media_inquiry?.urgency).toBe('high');
      expect(result.media_inquiry?.subject).toContain('Live TV interview');
    });

    it('should validate required phone number for media inquiries', async () => {
      const inquiryWithoutPhone = {
        ...mockStandardSubmission,
        phone: undefined as any
      };

      const phoneRequiredResponse = {
        error: 'Validation failed',
        validation_errors: ['Phone number is required for media inquiries'],
        correlation_id: 'corr-phone-required',
        success: false
      };

      mockFetch.mockResolvedValue(createMockResponse(phoneRequiredResponse));

      const result = await client.submitMediaInquiry(inquiryWithoutPhone);

      expect(result.success).toBe(false);
      expect(result.validation_errors).toContain('Phone number is required for media inquiries');
    });

    it('should handle future deadline correctly', async () => {
      const futureDate = '2024-03-20';
      const futureDeadlineInquiry = {
        ...mockStandardSubmission,
        deadline: futureDate,
        urgency: 'low' as const,
        subject: 'Feature story with flexible timeline - advance planning'
      };

      mockFetch.mockResolvedValue(createMockResponse({
        ...mockSubmissionResponse,
        media_inquiry: {
          ...mockMediaInquiry,
          deadline: futureDate,
          urgency: 'low',
          subject: futureDeadlineInquiry.subject
        }
      }));

      const result = await client.submitMediaInquiry(futureDeadlineInquiry);

      expect(result.media_inquiry?.deadline).toBe(futureDate);
      expect(result.media_inquiry?.urgency).toBe('low');
    });
  });
});