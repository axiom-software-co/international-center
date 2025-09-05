import { describe, it, expect, vi, beforeEach } from 'vitest';
import { ref, nextTick } from 'vue';
import { useDonationsInquiry, useDonationsInquirySubmission } from './useDonationsInquiry';
import { DonationsInquiryRestClient } from '../rest/DonationsInquiryRestClient';
import { RestError } from '../rest/BaseRestClient';
import type { DonationsInquirySubmission, InquirySubmissionResponse } from '../inquiries/types';

// Mock the DonationsInquiryRestClient with hoisted functions
const {
  mockSubmitDonationsInquiry,
  mockGetDonationsInquiry,
  MockedDonationsInquiryRestClient
} = vi.hoisted(() => {
  const mockSubmit = vi.fn();
  const mockGet = vi.fn();
  
  return {
    mockSubmitDonationsInquiry: mockSubmit,
    mockGetDonationsInquiry: mockGet,
    MockedDonationsInquiryRestClient: vi.fn().mockImplementation(() => ({
      submitDonationsInquiry: mockSubmit,
      getDonationsInquiry: mockGet,
    }))
  };
});

vi.mock('../rest/DonationsInquiryRestClient', () => ({
  DonationsInquiryRestClient: MockedDonationsInquiryRestClient
}));

describe('useDonationsInquiry composables', () => {
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

  const mockSuccessResponse: InquirySubmissionResponse = {
    donations_inquiry: {
      inquiry_id: '456e7890-e89b-12d3-a456-426614174001',
      contact_name: 'Mary Johnson',
      email: 'mary.johnson@email.com',
      phone: '+1-555-987-6543',
      donor_type: 'individual',
      interest_area: 'research-funding',
      preferred_amount_range: '1000-5000',
      donation_frequency: 'monthly',
      message: mockIndividualSubmission.message,
      status: 'new',
      priority: 'medium',
      source: 'website',
      created_at: '2024-03-15T10:00:00Z',
      updated_at: '2024-03-15T10:00:00Z',
      created_by: 'system',
      updated_by: 'system',
      is_deleted: false
    },
    correlation_id: 'corr-456-789-012',
    success: true,
    message: 'Donations inquiry submitted successfully'
  };

  beforeEach(() => {
    vi.clearAllMocks();
  });

  describe('useDonationsInquirySubmission', () => {

    it('should initialize with correct default values', () => {
      const { isSubmitting, error, response, isSuccess, isError } = useDonationsInquirySubmission();

      expect(isSubmitting.value).toBe(false);
      expect(error.value).toBe(null);
      expect(response.value).toBe(null);
      expect(isSuccess.value).toBe(false);
      expect(isError.value).toBe(false);
    });

    it('should submit individual donations inquiry successfully', async () => {
      mockSubmitDonationsInquiry.mockResolvedValueOnce(mockSuccessResponse);

      const { submitInquiry, isSubmitting, response, isSuccess, error } = useDonationsInquirySubmission();

      expect(isSubmitting.value).toBe(false);

      const submissionPromise = submitInquiry(mockIndividualSubmission);
      
      // Should be submitting during the request
      expect(isSubmitting.value).toBe(true);

      await submissionPromise;
      await nextTick();

      // Should complete successfully
      expect(isSubmitting.value).toBe(false);
      expect(isSuccess.value).toBe(true);
      expect(error.value).toBe(null);
      expect(response.value).toEqual(mockSuccessResponse);
      expect(response.value?.donations_inquiry?.donor_type).toBe('individual');
      expect(mockSubmitDonationsInquiry).toHaveBeenCalledWith(mockIndividualSubmission);
    });

    it('should handle corporate donations inquiry correctly', async () => {
      const corporateResponse = {
        ...mockSuccessResponse,
        donations_inquiry: {
          ...mockSuccessResponse.donations_inquiry!,
          contact_name: 'Robert Wilson',
          email: 'robert.wilson@foundation.org',
          organization: 'Wilson Foundation',
          donor_type: 'foundation',
          interest_area: 'clinic-development',
          preferred_amount_range: '25000-100000',
          donation_frequency: 'annually',
          message: mockCorporateSubmission.message
        }
      };

      mockSubmitDonationsInquiry.mockResolvedValueOnce(corporateResponse);

      const { submitInquiry, response, isSuccess } = useDonationsInquirySubmission();

      await submitInquiry(mockCorporateSubmission);
      await nextTick();

      expect(isSuccess.value).toBe(true);
      expect(response.value?.donations_inquiry?.donor_type).toBe('foundation');
      expect(response.value?.donations_inquiry?.organization).toBe('Wilson Foundation');
      expect(response.value?.donations_inquiry?.interest_area).toBe('clinic-development');
    });

    it('should handle different donor types correctly', async () => {
      const estateInquiry = {
        ...mockIndividualSubmission,
        donor_type: 'estate' as const,
        message: 'We are executing an estate donation as specified in the will to support medical research.'
      };

      const estateResponse = {
        ...mockSuccessResponse,
        donations_inquiry: {
          ...mockSuccessResponse.donations_inquiry!,
          donor_type: 'estate',
          message: estateInquiry.message
        }
      };

      mockSubmitDonationsInquiry.mockResolvedValueOnce(estateResponse);

      const { submitInquiry, response, isSuccess } = useDonationsInquirySubmission();

      await submitInquiry(estateInquiry);
      await nextTick();

      expect(isSuccess.value).toBe(true);
      expect(response.value?.donations_inquiry?.donor_type).toBe('estate');
      expect(response.value?.donations_inquiry?.message).toContain('estate donation');
    });

    it('should handle different interest areas correctly', async () => {
      const patientCareInquiry = {
        ...mockIndividualSubmission,
        interest_area: 'patient-care' as const,
        message: 'I want to support patient care programs and help improve treatment outcomes for patients.'
      };

      const patientCareResponse = {
        ...mockSuccessResponse,
        donations_inquiry: {
          ...mockSuccessResponse.donations_inquiry!,
          interest_area: 'patient-care',
          message: patientCareInquiry.message
        }
      };

      mockSubmitDonationsInquiry.mockResolvedValueOnce(patientCareResponse);

      const { submitInquiry, response, isSuccess } = useDonationsInquirySubmission();

      await submitInquiry(patientCareInquiry);
      await nextTick();

      expect(isSuccess.value).toBe(true);
      expect(response.value?.donations_inquiry?.interest_area).toBe('patient-care');
    });

    it('should handle different amount ranges correctly', async () => {
      const largeAmountInquiry = {
        ...mockCorporateSubmission,
        preferred_amount_range: 'over-100000' as const,
        message: 'Our foundation is prepared to make a substantial donation exceeding $100,000.'
      };

      const largeAmountResponse = {
        ...mockSuccessResponse,
        donations_inquiry: {
          ...mockSuccessResponse.donations_inquiry!,
          preferred_amount_range: 'over-100000'
        }
      };

      mockSubmitDonationsInquiry.mockResolvedValueOnce(largeAmountResponse);

      const { submitInquiry, response, isSuccess } = useDonationsInquirySubmission();

      await submitInquiry(largeAmountInquiry);
      await nextTick();

      expect(isSuccess.value).toBe(true);
      expect(response.value?.donations_inquiry?.preferred_amount_range).toBe('over-100000');
    });

    it('should handle different donation frequencies correctly', async () => {
      const quarterlyInquiry = {
        ...mockIndividualSubmission,
        donation_frequency: 'quarterly' as const,
        message: 'I would like to set up quarterly donations to provide consistent support.'
      };

      const quarterlyResponse = {
        ...mockSuccessResponse,
        donations_inquiry: {
          ...mockSuccessResponse.donations_inquiry!,
          donation_frequency: 'quarterly'
        }
      };

      mockSubmitDonationsInquiry.mockResolvedValueOnce(quarterlyResponse);

      const { submitInquiry, response, isSuccess } = useDonationsInquirySubmission();

      await submitInquiry(quarterlyInquiry);
      await nextTick();

      expect(isSuccess.value).toBe(true);
      expect(response.value?.donations_inquiry?.donation_frequency).toBe('quarterly');
    });

    it('should handle submission errors', async () => {
      const errorMessage = 'Network connection failed';
      mockSubmitDonationsInquiry.mockRejectedValueOnce(new Error(errorMessage));

      const { submitInquiry, isSubmitting, error, isError, isSuccess } = useDonationsInquirySubmission();

      const submissionPromise = submitInquiry(mockIndividualSubmission);
      
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
          'Donor type is required',
          'Message must be at least 20 characters'
        ],
        correlation_id: 'corr-validation-error',
        success: false,
        message: 'Please correct the following errors'
      };

      mockSubmitDonationsInquiry.mockResolvedValueOnce(validationErrorResponse);

      const { submitInquiry, error, isError, response, isSuccess } = useDonationsInquirySubmission();

      await submitInquiry(mockIndividualSubmission);
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
        retry_after: 120
      };

      mockSubmitDonationsInquiry.mockResolvedValueOnce(rateLimitResponse);

      const { submitInquiry, error, isError, response } = useDonationsInquirySubmission();

      await submitInquiry(mockIndividualSubmission);
      await nextTick();

      expect(isError.value).toBe(true);
      expect(error.value).toBe('Rate limit exceeded');
      expect(response.value?.retry_after).toBe(120);
    });

    it('should reset state correctly', async () => {
      mockSubmitDonationsInquiry.mockResolvedValueOnce(mockSuccessResponse);

      const { submitInquiry, error, response, reset } = useDonationsInquirySubmission();

      // First submission
      await submitInquiry(mockIndividualSubmission);
      await nextTick();

      expect(response.value).toEqual(mockSuccessResponse);

      // Reset state
      reset();

      expect(error.value).toBe(null);
      expect(response.value).toBe(null);
    });

    it('should handle minimal submission without optional fields', async () => {
      const minimalSubmission = {
        contact_name: 'Simple Donor',
        email: 'simple@donor.com',
        donor_type: 'individual' as const,
        message: 'I would like to make a simple donation to support your important work.'
      };

      const minimalResponse = {
        ...mockSuccessResponse,
        donations_inquiry: {
          ...mockSuccessResponse.donations_inquiry!,
          contact_name: 'Simple Donor',
          email: 'simple@donor.com',
          message: minimalSubmission.message,
          phone: undefined,
          organization: undefined,
          interest_area: undefined,
          preferred_amount_range: undefined,
          donation_frequency: undefined
        }
      };

      mockSubmitDonationsInquiry.mockResolvedValueOnce(minimalResponse);

      const { submitInquiry, response, isSuccess } = useDonationsInquirySubmission();

      await submitInquiry(minimalSubmission);
      await nextTick();

      expect(isSuccess.value).toBe(true);
      expect(response.value?.donations_inquiry?.contact_name).toBe('Simple Donor');
    });

    it('should handle undisclosed amount preference', async () => {
      const undisclosedInquiry = {
        ...mockIndividualSubmission,
        preferred_amount_range: 'undisclosed' as const,
        message: 'I prefer not to disclose the amount at this time but am committed to supporting your cause.'
      };

      const undisclosedResponse = {
        ...mockSuccessResponse,
        donations_inquiry: {
          ...mockSuccessResponse.donations_inquiry!,
          preferred_amount_range: 'undisclosed'
        }
      };

      mockSubmitDonationsInquiry.mockResolvedValueOnce(undisclosedResponse);

      const { submitInquiry, response, isSuccess } = useDonationsInquirySubmission();

      await submitInquiry(undisclosedInquiry);
      await nextTick();

      expect(isSuccess.value).toBe(true);
      expect(response.value?.donations_inquiry?.preferred_amount_range).toBe('undisclosed');
    });
  });

  describe('useDonationsInquiry', () => {
    const mockDonationsInquiry = {
      inquiry_id: '456e7890-e89b-12d3-a456-426614174001',
      contact_name: 'Mary Johnson',
      email: 'mary.johnson@email.com',
      donor_type: 'individual' as const,
      interest_area: 'research-funding' as const,
      preferred_amount_range: '1000-5000' as const,
      donation_frequency: 'monthly' as const,
      message: 'I would like to make a donation to support your research initiatives.',
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
      const { inquiry, loading, error } = useDonationsInquiry(ref(null));

      expect(inquiry.value).toBe(null);
      expect(loading.value).toBe(false);
      expect(error.value).toBe(null);
    });

    it('should fetch donations inquiry by ID', async () => {
      const mockResponse = {
        donations_inquiry: mockDonationsInquiry,
        correlation_id: 'corr-get-456'
      };

      mockGetDonationsInquiry.mockResolvedValueOnce(mockResponse);

      const inquiryId = ref('456e7890-e89b-12d3-a456-426614174001');
      const { inquiry, loading, error } = useDonationsInquiry(inquiryId);

      await nextTick();

      expect(inquiry.value).toEqual(mockDonationsInquiry);
      expect(loading.value).toBe(false);
      expect(error.value).toBe(null);
      expect(mockGetDonationsInquiry).toHaveBeenCalledWith('456e7890-e89b-12d3-a456-426614174001');
    });

    it('should handle inquiry not found', async () => {
      const notFoundResponse = {
        error: 'Donations inquiry not found',
        correlation_id: 'corr-not-found',
        success: false
      };

      mockGetDonationsInquiry.mockResolvedValueOnce(notFoundResponse);

      const inquiryId = ref('non-existent-id');
      const { inquiry, loading, error } = useDonationsInquiry(inquiryId);

      await nextTick();

      expect(inquiry.value).toBe(null);
      expect(loading.value).toBe(false);
      expect(error.value).toBe('Donations inquiry not found');
    });

    it('should handle fetch errors', async () => {
      const errorMessage = 'Failed to fetch donations inquiry';
      mockGetDonationsInquiry.mockRejectedValueOnce(new Error(errorMessage));

      const inquiryId = ref('456e7890-e89b-12d3-a456-426614174001');
      const { inquiry, error } = useDonationsInquiry(inquiryId);

      await nextTick();

      expect(inquiry.value).toBe(null);
      expect(error.value).toBe(errorMessage);
    });

    it('should refetch when inquiry ID changes', async () => {
      const mockResponse1 = {
        donations_inquiry: { ...mockDonationsInquiry, inquiry_id: 'id-1' },
        correlation_id: 'corr-1'
      };

      const mockResponse2 = {
        donations_inquiry: { ...mockDonationsInquiry, inquiry_id: 'id-2', donor_type: 'foundation' as const },
        correlation_id: 'corr-2'
      };

      mockGetDonationsInquiry.mockResolvedValueOnce(mockResponse1);

      const inquiryId = ref('id-1');
      const { inquiry } = useDonationsInquiry(inquiryId);

      await nextTick();
      expect(inquiry.value?.inquiry_id).toBe('id-1');

      // Change ID and mock second response
      mockGetDonationsInquiry.mockResolvedValueOnce(mockResponse2);
      inquiryId.value = 'id-2';

      await nextTick();
      expect(inquiry.value?.inquiry_id).toBe('id-2');
      expect(inquiry.value?.donor_type).toBe('foundation');
      expect(mockGetDonationsInquiry).toHaveBeenCalledTimes(2);
    });

    it('should handle loading states correctly', async () => {
      const mockResponse = {
        donations_inquiry: mockDonationsInquiry,
        correlation_id: 'corr-loading-test'
      };

      let resolvePromise: (value: any) => void;
      const pendingPromise = new Promise(resolve => {
        resolvePromise = resolve;
      });
      mockGetDonationsInquiry.mockReturnValueOnce(pendingPromise);

      const inquiryId = ref('456e7890-e89b-12d3-a456-426614174001');
      const { inquiry, loading, error } = useDonationsInquiry(inquiryId);

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
      expect(inquiry.value).toEqual(mockDonationsInquiry);
      expect(error.value).toBe(null);
    });
  });

  describe('error handling edge cases', () => {
    it('should handle network timeouts', async () => {
      const timeoutError = new Error('Request timeout');
      mockSubmitDonationsInquiry.mockRejectedValueOnce(timeoutError);

      const { submitInquiry, error, isError } = useDonationsInquirySubmission();

      const mockSubmissionData: DonationsInquirySubmission = {
        contact_name: 'Test Donor',
        email: 'test@example.com',
        donor_type: 'individual',
        message: 'This is a test message for timeout scenario testing.'
      };

      await submitInquiry(mockSubmissionData);
      await nextTick();

      expect(isError.value).toBe(true);
      expect(error.value).toBe('Request timeout');
    });

    it('should handle malformed responses gracefully', async () => {
      mockSubmitDonationsInquiry.mockResolvedValueOnce(null);

      const { submitInquiry, error, isError } = useDonationsInquirySubmission();

      const mockSubmissionData: DonationsInquirySubmission = {
        contact_name: 'Test Donor',
        email: 'test@example.com',
        donor_type: 'individual',
        message: 'This is a test message for malformed response scenario.'
      };

      await submitInquiry(mockSubmissionData);
      await nextTick();

      expect(isError.value).toBe(true);
      expect(error.value).toContain('response');
    });
  });

  describe('donations-specific business logic', () => {
    it('should handle equipment interest area correctly', async () => {
      const equipmentInquiry = {
        ...mockCorporateSubmission,
        interest_area: 'equipment' as const,
        message: 'Our company would like to fund medical equipment purchases for the clinic.'
      };

      const equipmentResponse = {
        ...mockSuccessResponse,
        donations_inquiry: {
          ...mockSuccessResponse.donations_inquiry!,
          interest_area: 'equipment'
        }
      };

      mockSubmitDonationsInquiry.mockResolvedValueOnce(equipmentResponse);

      const { submitInquiry, response, isSuccess } = useDonationsInquirySubmission();

      await submitInquiry(equipmentInquiry);
      await nextTick();

      expect(isSuccess.value).toBe(true);
      expect(response.value?.donations_inquiry?.interest_area).toBe('equipment');
    });

    it('should handle general support interest area', async () => {
      const generalSupportInquiry = {
        ...mockIndividualSubmission,
        interest_area: 'general-support' as const,
        message: 'I want to support the organization in whatever way is most needed.'
      };

      const generalSupportResponse = {
        ...mockSuccessResponse,
        donations_inquiry: {
          ...mockSuccessResponse.donations_inquiry!,
          interest_area: 'general-support'
        }
      };

      mockSubmitDonationsInquiry.mockResolvedValueOnce(generalSupportResponse);

      const { submitInquiry, response, isSuccess } = useDonationsInquirySubmission();

      await submitInquiry(generalSupportInquiry);
      await nextTick();

      expect(isSuccess.value).toBe(true);
      expect(response.value?.donations_inquiry?.interest_area).toBe('general-support');
    });
  });
});