import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { VolunteerInquiryRestClient } from './VolunteerInquiryRestClient';
import type { VolunteerApplication, VolunteerApplicationSubmission, InquirySubmissionResponse } from '../inquiries/types';

// Mock the BaseRestClient
vi.mock('./BaseRestClient');

describe('VolunteerInquiryRestClient', () => {
  let client: VolunteerInquiryRestClient;
  let mockPost: ReturnType<typeof vi.fn>;
  let mockGet: ReturnType<typeof vi.fn>;

  const mockVolunteerApplication: VolunteerApplication = {
    application_id: '123e4567-e89b-12d3-a456-426614174000',
    first_name: 'Maria',
    last_name: 'Rodriguez',
    email: 'maria.rodriguez@email.com',
    phone: '+1-555-123-4567',
    age: 28,
    volunteer_interest: 'patient-support',
    availability: '4-8-hours',
    motivation: 'I have personal experience with chronic illness and want to help others navigate their healthcare journey with compassion and understanding.',
    experience: 'Volunteer experience at local hospice for 2 years, providing emotional support to patients and families.',
    schedule_preferences: 'Weekday afternoons and some weekend mornings work best for my schedule.',
    status: 'new',
    priority: 'medium',
    source: 'website',
    created_at: '2024-03-15T10:00:00Z',
    updated_at: '2024-03-15T10:00:00Z',
    created_by: 'system',
    updated_by: 'system',
    is_deleted: false
  };

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

  const mockSubmissionResponse: InquirySubmissionResponse = {
    volunteer_application: mockVolunteerApplication,
    correlation_id: 'corr-123-456-789',
    success: true,
    message: 'Volunteer application submitted successfully'
  };

  beforeEach(() => {
    mockPost = vi.fn();
    mockGet = vi.fn();
    
    client = new VolunteerInquiryRestClient();
    // Mock the inherited methods from BaseRestClient
    (client as any).post = mockPost;
    (client as any).get = mockGet;
  });

  afterEach(() => {
    vi.clearAllMocks();
  });

  describe('submitVolunteerApplication - API Gateway Compliance', () => {
    it('should submit volunteer application to correct API Gateway endpoint', async () => {
      mockPost.mockResolvedValue(mockSubmissionResponse);

      const result = await client.submitVolunteerApplication(mockPatientSupportSubmission);

      // Test API Gateway specification compliance - should use /api/v1/inquiries/volunteers endpoint
      expect(mockPost).toHaveBeenCalledWith('/api/v1/inquiries/volunteers', mockPatientSupportSubmission);
      expect(result).toEqual(mockSubmissionResponse);
      expect(result.success).toBe(true);
      expect(result.volunteer_application?.volunteer_interest).toBe('patient-support');
    });

    it('should submit patient support volunteer application with complete data', async () => {
      mockPost.mockResolvedValue(mockSubmissionResponse);

      const result = await client.submitVolunteerApplication(mockPatientSupportSubmission);

      expect(mockPost).toHaveBeenCalledWith('/api/v1/inquiries/volunteers', mockPatientSupportSubmission);
      expect(result.volunteer_application?.first_name).toBe('Maria');
      expect(result.volunteer_application?.last_name).toBe('Rodriguez');
      expect(result.volunteer_application?.age).toBe(28);
      expect(result.volunteer_application?.volunteer_interest).toBe('patient-support');
      expect(result.volunteer_application?.availability).toBe('4-8-hours');
      expect(result.volunteer_application?.experience).toContain('hospice');
    });

    it('should submit research support volunteer application', async () => {
      const researchResponse = {
        ...mockSubmissionResponse,
        volunteer_application: {
          ...mockVolunteerApplication,
          application_id: '456e7890-e89b-12d3-a456-426614174001',
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
        }
      };

      mockPost.mockResolvedValue(researchResponse);

      const result = await client.submitVolunteerApplication(mockResearchSupportSubmission);

      expect(mockPost).toHaveBeenCalledWith('/api/v1/inquiries/volunteers', mockResearchSupportSubmission);
      expect(result.volunteer_application?.volunteer_interest).toBe('research-support');
      expect(result.volunteer_application?.availability).toBe('8-16-hours');
      expect(result.volunteer_application?.age).toBe(22);
    });

    it('should handle different volunteer interests correctly', async () => {
      const communityOutreachSubmission = {
        ...mockPatientSupportSubmission,
        first_name: 'Sarah',
        last_name: 'Johnson',
        email: 'sarah.johnson@email.com',
        volunteer_interest: 'community-outreach' as const,
        motivation: 'I want to help raise awareness about preventive healthcare and wellness in underserved communities.'
      };

      const communityResponse = {
        ...mockSubmissionResponse,
        volunteer_application: {
          ...mockVolunteerApplication,
          first_name: 'Sarah',
          last_name: 'Johnson',
          email: 'sarah.johnson@email.com',
          volunteer_interest: 'community-outreach',
          motivation: communityOutreachSubmission.motivation
        }
      };

      mockPost.mockResolvedValue(communityResponse);

      const result = await client.submitVolunteerApplication(communityOutreachSubmission);

      expect(mockPost).toHaveBeenCalledWith('/api/v1/inquiries/volunteers', communityOutreachSubmission);
      expect(result.volunteer_application?.volunteer_interest).toBe('community-outreach');
    });

    it('should handle administrative support volunteer applications', async () => {
      const adminSupportSubmission = {
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
        ...mockSubmissionResponse,
        volunteer_application: {
          ...mockVolunteerApplication,
          first_name: 'Michael',
          last_name: 'Wilson',
          email: 'michael.wilson@email.com',
          volunteer_interest: 'administrative-support',
          availability: '2-4-hours',
          motivation: adminSupportSubmission.motivation,
          experience: adminSupportSubmission.experience
        }
      };

      mockPost.mockResolvedValue(adminResponse);

      const result = await client.submitVolunteerApplication(adminSupportSubmission);

      expect(mockPost).toHaveBeenCalledWith('/api/v1/inquiries/volunteers', adminSupportSubmission);
      expect(result.volunteer_application?.volunteer_interest).toBe('administrative-support');
      expect(result.volunteer_application?.availability).toBe('2-4-hours');
    });

    it('should handle multiple interests volunteer applications', async () => {
      const multipleInterestsSubmission = {
        ...mockPatientSupportSubmission,
        first_name: 'Jennifer',
        last_name: 'Martinez',
        email: 'jennifer.martinez@email.com',
        volunteer_interest: 'multiple' as const,
        availability: '16-hours-plus' as const,
        motivation: 'I have diverse skills and want to contribute across multiple areas - patient support, community outreach, and administrative tasks.'
      };

      const multipleResponse = {
        ...mockSubmissionResponse,
        volunteer_application: {
          ...mockVolunteerApplication,
          first_name: 'Jennifer',
          last_name: 'Martinez',
          email: 'jennifer.martinez@email.com',
          volunteer_interest: 'multiple',
          availability: '16-hours-plus',
          motivation: multipleInterestsSubmission.motivation
        }
      };

      mockPost.mockResolvedValue(multipleResponse);

      const result = await client.submitVolunteerApplication(multipleInterestsSubmission);

      expect(mockPost).toHaveBeenCalledWith('/api/v1/inquiries/volunteers', multipleInterestsSubmission);
      expect(result.volunteer_application?.volunteer_interest).toBe('multiple');
      expect(result.volunteer_application?.availability).toBe('16-hours-plus');
    });

    it('should handle different availability options correctly', async () => {
      const flexibleAvailabilitySubmission = {
        ...mockPatientSupportSubmission,
        availability: 'flexible' as const,
        motivation: 'I have a flexible schedule and can adapt my availability based on program needs and patient requirements.'
      };

      const flexibleResponse = {
        ...mockSubmissionResponse,
        volunteer_application: {
          ...mockVolunteerApplication,
          availability: 'flexible',
          motivation: flexibleAvailabilitySubmission.motivation
        }
      };

      mockPost.mockResolvedValue(flexibleResponse);

      const result = await client.submitVolunteerApplication(flexibleAvailabilitySubmission);

      expect(mockPost).toHaveBeenCalledWith('/api/v1/inquiries/volunteers', flexibleAvailabilitySubmission);
      expect(result.volunteer_application?.availability).toBe('flexible');
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
        ...mockSubmissionResponse,
        volunteer_application: {
          ...mockVolunteerApplication,
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

      mockPost.mockResolvedValue(minimalResponse);

      const result = await client.submitVolunteerApplication(minimalSubmission);

      expect(mockPost).toHaveBeenCalledWith('/api/v1/inquiries/volunteers', minimalSubmission);
      expect(result.volunteer_application?.volunteer_interest).toBe('other');
    });

    it('should handle submission errors from API Gateway', async () => {
      const networkError = new Error('Network connection failed');
      mockPost.mockRejectedValue(networkError);

      await expect(client.submitVolunteerApplication(mockPatientSupportSubmission))
        .rejects.toThrow('Network connection failed');
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

      mockPost.mockResolvedValue(validationErrorResponse);

      const result = await client.submitVolunteerApplication(mockPatientSupportSubmission);

      expect(mockPost).toHaveBeenCalledWith('/api/v1/inquiries/volunteers', mockPatientSupportSubmission);
      expect(result.success).toBe(false);
      expect(result.error).toBe('Validation failed');
      expect(result.validation_errors).toBeDefined();
    });

    it('should handle rate limiting from API Gateway', async () => {
      const rateLimitResponse = {
        error: 'Rate limit exceeded',
        correlation_id: 'corr-rate-limit',
        success: false,
        message: 'Too many volunteer applications. Please try again later.',
        retry_after: 300
      };

      mockPost.mockResolvedValue(rateLimitResponse);

      const result = await client.submitVolunteerApplication(mockPatientSupportSubmission);

      expect(mockPost).toHaveBeenCalledWith('/api/v1/inquiries/volunteers', mockPatientSupportSubmission);
      expect(result.success).toBe(false);
      expect(result.error).toBe('Rate limit exceeded');
      expect(result.retry_after).toBe(300);
    });

    it('should handle age validation constraints from database', async () => {
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

      mockPost.mockResolvedValue(ageValidationResponse);

      const result = await client.submitVolunteerApplication(underageApplication);

      expect(mockPost).toHaveBeenCalledWith('/api/v1/inquiries/volunteers', underageApplication);
      expect(result.success).toBe(false);
      expect(result.error).toBe('Age validation failed');
    });
  });

  describe('getVolunteerApplication - API Gateway Compliance', () => {
    it('should fetch volunteer application by ID using correct API Gateway endpoint', async () => {
      const mockResponse = {
        volunteer_application: mockVolunteerApplication,
        correlation_id: 'corr-get-123'
      };

      mockGet.mockResolvedValue(mockResponse);

      const result = await client.getVolunteerApplication('123e4567-e89b-12d3-a456-426614174000');

      // Test API Gateway specification compliance - should use /api/v1/inquiries/volunteers/{id} endpoint
      expect(mockGet).toHaveBeenCalledWith('/api/v1/inquiries/volunteers/123e4567-e89b-12d3-a456-426614174000');
      expect(result).toEqual(mockResponse);
      expect(result.volunteer_application?.application_id).toBe('123e4567-e89b-12d3-a456-426614174000');
    });

    it('should fetch volunteer application with complete volunteer data', async () => {
      const mockResponse = {
        volunteer_application: mockVolunteerApplication,
        correlation_id: 'corr-get-123'
      };

      mockGet.mockResolvedValue(mockResponse);

      const result = await client.getVolunteerApplication('123e4567-e89b-12d3-a456-426614174000');

      expect(result.volunteer_application?.first_name).toBe('Maria');
      expect(result.volunteer_application?.last_name).toBe('Rodriguez');
      expect(result.volunteer_application?.volunteer_interest).toBe('patient-support');
      expect(result.volunteer_application?.availability).toBe('4-8-hours');
      expect(result.volunteer_application?.status).toBe('new');
      expect(result.volunteer_application?.priority).toBe('medium');
    });

    it('should handle volunteer application not found', async () => {
      const notFoundResponse = {
        error: 'Volunteer application not found',
        correlation_id: 'corr-not-found',
        success: false
      };

      mockGet.mockResolvedValue(notFoundResponse);

      const result = await client.getVolunteerApplication('non-existent-id');

      expect(mockGet).toHaveBeenCalledWith('/api/v1/inquiries/volunteers/non-existent-id');
      expect(result.success).toBe(false);
      expect(result.error).toBe('Volunteer application not found');
    });

    it('should handle fetch errors from API Gateway', async () => {
      const networkError = new Error('Failed to fetch volunteer application');
      mockGet.mockRejectedValue(networkError);

      await expect(client.getVolunteerApplication('123e4567-e89b-12d3-a456-426614174000'))
        .rejects.toThrow('Failed to fetch volunteer application');
    });

    it('should handle different volunteer application statuses', async () => {
      const underReviewApplication = {
        ...mockVolunteerApplication,
        status: 'under-review' as const,
        priority: 'high' as const
      };

      const mockResponse = {
        volunteer_application: underReviewApplication,
        correlation_id: 'corr-under-review'
      };

      mockGet.mockResolvedValue(mockResponse);

      const result = await client.getVolunteerApplication('123e4567-e89b-12d3-a456-426614174000');

      expect(result.volunteer_application?.status).toBe('under-review');
      expect(result.volunteer_application?.priority).toBe('high');
    });

    it('should handle volunteer applications with missing optional fields', async () => {
      const minimalApplication = {
        ...mockVolunteerApplication,
        experience: undefined,
        schedule_preferences: undefined
      };

      const mockResponse = {
        volunteer_application: minimalApplication,
        correlation_id: 'corr-minimal'
      };

      mockGet.mockResolvedValue(mockResponse);

      const result = await client.getVolunteerApplication('123e4567-e89b-12d3-a456-426614174000');

      expect(result.volunteer_application?.experience).toBeUndefined();
      expect(result.volunteer_application?.schedule_preferences).toBeUndefined();
      expect(result.volunteer_application?.motivation).toBeDefined(); // Required field should still be present
    });
  });

  describe('Error Handling and Edge Cases', () => {
    it('should handle malformed responses gracefully', async () => {
      mockPost.mockResolvedValue(null);

      await expect(client.submitVolunteerApplication(mockPatientSupportSubmission))
        .rejects.toThrow();
    });

    it('should handle timeout errors from API Gateway', async () => {
      const timeoutError = new Error('Request timeout');
      mockPost.mockRejectedValue(timeoutError);

      await expect(client.submitVolunteerApplication(mockPatientSupportSubmission))
        .rejects.toThrow('Request timeout');
    });

    it('should validate API Gateway endpoint format', async () => {
      mockPost.mockResolvedValue(mockSubmissionResponse);

      await client.submitVolunteerApplication(mockPatientSupportSubmission);

      // Ensure endpoint follows API Gateway specification exactly
      const callArgs = mockPost.mock.calls[0];
      expect(callArgs[0]).toBe('/api/v1/inquiries/volunteers');
      expect(callArgs[0]).toMatch(/^\/api\/v1\/inquiries\/volunteers$/);
    });

    it('should validate API Gateway GET endpoint format', async () => {
      const mockResponse = {
        volunteer_application: mockVolunteerApplication,
        correlation_id: 'corr-test'
      };

      mockGet.mockResolvedValue(mockResponse);

      await client.getVolunteerApplication('test-id');

      // Ensure GET endpoint follows API Gateway specification exactly
      const callArgs = mockGet.mock.calls[0];
      expect(callArgs[0]).toBe('/api/v1/inquiries/volunteers/test-id');
      expect(callArgs[0]).toMatch(/^\/api\/v1\/inquiries\/volunteers\/[a-zA-Z0-9-]+$/);
    });
  });

  describe('Volunteer-Specific Business Logic', () => {
    it('should handle healthcare professional volunteer applications with priority escalation', async () => {
      const healthcareProfessionalSubmission = {
        ...mockPatientSupportSubmission,
        first_name: 'Dr. Emily',
        last_name: 'Healthcare',
        email: 'dr.emily@hospital.com',
        age: 45,
        volunteer_interest: 'patient-support' as const,
        availability: '8-16-hours' as const,
        motivation: 'As a retired physician, I want to continue helping patients during their treatment and recovery.',
        experience: 'Medical Doctor with 20 years of clinical experience in internal medicine and patient care.'
      };

      const highPriorityResponse = {
        ...mockSubmissionResponse,
        volunteer_application: {
          ...mockVolunteerApplication,
          first_name: 'Dr. Emily',
          last_name: 'Healthcare',
          email: 'dr.emily@hospital.com',
          age: 45,
          volunteer_interest: 'patient-support',
          availability: '8-16-hours',
          motivation: healthcareProfessionalSubmission.motivation,
          experience: healthcareProfessionalSubmission.experience,
          priority: 'high' // Higher priority for healthcare professionals
        }
      };

      mockPost.mockResolvedValue(highPriorityResponse);

      const result = await client.submitVolunteerApplication(healthcareProfessionalSubmission);

      expect(result.volunteer_application?.priority).toBe('high');
      expect(result.volunteer_application?.experience).toContain('Medical Doctor');
    });

    it('should handle research volunteer applications with academic credentials', async () => {
      const academicResearcherSubmission = {
        ...mockResearchSupportSubmission,
        first_name: 'Prof. Research',
        last_name: 'Scientist',
        email: 'prof.scientist@university.edu',
        age: 50,
        experience: 'PhD in Biomedical Sciences, published researcher with 25+ years in clinical research.'
      };

      const academicResponse = {
        ...mockSubmissionResponse,
        volunteer_application: {
          ...mockVolunteerApplication,
          first_name: 'Prof. Research',
          last_name: 'Scientist',
          email: 'prof.scientist@university.edu',
          age: 50,
          volunteer_interest: 'research-support',
          experience: academicResearcherSubmission.experience,
          priority: 'high'
        }
      };

      mockPost.mockResolvedValue(academicResponse);

      const result = await client.submitVolunteerApplication(academicResearcherSubmission);

      expect(result.volunteer_application?.volunteer_interest).toBe('research-support');
      expect(result.volunteer_application?.experience).toContain('PhD');
      expect(result.volunteer_application?.priority).toBe('high');
    });

    it('should handle volunteer applications with cultural and linguistic skills', async () => {
      const culturalVolunteerSubmission = {
        ...mockPatientSupportSubmission,
        first_name: 'Maria',
        last_name: 'Culturale',
        email: 'maria.culturale@community.org',
        volunteer_interest: 'community-outreach' as const,
        motivation: 'I want to help bridge healthcare access gaps in underserved communities and promote health education.',
        experience: 'Community organizer for 10 years, fluent in Spanish and English, experience with health promotion programs.',
        schedule_preferences: 'Evenings and weekends work best for community events and outreach activities.'
      };

      const culturalResponse = {
        ...mockSubmissionResponse,
        volunteer_application: {
          ...mockVolunteerApplication,
          first_name: 'Maria',
          last_name: 'Culturale',
          email: 'maria.culturale@community.org',
          volunteer_interest: 'community-outreach',
          motivation: culturalVolunteerSubmission.motivation,
          experience: culturalVolunteerSubmission.experience,
          schedule_preferences: culturalVolunteerSubmission.schedule_preferences
        }
      };

      mockPost.mockResolvedValue(culturalResponse);

      const result = await client.submitVolunteerApplication(culturalVolunteerSubmission);

      expect(result.volunteer_application?.volunteer_interest).toBe('community-outreach');
      expect(result.volunteer_application?.experience).toContain('fluent in Spanish');
      expect(result.volunteer_application?.schedule_preferences).toContain('community events');
    });
  });
});