import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { DonationsInquiryRestClient } from './DonationsInquiryRestClient';
import type { DonationsInquiry, DonationsInquirySubmission, InquirySubmissionResponse } from '../inquiries/types';

// Mock the BaseRestClient
vi.mock('./BaseRestClient');

describe('DonationsInquiryRestClient', () => {
  let client: DonationsInquiryRestClient;
  let mockPost: ReturnType<typeof vi.fn>;
  let mockGet: ReturnType<typeof vi.fn>;

  const mockDonationsInquiry: DonationsInquiry = {
    inquiry_id: '456e7890-e89b-12d3-a456-426614174001',
    contact_name: 'Mary Johnson',
    email: 'mary.johnson@email.com',
    donor_type: 'individual',
    message: 'I would like to make a donation to support your research initiatives and patient care programs.',
    status: 'new',
    priority: 'medium',
    source: 'website',
    created_at: '2024-03-15T10:00:00Z',
    updated_at: '2024-03-15T10:00:00Z',
    created_by: 'system',
    updated_by: 'system',
    is_deleted: false
  };

  const mockIndividualSubmission: DonationsInquirySubmission = {
    contact_name: 'Mary Johnson',
    email: 'mary.johnson@email.com',
    phone: '+1-555-987-6543',
    donor_type: 'individual',
    interest_area: 'research-funding',
    preferred_amount_range: '1000-5000',
    donation_frequency: 'monthly',
    message: 'I would like to make a donation to support your research initiatives and patient care programs.'
  };

  const mockCorporateSubmission: DonationsInquirySubmission = {
    contact_name: 'Robert Wilson',
    email: 'robert.wilson@foundation.org',
    organization: 'Wilson Foundation',
    donor_type: 'foundation',
    interest_area: 'clinic-development',
    preferred_amount_range: '25000-100000',
    donation_frequency: 'annually',
    message: 'Our foundation is interested in funding medical research projects and clinic development initiatives to improve patient care and advance medical science.'
  };

  const mockSubmissionResponse: InquirySubmissionResponse = {
    donations_inquiry: mockDonationsInquiry,
    correlation_id: 'corr-456-789-012',
    success: true,
    message: 'Donations inquiry submitted successfully'
  };

  beforeEach(() => {
    mockPost = vi.fn();
    mockGet = vi.fn();
    
    client = new DonationsInquiryRestClient();
    // Mock the inherited methods from BaseRestClient
    (client as any).post = mockPost;
    (client as any).get = mockGet;
  });

  afterEach(() => {
    vi.clearAllMocks();
  });

  describe('submitDonationsInquiry', () => {
    it('should submit individual donor inquiry with valid data', async () => {
      mockPost.mockResolvedValue(mockSubmissionResponse);

      const result = await client.submitDonationsInquiry(mockIndividualSubmission);

      expect(mockPost).toHaveBeenCalledWith('/api/inquiries/donations', mockIndividualSubmission);
      expect(result).toEqual(mockSubmissionResponse);
      expect(result.success).toBe(true);
      expect(result.donations_inquiry?.donor_type).toBe('individual');
    });

    it('should submit corporate donor inquiry with organization', async () => {
      const corporateResponse = {
        ...mockSubmissionResponse,
        donations_inquiry: {
          ...mockDonationsInquiry,
          contact_name: 'Robert Wilson',
          email: 'robert.wilson@foundation.org',
          organization: 'Wilson Foundation',
          donor_type: 'foundation',
          interest_area: 'clinic-development',
          preferred_amount_range: '25000-100000',
          donation_frequency: 'annually'
        }
      };

      mockPost.mockResolvedValue(corporateResponse);

      const result = await client.submitDonationsInquiry(mockCorporateSubmission);

      expect(mockPost).toHaveBeenCalledWith('/api/inquiries/donations', mockCorporateSubmission);
      expect(result.donations_inquiry?.donor_type).toBe('foundation');
      expect(result.donations_inquiry?.organization).toBe('Wilson Foundation');
      expect(result.donations_inquiry?.interest_area).toBe('clinic-development');
    });

    it('should handle different donor types correctly', async () => {
      const estateInquiry = {
        ...mockIndividualSubmission,
        donor_type: 'estate' as const,
        message: 'We are executing an estate donation as specified in the will.'
      };
      const estateResponse = {
        ...mockSubmissionResponse,
        donations_inquiry: {
          ...mockDonationsInquiry,
          donor_type: 'estate',
          message: estateInquiry.message
        }
      };

      mockPost.mockResolvedValue(estateResponse);

      const result = await client.submitDonationsInquiry(estateInquiry);

      expect(result.donations_inquiry?.donor_type).toBe('estate');
      expect(result.donations_inquiry?.message).toContain('estate donation');
    });

    it('should handle different interest areas correctly', async () => {
      const patientCareInquiry = {
        ...mockIndividualSubmission,
        interest_area: 'patient-care' as const,
        message: 'I want to support patient care programs and equipment purchases.'
      };
      const patientCareResponse = {
        ...mockSubmissionResponse,
        donations_inquiry: {
          ...mockDonationsInquiry,
          interest_area: 'patient-care',
          message: patientCareInquiry.message
        }
      };

      mockPost.mockResolvedValue(patientCareResponse);

      const result = await client.submitDonationsInquiry(patientCareInquiry);

      expect(result.donations_inquiry?.interest_area).toBe('patient-care');
      expect(result.donations_inquiry?.message).toContain('patient care');
    });

    it('should handle different donation frequencies', async () => {
      const quarterlyInquiry = {
        ...mockIndividualSubmission,
        donation_frequency: 'quarterly' as const,
        message: 'I would like to set up quarterly donations to support ongoing research.'
      };
      const quarterlyResponse = {
        ...mockSubmissionResponse,
        donations_inquiry: {
          ...mockDonationsInquiry,
          donation_frequency: 'quarterly'
        }
      };

      mockPost.mockResolvedValue(quarterlyResponse);

      const result = await client.submitDonationsInquiry(quarterlyInquiry);

      expect(result.donations_inquiry?.donation_frequency).toBe('quarterly');
    });

    it('should handle different amount ranges', async () => {
      const largeAmountInquiry = {
        ...mockCorporateSubmission,
        preferred_amount_range: 'over-100000' as const,
        message: 'Our foundation is prepared to make a substantial donation exceeding $100,000.'
      };
      const largeAmountResponse = {
        ...mockSubmissionResponse,
        donations_inquiry: {
          ...mockDonationsInquiry,
          preferred_amount_range: 'over-100000'
        }
      };

      mockPost.mockResolvedValue(largeAmountResponse);

      const result = await client.submitDonationsInquiry(largeAmountInquiry);

      expect(result.donations_inquiry?.preferred_amount_range).toBe('over-100000');
    });

    it('should handle optional fields correctly', async () => {
      const minimalInquiry = {
        contact_name: 'Simple Donor',
        email: 'simple@donor.com',
        donor_type: 'individual' as const,
        message: 'I would like to make a simple donation without specific preferences.'
      };

      mockPost.mockResolvedValue(mockSubmissionResponse);

      const result = await client.submitDonationsInquiry(minimalInquiry);

      expect(mockPost).toHaveBeenCalledWith('/api/inquiries/donations', minimalInquiry);
      expect(result.success).toBe(true);
    });

    it('should handle submission errors', async () => {
      const errorResponse = {
        error: 'Validation failed',
        correlation_id: 'corr-error-456',
        success: false,
        message: 'Invalid donor type'
      };

      mockPost.mockRejectedValue(new Error('Network error'));

      await expect(client.submitDonationsInquiry(mockIndividualSubmission))
        .rejects.toThrow('Network error');
    });

    it('should handle validation errors from backend', async () => {
      const validationErrorResponse = {
        error: 'Validation failed',
        validation_errors: [
          'Donor type is required',
          'Message must be at least 20 characters'
        ],
        correlation_id: 'corr-validation-error',
        success: false
      };

      mockPost.mockResolvedValue(validationErrorResponse);

      const result = await client.submitDonationsInquiry(mockIndividualSubmission);

      expect(result.success).toBe(false);
      expect(result.error).toBe('Validation failed');
      expect(result.validation_errors).toContain('Donor type is required');
    });

    it('should handle rate limiting responses', async () => {
      const rateLimitResponse = {
        error: 'Rate limit exceeded',
        correlation_id: 'corr-rate-limit',
        success: false,
        message: 'Too many requests. Please try again later.',
        retry_after: 120
      };

      mockPost.mockResolvedValue(rateLimitResponse);

      const result = await client.submitDonationsInquiry(mockIndividualSubmission);

      expect(result.success).toBe(false);
      expect(result.error).toBe('Rate limit exceeded');
      expect(result.retry_after).toBe(120);
    });
  });

  describe('getDonationsInquiry', () => {
    it('should retrieve donations inquiry by ID', async () => {
      const getResponse = {
        donations_inquiry: mockDonationsInquiry,
        correlation_id: 'corr-get-456'
      };

      mockGet.mockResolvedValue(getResponse);

      const result = await client.getDonationsInquiry('456e7890-e89b-12d3-a456-426614174001');

      expect(mockGet).toHaveBeenCalledWith('/api/inquiries/donations/456e7890-e89b-12d3-a456-426614174001');
      expect(result).toEqual(getResponse);
      expect(result.donations_inquiry?.inquiry_id).toBe('456e7890-e89b-12d3-a456-426614174001');
    });

    it('should handle inquiry not found', async () => {
      const notFoundResponse = {
        error: 'Donations inquiry not found',
        correlation_id: 'corr-not-found',
        success: false
      };

      mockGet.mockResolvedValue(notFoundResponse);

      const result = await client.getDonationsInquiry('non-existent-id');

      expect(result.error).toBe('Donations inquiry not found');
      expect(result.success).toBe(false);
    });
  });

  describe('error handling', () => {
    it('should handle network errors appropriately', async () => {
      const networkError = new Error('Network connection failed');
      mockPost.mockRejectedValue(networkError);

      await expect(client.submitDonationsInquiry(mockIndividualSubmission))
        .rejects.toThrow('Network connection failed');
    });

    it('should handle timeout errors', async () => {
      const timeoutError = new Error('Request timeout');
      mockPost.mockRejectedValue(timeoutError);

      await expect(client.submitDonationsInquiry(mockIndividualSubmission))
        .rejects.toThrow('Request timeout');
    });

    it('should handle malformed responses', async () => {
      mockPost.mockResolvedValue(null);

      await expect(client.submitDonationsInquiry(mockIndividualSubmission))
        .rejects.toThrow();
    });
  });

  describe('request formatting', () => {
    it('should properly format individual donor submission data', async () => {
      mockPost.mockResolvedValue(mockSubmissionResponse);

      await client.submitDonationsInquiry(mockIndividualSubmission);

      const calledWith = mockPost.mock.calls[0][1];
      expect(calledWith).toMatchObject({
        contact_name: 'Mary Johnson',
        email: 'mary.johnson@email.com',
        donor_type: 'individual',
        interest_area: 'research-funding',
        preferred_amount_range: '1000-5000',
        donation_frequency: 'monthly',
        message: expect.stringContaining('research initiatives')
      });
    });

    it('should properly format corporate donor submission data', async () => {
      mockPost.mockResolvedValue(mockSubmissionResponse);

      await client.submitDonationsInquiry(mockCorporateSubmission);

      const calledWith = mockPost.mock.calls[0][1];
      expect(calledWith).toMatchObject({
        contact_name: 'Robert Wilson',
        email: 'robert.wilson@foundation.org',
        organization: 'Wilson Foundation',
        donor_type: 'foundation',
        interest_area: 'clinic-development',
        preferred_amount_range: '25000-100000',
        donation_frequency: 'annually'
      });
    });

    it('should include optional fields when provided', async () => {
      const submissionWithOptionals = {
        ...mockIndividualSubmission,
        phone: '+1-555-123-4567'
      };

      mockPost.mockResolvedValue(mockSubmissionResponse);

      await client.submitDonationsInquiry(submissionWithOptionals);

      const calledWith = mockPost.mock.calls[0][1];
      expect(calledWith.phone).toBe('+1-555-123-4567');
    });

    it('should not include undefined optional fields', async () => {
      const submissionWithUndefined = {
        ...mockIndividualSubmission,
        phone: undefined,
        organization: undefined
      };

      mockPost.mockResolvedValue(mockSubmissionResponse);

      await client.submitDonationsInquiry(submissionWithUndefined);

      const calledWith = mockPost.mock.calls[0][1];
      expect(calledWith.phone).toBeUndefined();
      expect(calledWith.organization).toBeUndefined();
    });
  });

  describe('response handling', () => {
    it('should properly parse successful submission response', async () => {
      mockPost.mockResolvedValue(mockSubmissionResponse);

      const result = await client.submitDonationsInquiry(mockIndividualSubmission);

      expect(result.success).toBe(true);
      expect(result.message).toBe('Donations inquiry submitted successfully');
      expect(result.correlation_id).toBe('corr-456-789-012');
      expect(result.donations_inquiry).toMatchObject(mockDonationsInquiry);
    });

    it('should handle responses with correlation IDs', async () => {
      mockPost.mockResolvedValue(mockSubmissionResponse);

      const result = await client.submitDonationsInquiry(mockIndividualSubmission);

      expect(result.correlation_id).toBeTruthy();
      expect(typeof result.correlation_id).toBe('string');
    });
  });

  describe('domain-specific donations logic', () => {
    it('should handle undisclosed amount preferences', async () => {
      const undisclosedInquiry = {
        ...mockIndividualSubmission,
        preferred_amount_range: 'undisclosed' as const,
        message: 'I prefer not to disclose the amount at this time.'
      };

      mockPost.mockResolvedValue({
        ...mockSubmissionResponse,
        donations_inquiry: {
          ...mockDonationsInquiry,
          preferred_amount_range: 'undisclosed'
        }
      });

      const result = await client.submitDonationsInquiry(undisclosedInquiry);

      expect(result.donations_inquiry?.preferred_amount_range).toBe('undisclosed');
    });

    it('should handle general support interest area', async () => {
      const generalSupportInquiry = {
        ...mockIndividualSubmission,
        interest_area: 'general-support' as const,
        message: 'I want to support the organization in whatever way is most needed.'
      };

      mockPost.mockResolvedValue({
        ...mockSubmissionResponse,
        donations_inquiry: {
          ...mockDonationsInquiry,
          interest_area: 'general-support'
        }
      });

      const result = await client.submitDonationsInquiry(generalSupportInquiry);

      expect(result.donations_inquiry?.interest_area).toBe('general-support');
    });

    it('should handle equipment funding interest', async () => {
      const equipmentInquiry = {
        ...mockCorporateSubmission,
        interest_area: 'equipment' as const,
        message: 'Our company would like to fund medical equipment purchases for the clinic.'
      };

      mockPost.mockResolvedValue({
        ...mockSubmissionResponse,
        donations_inquiry: {
          ...mockDonationsInquiry,
          interest_area: 'equipment'
        }
      });

      const result = await client.submitDonationsInquiry(equipmentInquiry);

      expect(result.donations_inquiry?.interest_area).toBe('equipment');
    });

    it('should handle one-time donation frequency', async () => {
      const oneTimeInquiry = {
        ...mockIndividualSubmission,
        donation_frequency: 'one-time' as const,
        message: 'I would like to make a single donation in memory of my loved one.'
      };

      mockPost.mockResolvedValue({
        ...mockSubmissionResponse,
        donations_inquiry: {
          ...mockDonationsInquiry,
          donation_frequency: 'one-time'
        }
      });

      const result = await client.submitDonationsInquiry(oneTimeInquiry);

      expect(result.donations_inquiry?.donation_frequency).toBe('one-time');
    });
  });
});