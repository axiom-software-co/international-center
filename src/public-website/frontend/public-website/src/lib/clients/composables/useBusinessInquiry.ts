// Business Inquiry Composables - Vue 3 Composition API composables for business inquiry data
// Provides clean interface for Vue components to interact with business inquiry domain

import { ref, computed, watch, type Ref } from 'vue';
import { BusinessInquiryRestClient } from '../rest/BusinessInquiryRestClient';
import type { 
  BusinessInquiry, 
  BusinessInquirySubmission, 
  InquirySubmissionResponse, 
  InquiryGetResponse 
} from '../inquiries/types';

// Create singleton instance
const businessInquiryClient = new BusinessInquiryRestClient();

// Submission composable interface
export interface UseBusinessInquirySubmissionResult {
  isSubmitting: Ref<boolean>;
  error: Ref<string | null>;
  response: Ref<InquirySubmissionResponse | null>;
  isSuccess: Ref<boolean>;
  isError: Ref<boolean>;
  submitInquiry: (submission: BusinessInquirySubmission) => Promise<void>;
  reset: () => void;
}

// Single inquiry fetch interface
export interface UseBusinessInquiryResult {
  inquiry: Ref<BusinessInquiry | null>;
  loading: Ref<boolean>;
  error: Ref<string | null>;
  refetch: () => Promise<void>;
}

// Business inquiry submission composable
export function useBusinessInquirySubmission(): UseBusinessInquirySubmissionResult {
  const isSubmitting = ref(false);
  const error = ref<string | null>(null);
  const response = ref<InquirySubmissionResponse | null>(null);

  const isSuccess = computed(() => response.value?.success === true);
  const isError = computed(() => error.value !== null || response.value?.success === false);

  const submitInquiry = async (submission: BusinessInquirySubmission) => {
    try {
      isSubmitting.value = true;
      error.value = null;
      response.value = null;

      const result = await businessInquiryClient.submitBusinessInquiry(submission);
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
      const errorMessage = err instanceof Error ? err.message : 'Failed to submit business inquiry';
      error.value = errorMessage;
      console.error('Error submitting business inquiry:', err);
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

// Business inquiry fetch composable
export function useBusinessInquiry(inquiryId: Ref<string | null> | string | null): UseBusinessInquiryResult {
  const inquiryIdRef = typeof inquiryId === 'string' ? ref(inquiryId) : inquiryId || ref(null);
  
  const inquiry = ref<BusinessInquiry | null>(null);
  const loading = ref(false);
  const error = ref<string | null>(null);

  const fetchInquiry = async () => {
    if (!inquiryIdRef.value) {
      inquiry.value = null;
      loading.value = false;
      return;
    }

    try {
      loading.value = true;
      error.value = null;

      // Backend returns: { business_inquiry: {...}, correlation_id: string }
      const response: InquiryGetResponse = await businessInquiryClient.getBusinessInquiry(inquiryIdRef.value);
      
      if (response.business_inquiry) {
        inquiry.value = response.business_inquiry;
      } else if (response.error) {
        error.value = response.error;
      } else {
        throw new Error('Failed to fetch business inquiry');
      }
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : 'Failed to fetch business inquiry';
      error.value = errorMessage;
      console.error('Error fetching business inquiry:', err);
      inquiry.value = null;
    } finally {
      loading.value = false;
    }
  };

  // Watch for inquiry ID changes
  watch(inquiryIdRef, fetchInquiry, { immediate: true });

  return {
    inquiry,
    loading,
    error,
    refetch: fetchInquiry,
  };
}