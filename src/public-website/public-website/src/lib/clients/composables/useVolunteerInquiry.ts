// Volunteer Inquiry Composables - Vue 3 Composition API composables for volunteer inquiry data
// Provides clean interface for Vue components to interact with volunteer application domain

import { ref, computed, watch, type Ref } from 'vue';
import { VolunteerInquiryRestClient } from '../rest/VolunteerInquiryRestClient';
import type { 
  VolunteerApplication, 
  VolunteerApplicationSubmission, 
  InquirySubmissionResponse, 
  InquiryGetResponse 
} from '../inquiries/types';

// Create singleton instance
const volunteerInquiryClient = new VolunteerInquiryRestClient();

// Submission composable interface
export interface UseVolunteerInquirySubmissionResult {
  isSubmitting: Ref<boolean>;
  error: Ref<string | null>;
  response: Ref<InquirySubmissionResponse | null>;
  isSuccess: Ref<boolean>;
  isError: Ref<boolean>;
  submitInquiry: (submission: VolunteerApplicationSubmission) => Promise<void>;
  reset: () => void;
}

// Single application fetch interface
export interface UseVolunteerInquiryResult {
  application: Ref<VolunteerApplication | null>;
  loading: Ref<boolean>;
  error: Ref<string | null>;
  refetch: () => Promise<void>;
}

// Volunteer application submission composable
export function useVolunteerInquirySubmission(): UseVolunteerInquirySubmissionResult {
  const isSubmitting = ref(false);
  const error = ref<string | null>(null);
  const response = ref<InquirySubmissionResponse | null>(null);

  const isSuccess = computed(() => response.value?.success === true);
  const isError = computed(() => error.value !== null || response.value?.success === false);

  const submitInquiry = async (submission: VolunteerApplicationSubmission) => {
    try {
      isSubmitting.value = true;
      error.value = null;
      response.value = null;

      const result = await volunteerInquiryClient.submitVolunteerApplication(submission);
      response.value = result;

      // Handle malformed or null responses
      if (!result) {
        error.value = 'Invalid response from server';
        return;
      }

      // Handle backend validation errors
      if (!result.success && result.error) {
        error.value = result.error;
      }
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : 'Failed to submit volunteer application';
      error.value = errorMessage;
      console.error('Error submitting volunteer application:', err);
    } finally {
      isSubmitting.value = false;
    }
  };

  const reset = () => {
    error.value = null;
    response.value = null;
  };

  return {
    isSubmitting,
    error,
    response,
    isSuccess,
    isError,
    submitInquiry,
    reset,
  };
}

// Volunteer application fetch composable
export function useVolunteerInquiry(applicationId: Ref<string | null> | string | null): UseVolunteerInquiryResult {
  const applicationIdRef = typeof applicationId === 'string' ? ref(applicationId) : applicationId || ref(null);
  
  const application = ref<VolunteerApplication | null>(null);
  const loading = ref(false);
  const error = ref<string | null>(null);

  const fetchApplication = async () => {
    if (!applicationIdRef.value) {
      application.value = null;
      loading.value = false;
      return;
    }

    try {
      loading.value = true;
      error.value = null;

      // Backend returns: { volunteer_application: {...}, correlation_id: string }
      const response: InquiryGetResponse = await volunteerInquiryClient.getVolunteerApplication(applicationIdRef.value);
      
      if (response.volunteer_application) {
        application.value = response.volunteer_application;
      } else if (response.error) {
        error.value = response.error;
      } else {
        throw new Error('Failed to fetch volunteer application');
      }
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : 'Failed to fetch volunteer application';
      error.value = errorMessage;
      console.error('Error fetching volunteer application:', err);
      application.value = null;
    } finally {
      loading.value = false;
    }
  };

  // Watch for application ID changes
  watch(applicationIdRef, fetchApplication, { immediate: true });

  return {
    application,
    loading,
    error,
    refetch: fetchApplication,
  };
}