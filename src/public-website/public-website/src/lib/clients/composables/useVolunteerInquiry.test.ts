import { describe, it, expect, vi, beforeEach } from 'vitest';
import { ref, nextTick } from 'vue';
import { useVolunteerInquiry, useVolunteerInquirySubmission } from './useVolunteerInquiry';
import { VolunteerInquiryRestClient } from '../rest/VolunteerInquiryRestClient';
import { RestError } from '../rest/BaseRestClient';
import type { VolunteerApplicationSubmission, InquirySubmissionResponse } from '../inquiries/types';

// Mock the VolunteerInquiryRestClient with hoisted functions
const {
  mockSubmitVolunteerApplication,
  mockGetVolunteerApplication,
  MockedVolunteerInquiryRestClient
} = vi.hoisted(() => {
  const mockSubmit = vi.fn();
  const mockGet = vi.fn();
  
  return {
    mockSubmitVolunteerApplication: mockSubmit,
    mockGetVolunteerApplication: mockGet,
    MockedVolunteerInquiryRestClient: vi.fn().mockImplementation(() => ({
      submitVolunteerApplication: mockSubmit,
      getVolunteerApplication: mockGet,
    }))
  };
});

vi.mock('../rest/VolunteerInquiryRestClient', () => ({
  VolunteerInquiryRestClient: MockedVolunteerInquiryRestClient
}));

describe('useVolunteerInquiry composables', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  describe('useVolunteerInquirySubmission', () => {
    const mockPatientSupportSubmission: VolunteerApplicationSubmission = {
      first_name: 'Maria',
      last_name: 'Rodriguez',
      email: 'maria.rodriguez@email.com',
      phone: '+1-555-123-4567',
      age: 28,
      volunteer_interest: 'patient-support',
      availability: '4-8-hours',
      motivation: 'I have personal experience with chronic illness and want to help others navigate their healthcare journey with compassion and understanding.',
      experience: 'Volunteer experience at local hospice for 2 years, providing emotional support to patients and families.',
      schedule_preferences: 'Weekday afternoons and some weekend mornings work best for my schedule.'
    };

    const mockResearchSupportSubmission: VolunteerApplicationSubmission = {
      first_name: 'David',
      last_name: 'Chen',
      email: 'david.chen@university.edu',
      phone: '+1-555-987-6543',
      age: 22,
      volunteer_interest: 'research-support',
      availability: '8-16-hours',
      motivation: 'As a pre-med student, I want to contribute to clinical research that advances patient care and treatment outcomes.',
      experience: 'Research assistant in university biochemistry lab, experience with data collection and analysis.',
      schedule_preferences: 'Flexible schedule, available most weekdays and weekends.'
    };

    const mockSuccessResponse: InquirySubmissionResponse = {
      volunteer_application: {
        application_id: '123e4567-e89b-12d3-a456-426614174000',
        first_name: 'Maria',
        last_name: 'Rodriguez',
        email: 'maria.rodriguez@email.com',
        phone: '+1-555-123-4567',
        age: 28,
        volunteer_interest: 'patient-support',
        availability: '4-8-hours',
        motivation: mockPatientSupportSubmission.motivation,
        experience: mockPatientSupportSubmission.experience,
        schedule_preferences: mockPatientSupportSubmission.schedule_preferences,
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
      message: 'Volunteer application submitted successfully'
    };

    it('should initialize with correct default values', () => {
      const { isSubmitting, error, response, isSuccess, isError } = useVolunteerInquirySubmission();

      expect(isSubmitting.value).toBe(false);
      expect(error.value).toBe(null);
      expect(response.value).toBe(null);
      expect(isSuccess.value).toBe(false);
      expect(isError.value).toBe(false);
    });

    it('should submit volunteer application successfully for patient support', async () => {
      mockSubmitVolunteerApplication.mockResolvedValueOnce(mockSuccessResponse);

      const { submitInquiry, isSubmitting, response, isSuccess, error } = useVolunteerInquirySubmission();

      expect(isSubmitting.value).toBe(false);

      const submissionPromise = submitInquiry(mockPatientSupportSubmission);
      
      // Should be submitting during the request
      expect(isSubmitting.value).toBe(true);

      await submissionPromise;
      await nextTick();

      // Should complete successfully
      expect(isSubmitting.value).toBe(false);
      expect(isSuccess.value).toBe(true);
      expect(error.value).toBe(null);
      expect(response.value).toEqual(mockSuccessResponse);
      expect(response.value?.volunteer_application?.volunteer_interest).toBe('patient-support');
      expect(response.value?.volunteer_application?.age).toBe(28);
      expect(mockSubmitVolunteerApplication).toHaveBeenCalledWith(mockPatientSupportSubmission);
    });

    it('should handle research support volunteer application correctly', async () => {
      const researchResponse = {
        ...mockSuccessResponse,
        volunteer_application: {
          ...mockSuccessResponse.volunteer_application!,
          application_id: '456e7890-e89b-12d3-a456-426614174001',
          first_name: 'David',
          last_name: 'Chen',
          email: 'david.chen@university.edu',
          phone: '+1-555-987-6543',
          age: 22,
          volunteer_interest: 'research-support',
          availability: '8-16-hours',
          motivation: mockResearchSupportSubmission.motivation,
          experience: mockResearchSupportSubmission.experience,
          schedule_preferences: mockResearchSupportSubmission.schedule_preferences
        }
      };

      mockSubmitVolunteerApplication.mockResolvedValueOnce(researchResponse);

      const { submitInquiry, response, isSuccess } = useVolunteerInquirySubmission();

      await submitInquiry(mockResearchSupportSubmission);
      await nextTick();

      expect(isSuccess.value).toBe(true);
      expect(response.value?.volunteer_application?.volunteer_interest).toBe('research-support');
      expect(response.value?.volunteer_application?.availability).toBe('8-16-hours');
      expect(response.value?.volunteer_application?.age).toBe(22);
    });

    it('should handle different volunteer interests correctly', async () => {
      const communityOutreachApplication = {
        ...mockPatientSupportSubmission,
        first_name: 'Sarah',
        last_name: 'Johnson',
        email: 'sarah.johnson@email.com',
        volunteer_interest: 'community-outreach' as const,
        motivation: 'I want to help raise awareness about preventive healthcare and wellness in underserved communities.'
      };

      const communityResponse = {
        ...mockSuccessResponse,
        volunteer_application: {
          ...mockSuccessResponse.volunteer_application!,
          first_name: 'Sarah',
          last_name: 'Johnson',
          email: 'sarah.johnson@email.com',
          volunteer_interest: 'community-outreach',
          motivation: communityOutreachApplication.motivation
        }
      };

      mockSubmitVolunteerApplication.mockResolvedValueOnce(communityResponse);

      const { submitInquiry, response, isSuccess } = useVolunteerInquirySubmission();

      await submitInquiry(communityOutreachApplication);
      await nextTick();

      expect(isSuccess.value).toBe(true);
      expect(response.value?.volunteer_application?.volunteer_interest).toBe('community-outreach');
    });

    it('should handle administrative support volunteer correctly', async () => {
      const adminSupportApplication = {
        ...mockPatientSupportSubmission,
        first_name: 'Michael',
        last_name: 'Wilson',
        email: 'michael.wilson@email.com',
        volunteer_interest: 'administrative-support' as const,
        availability: '2-4-hours' as const,
        motivation: 'I have office management experience and want to help with clerical tasks to support patient care operations.',
        experience: 'Administrative assistant experience in healthcare setting for 5 years.'
      };

      const adminResponse = {
        ...mockSuccessResponse,
        volunteer_application: {
          ...mockSuccessResponse.volunteer_application!,
          first_name: 'Michael',
          last_name: 'Wilson',
          email: 'michael.wilson@email.com',
          volunteer_interest: 'administrative-support',
          availability: '2-4-hours',
          motivation: adminSupportApplication.motivation,
          experience: adminSupportApplication.experience
        }
      };

      mockSubmitVolunteerApplication.mockResolvedValueOnce(adminResponse);

      const { submitInquiry, response, isSuccess } = useVolunteerInquirySubmission();

      await submitInquiry(adminSupportApplication);
      await nextTick();

      expect(isSuccess.value).toBe(true);
      expect(response.value?.volunteer_application?.volunteer_interest).toBe('administrative-support');
      expect(response.value?.volunteer_application?.availability).toBe('2-4-hours');
    });

    it('should handle multiple interests correctly', async () => {
      const multipleInterestsApplication = {
        ...mockPatientSupportSubmission,
        first_name: 'Jennifer',
        last_name: 'Martinez',
        email: 'jennifer.martinez@email.com',
        volunteer_interest: 'multiple' as const,
        availability: '16-hours-plus' as const,
        motivation: 'I have diverse skills and want to contribute across multiple areas - patient support, community outreach, and administrative tasks.'
      };

      const multipleResponse = {
        ...mockSuccessResponse,
        volunteer_application: {
          ...mockSuccessResponse.volunteer_application!,
          first_name: 'Jennifer',
          last_name: 'Martinez',
          email: 'jennifer.martinez@email.com',
          volunteer_interest: 'multiple',
          availability: '16-hours-plus',
          motivation: multipleInterestsApplication.motivation
        }
      };

      mockSubmitVolunteerApplication.mockResolvedValueOnce(multipleResponse);

      const { submitInquiry, response, isSuccess } = useVolunteerInquirySubmission();

      await submitInquiry(multipleInterestsApplication);
      await nextTick();

      expect(isSuccess.value).toBe(true);
      expect(response.value?.volunteer_application?.volunteer_interest).toBe('multiple');
      expect(response.value?.volunteer_application?.availability).toBe('16-hours-plus');
    });

    it('should handle different availability options correctly', async () => {
      const flexibleAvailabilityApplication = {
        ...mockPatientSupportSubmission,
        availability: 'flexible' as const,
        motivation: 'I have a flexible schedule and can adapt my availability based on program needs and patient requirements.'
      };

      const flexibleResponse = {
        ...mockSuccessResponse,
        volunteer_application: {
          ...mockSuccessResponse.volunteer_application!,
          availability: 'flexible',
          motivation: flexibleAvailabilityApplication.motivation
        }
      };

      mockSubmitVolunteerApplication.mockResolvedValueOnce(flexibleResponse);

      const { submitInquiry, response, isSuccess } = useVolunteerInquirySubmission();

      await submitInquiry(flexibleAvailabilityApplication);
      await nextTick();

      expect(isSuccess.value).toBe(true);
      expect(response.value?.volunteer_application?.availability).toBe('flexible');
    });

    it('should handle submission errors', async () => {
      const errorMessage = 'Network connection failed';
      mockSubmitVolunteerApplication.mockRejectedValueOnce(new Error(errorMessage));

      const { submitInquiry, isSubmitting, error, isError, isSuccess } = useVolunteerInquirySubmission();

      const submissionPromise = submitInquiry(mockPatientSupportSubmission);
      
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
          'Age must be 18 or older',
          'Motivation must be at least 20 characters',
          'Phone number is required for volunteer applications'
        ],
        correlation_id: 'corr-validation-error',
        success: false,
        message: 'Please correct the following errors'
      };

      mockSubmitVolunteerApplication.mockResolvedValueOnce(validationErrorResponse);

      const { submitInquiry, error, isError, response, isSuccess } = useVolunteerInquirySubmission();

      await submitInquiry(mockPatientSupportSubmission);
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
        message: 'Too many volunteer applications. Please try again later.',
        retry_after: 300
      };

      mockSubmitVolunteerApplication.mockResolvedValueOnce(rateLimitResponse);

      const { submitInquiry, error, isError, response } = useVolunteerInquirySubmission();

      await submitInquiry(mockPatientSupportSubmission);
      await nextTick();

      expect(isError.value).toBe(true);
      expect(error.value).toBe('Rate limit exceeded');
      expect(response.value?.retry_after).toBe(300);
    });

    it('should reset state correctly', async () => {
      mockSubmitVolunteerApplication.mockResolvedValueOnce(mockSuccessResponse);

      const { submitInquiry, error, response, reset } = useVolunteerInquirySubmission();

      // First submission
      await submitInquiry(mockPatientSupportSubmission);
      await nextTick();

      expect(response.value).toEqual(mockSuccessResponse);

      // Reset state
      reset();

      expect(error.value).toBe(null);
      expect(response.value).toBe(null);
    });

    it('should handle minimal application without optional fields', async () => {
      const minimalSubmission = {
        first_name: 'John',
        last_name: 'Doe',
        email: 'john.doe@email.com',
        phone: '+1-555-000-0000',
        age: 25,
        volunteer_interest: 'other' as const,
        availability: '2-4-hours' as const,
        motivation: 'I want to help patients and contribute to the healthcare community in any way possible.'
      };

      const minimalResponse = {
        ...mockSuccessResponse,
        volunteer_application: {
          ...mockSuccessResponse.volunteer_application!,
          first_name: 'John',
          last_name: 'Doe',
          email: 'john.doe@email.com',
          phone: '+1-555-000-0000',
          age: 25,
          volunteer_interest: 'other',
          availability: '2-4-hours',
          motivation: minimalSubmission.motivation,
          experience: undefined,
          schedule_preferences: undefined
        }
      };

      mockSubmitVolunteerApplication.mockResolvedValueOnce(minimalResponse);

      const { submitInquiry, response, isSuccess } = useVolunteerInquirySubmission();

      await submitInquiry(minimalSubmission);
      await nextTick();

      expect(isSuccess.value).toBe(true);
      expect(response.value?.volunteer_application?.volunteer_interest).toBe('other');
    });

    it('should handle age verification constraints', async () => {
      const underageApplication = {
        ...mockPatientSupportSubmission,
        age: 17,
        first_name: 'Young',
        last_name: 'Applicant'
      };

      const ageValidationResponse = {
        error: 'Age validation failed',
        validation_errors: [
          'Applicant must be at least 18 years old to volunteer'
        ],
        correlation_id: 'corr-age-validation',
        success: false,
        message: 'Age requirement not met'
      };

      mockSubmitVolunteerApplication.mockResolvedValueOnce(ageValidationResponse);

      const { submitInquiry, error, isError } = useVolunteerInquirySubmission();

      await submitInquiry(underageApplication);
      await nextTick();

      expect(isError.value).toBe(true);
      expect(error.value).toBe('Age validation failed');
    });
  });

  describe('useVolunteerInquiry', () => {
    const mockVolunteerApplication = {
      application_id: '123e4567-e89b-12d3-a456-426614174000',
      first_name: 'Maria',
      last_name: 'Rodriguez',
      email: 'maria.rodriguez@email.com',
      phone: '+1-555-123-4567',
      age: 28,
      volunteer_interest: 'patient-support' as const,
      availability: '4-8-hours' as const,
      motivation: 'I have personal experience with chronic illness and want to help others navigate their healthcare journey.',
      experience: 'Volunteer experience at local hospice for 2 years.',
      schedule_preferences: 'Weekday afternoons work best.',
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
      const { application, loading, error } = useVolunteerInquiry(ref(null));

      expect(application.value).toBe(null);
      expect(loading.value).toBe(false);
      expect(error.value).toBe(null);
    });

    it('should fetch volunteer application by ID', async () => {
      const mockResponse = {
        volunteer_application: mockVolunteerApplication,
        correlation_id: 'corr-get-123'
      };

      mockGetVolunteerApplication.mockResolvedValueOnce(mockResponse);

      const applicationId = ref('123e4567-e89b-12d3-a456-426614174000');
      const { application, loading, error } = useVolunteerInquiry(applicationId);

      await nextTick();

      expect(application.value).toEqual(mockVolunteerApplication);
      expect(loading.value).toBe(false);
      expect(error.value).toBe(null);
      expect(mockGetVolunteerApplication).toHaveBeenCalledWith('123e4567-e89b-12d3-a456-426614174000');
    });

    it('should handle application not found', async () => {
      const notFoundResponse = {
        error: 'Volunteer application not found',
        correlation_id: 'corr-not-found',
        success: false
      };

      mockGetVolunteerApplication.mockResolvedValueOnce(notFoundResponse);

      const applicationId = ref('non-existent-id');
      const { application, loading, error } = useVolunteerInquiry(applicationId);

      await nextTick();

      expect(application.value).toBe(null);
      expect(loading.value).toBe(false);
      expect(error.value).toBe('Volunteer application not found');
    });

    it('should handle fetch errors', async () => {
      const errorMessage = 'Failed to fetch volunteer application';
      mockGetVolunteerApplication.mockRejectedValueOnce(new Error(errorMessage));

      const applicationId = ref('123e4567-e89b-12d3-a456-426614174000');
      const { application, error } = useVolunteerInquiry(applicationId);

      await nextTick();

      expect(application.value).toBe(null);
      expect(error.value).toBe(errorMessage);
    });

    it('should refetch when application ID changes', async () => {
      const mockResponse1 = {
        volunteer_application: { ...mockVolunteerApplication, application_id: 'id-1' },
        correlation_id: 'corr-1'
      };

      const mockResponse2 = {
        volunteer_application: { ...mockVolunteerApplication, application_id: 'id-2', first_name: 'Updated' },
        correlation_id: 'corr-2'
      };

      mockGetVolunteerApplication.mockResolvedValueOnce(mockResponse1);

      const applicationId = ref('id-1');
      const { application } = useVolunteerInquiry(applicationId);

      await nextTick();
      expect(application.value?.application_id).toBe('id-1');

      // Change ID and setup second response
      mockGetVolunteerApplication.mockResolvedValueOnce(mockResponse2);
      applicationId.value = 'id-2';

      await nextTick();
      expect(application.value?.application_id).toBe('id-2');
      expect(application.value?.first_name).toBe('Updated');
      expect(mockGetVolunteerApplication).toHaveBeenCalledTimes(2);
    });

    it('should handle loading states correctly', async () => {
      const mockResponse = {
        volunteer_application: mockVolunteerApplication,
        correlation_id: 'corr-loading-test'
      };

      let resolvePromise: (value: any) => void;
      const pendingPromise = new Promise(resolve => {
        resolvePromise = resolve;
      });
      mockGetVolunteerApplication.mockReturnValueOnce(pendingPromise);

      const applicationId = ref('123e4567-e89b-12d3-a456-426614174000');
      const { application, loading, error } = useVolunteerInquiry(applicationId);

      await nextTick();

      // Should be loading
      expect(loading.value).toBe(true);
      expect(application.value).toBe(null);
      expect(error.value).toBe(null);

      // Resolve the promise
      resolvePromise!(mockResponse);
      await nextTick();

      // Should complete loading
      expect(loading.value).toBe(false);
      expect(application.value).toEqual(mockVolunteerApplication);
      expect(error.value).toBe(null);
    });
  });

  describe('error handling edge cases', () => {
    it('should handle network timeouts', async () => {
      const timeoutError = new Error('Request timeout');
      mockSubmitVolunteerApplication.mockRejectedValueOnce(timeoutError);

      const { submitInquiry, error, isError } = useVolunteerInquirySubmission();

      const mockSubmissionData: VolunteerApplicationSubmission = {
        first_name: 'Test',
        last_name: 'User',
        email: 'test@example.com',
        phone: '+1-555-000-0000',
        age: 25,
        volunteer_interest: 'patient-support',
        availability: '2-4-hours',
        motivation: 'This is a test motivation message for timeout scenario testing purposes.'
      };

      await submitInquiry(mockSubmissionData);
      await nextTick();

      expect(isError.value).toBe(true);
      expect(error.value).toBe('Request timeout');
    });

    it('should handle malformed responses gracefully', async () => {
      mockSubmitVolunteerApplication.mockResolvedValueOnce(null);

      const { submitInquiry, error, isError } = useVolunteerInquirySubmission();

      const mockSubmissionData: VolunteerApplicationSubmission = {
        first_name: 'Test',
        last_name: 'User',
        email: 'test@example.com',
        phone: '+1-555-000-0000',
        age: 25,
        volunteer_interest: 'administrative-support',
        availability: '4-8-hours',
        motivation: 'This is a test motivation message for malformed response scenario testing.'
      };

      await submitInquiry(mockSubmissionData);
      await nextTick();

      expect(isError.value).toBe(true);
      expect(error.value).toContain('response');
    });
  });

  describe('volunteer-specific business logic', () => {
    it('should handle patient support interest with healthcare experience', async () => {
      const healthcareVolunteer = {
        first_name: 'Nurse',
        last_name: 'Helper',
        email: 'nurse.helper@email.com',
        phone: '+1-555-111-2222',
        age: 35,
        volunteer_interest: 'patient-support' as const,
        availability: '8-16-hours' as const,
        motivation: 'As a retired nurse, I want to continue helping patients during their treatment journey and recovery process.',
        experience: 'Registered nurse with 15 years of clinical experience in oncology and patient care.',
        schedule_preferences: 'Available weekdays and some weekends, prefer morning shifts.'
      };

      const healthcareResponse = {
        correlation_id: 'corr-healthcare',
        success: true,
        message: 'Healthcare volunteer application submitted successfully',
        volunteer_application: {
          application_id: '789e0123-e89b-12d3-a456-426614174003',
          ...healthcareVolunteer,
          status: 'new' as const,
          priority: 'high' as const, // High priority for healthcare experience
          source: 'website',
          created_at: '2024-03-15T10:00:00Z',
          updated_at: '2024-03-15T10:00:00Z',
          created_by: 'system',
          updated_by: 'system',
          is_deleted: false
        }
      };

      mockSubmitVolunteerApplication.mockResolvedValueOnce(healthcareResponse);

      const { submitInquiry, response, isSuccess } = useVolunteerInquirySubmission();

      await submitInquiry(healthcareVolunteer);
      await nextTick();

      expect(isSuccess.value).toBe(true);
      expect(response.value?.volunteer_application?.volunteer_interest).toBe('patient-support');
      expect(response.value?.volunteer_application?.priority).toBe('high');
      expect(response.value?.volunteer_application?.experience).toContain('nurse');
    });

    it('should handle research support with academic background', async () => {
      const academicVolunteer = {
        first_name: 'Professor',
        last_name: 'Science',
        email: 'prof.science@university.edu',
        phone: '+1-555-333-4444',
        age: 45,
        volunteer_interest: 'research-support' as const,
        availability: '16-hours-plus' as const,
        motivation: 'I have extensive research experience and want to contribute to clinical studies that advance medical knowledge.',
        experience: 'PhD in Biomedical Sciences, published researcher with 20+ years in clinical research.',
        schedule_preferences: 'Very flexible schedule, can adapt to research project timelines and deadlines.'
      };

      const academicResponse = {
        correlation_id: 'corr-academic',
        success: true,
        message: 'Academic volunteer application submitted successfully',
        volunteer_application: {
          application_id: '456e7890-e89b-12d3-a456-426614174004',
          ...academicVolunteer,
          status: 'new' as const,
          priority: 'high' as const,
          source: 'website',
          created_at: '2024-03-15T10:00:00Z',
          updated_at: '2024-03-15T10:00:00Z',
          created_by: 'system',
          updated_by: 'system',
          is_deleted: false
        }
      };

      mockSubmitVolunteerApplication.mockResolvedValueOnce(academicResponse);

      const { submitInquiry, response, isSuccess } = useVolunteerInquirySubmission();

      await submitInquiry(academicVolunteer);
      await nextTick();

      expect(isSuccess.value).toBe(true);
      expect(response.value?.volunteer_application?.volunteer_interest).toBe('research-support');
      expect(response.value?.volunteer_application?.availability).toBe('16-hours-plus');
      expect(response.value?.volunteer_application?.experience).toContain('PhD');
    });

    it('should handle community outreach with cultural competency', async () => {
      const communityVolunteer = {
        first_name: 'Community',
        last_name: 'Leader',
        email: 'community.leader@local.org',
        phone: '+1-555-555-6666',
        age: 38,
        volunteer_interest: 'community-outreach' as const,
        availability: '8-16-hours' as const,
        motivation: 'I want to help bridge healthcare access gaps in underserved communities and promote health education.',
        experience: 'Community organizer for 10 years, fluent in Spanish and English, experience with health promotion programs.',
        schedule_preferences: 'Evenings and weekends work best for community events and outreach activities.'
      };

      const communityResponse = {
        correlation_id: 'corr-community',
        success: true,
        message: 'Community outreach volunteer application submitted successfully',
        volunteer_application: {
          application_id: '321e6543-e89b-12d3-a456-426614174005',
          ...communityVolunteer,
          status: 'new' as const,
          priority: 'medium' as const,
          source: 'website',
          created_at: '2024-03-15T10:00:00Z',
          updated_at: '2024-03-15T10:00:00Z',
          created_by: 'system',
          updated_by: 'system',
          is_deleted: false
        }
      };

      mockSubmitVolunteerApplication.mockResolvedValueOnce(communityResponse);

      const { submitInquiry, response, isSuccess } = useVolunteerInquirySubmission();

      await submitInquiry(communityVolunteer);
      await nextTick();

      expect(isSuccess.value).toBe(true);
      expect(response.value?.volunteer_application?.volunteer_interest).toBe('community-outreach');
      expect(response.value?.volunteer_application?.experience).toContain('Community organizer');
    });
  });
});