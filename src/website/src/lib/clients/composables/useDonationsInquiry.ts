// Donations Inquiry Composables - Vue 3 Composition API composables for donations inquiry data
// Provides clean interface for Vue components to interact with donations inquiry domain

import { ref, computed, watch, type Ref } from 'vue';
import { DonationsInquiryRestClient } from '../rest/DonationsInquiryRestClient';
import type { 
  DonationsInquiry, 
  DonationsInquirySubmission, 
  InquirySubmissionResponse, 
  InquiryGetResponse 
} from '../inquiries/types';

// Create singleton instance
const donationsInquiryClient = new DonationsInquiryRestClient();

// Submission composable interface
export interface UseDonationsInquirySubmissionResult {
  isSubmitting: Ref<boolean>;
  error: Ref<string | null>;
  response: Ref<InquirySubmissionResponse | null>;
  isSuccess: Ref<boolean>;
  isError: Ref<boolean>;
  submitInquiry: (submission: DonationsInquirySubmission) => Promise<void>;
  reset: () => void;
}

// Single inquiry fetch interface
export interface UseDonationsInquiryResult {
  inquiry: Ref<DonationsInquiry | null>;
  loading: Ref<boolean>;
  error: Ref<string | null>;
  refetch: () => Promise<void>;
}

// Donations inquiry submission composable
export function useDonationsInquirySubmission(): UseDonationsInquirySubmissionResult {
  const isSubmitting = ref(false);
  const error = ref<string | null>(null);
  const response = ref<InquirySubmissionResponse | null>(null);

  const isSuccess = computed(() => response.value?.success === true);
  const isError = computed(() => error.value !== null || response.value?.success === false);

  const submitInquiry = async (submission: DonationsInquirySubmission) => {
    try {
      isSubmitting.value = true;
      error.value = null;
      response.value = null;

      const result = await donationsInquiryClient.submitDonationsInquiry(submission);
      response.value = result;

      // Handle backend validation errors
      if (!result.success && result.error) {
        error.value = result.error;
      }
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : 'Failed to submit donations inquiry';
      error.value = errorMessage;
      console.error('Error submitting donations inquiry:', err);

      // Check if response is malformed
      if (!response.value) {
        error.value = 'Invalid response from server';
      }
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

// Donations inquiry fetch composable
export function useDonationsInquiry(inquiryId: Ref<string | null> | string | null): UseDonationsInquiryResult {
  const inquiryIdRef = typeof inquiryId === 'string' ? ref(inquiryId) : inquiryId || ref(null);
  
  const inquiry = ref<DonationsInquiry | null>(null);
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

      // Backend returns: { donations_inquiry: {...}, correlation_id: string }
      const response: InquiryGetResponse = await donationsInquiryClient.getDonationsInquiry(inquiryIdRef.value);
      
      if (response.donations_inquiry) {
        inquiry.value = response.donations_inquiry;
      } else if (response.error) {
        error.value = response.error;
      } else {
        throw new Error('Failed to fetch donations inquiry');
      }
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : 'Failed to fetch donations inquiry';
      error.value = errorMessage;
      console.error('Error fetching donations inquiry:', err);
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