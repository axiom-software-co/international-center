// Media Inquiry Composables - Vue 3 Composition API composables for media inquiry data
// Provides clean interface for Vue components to interact with media inquiry domain

import { ref, computed, watch, type Ref } from 'vue';
import { MediaInquiryRestClient } from '../rest/MediaInquiryRestClient';
import type { 
  MediaInquiry, 
  MediaInquirySubmission, 
  InquirySubmissionResponse, 
  InquiryGetResponse 
} from '../inquiries/types';

// Create singleton instance
const mediaInquiryClient = new MediaInquiryRestClient();

// Submission composable interface
export interface UseMediaInquirySubmissionResult {
  isSubmitting: Ref<boolean>;
  error: Ref<string | null>;
  response: Ref<InquirySubmissionResponse | null>;
  isSuccess: Ref<boolean>;
  isError: Ref<boolean>;
  submitInquiry: (submission: MediaInquirySubmission) => Promise<void>;
  reset: () => void;
}

// Single inquiry fetch interface
export interface UseMediaInquiryResult {
  inquiry: Ref<MediaInquiry | null>;
  loading: Ref<boolean>;
  error: Ref<string | null>;
  refetch: () => Promise<void>;
}

// Media inquiry submission composable
export function useMediaInquirySubmission(): UseMediaInquirySubmissionResult {
  const isSubmitting = ref(false);
  const error = ref<string | null>(null);
  const response = ref<InquirySubmissionResponse | null>(null);

  const isSuccess = computed(() => response.value?.success === true);
  const isError = computed(() => error.value !== null || response.value?.success === false);

  const submitInquiry = async (submission: MediaInquirySubmission) => {
    try {
      isSubmitting.value = true;
      error.value = null;
      response.value = null;

      const result = await mediaInquiryClient.submitMediaInquiry(submission);
      response.value = result;

      // Handle backend validation errors
      if (!result.success && result.error) {
        error.value = result.error;
      }
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : 'Failed to submit media inquiry';
      error.value = errorMessage;
      console.error('Error submitting media inquiry:', err);

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

// Media inquiry fetch composable
export function useMediaInquiry(inquiryId: Ref<string | null> | string | null): UseMediaInquiryResult {
  const inquiryIdRef = typeof inquiryId === 'string' ? ref(inquiryId) : inquiryId || ref(null);
  
  const inquiry = ref<MediaInquiry | null>(null);
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

      // Backend returns: { media_inquiry: {...}, correlation_id: string }
      const response: InquiryGetResponse = await mediaInquiryClient.getMediaInquiry(inquiryIdRef.value);
      
      if (response.media_inquiry) {
        inquiry.value = response.media_inquiry;
      } else if (response.error) {
        error.value = response.error;
      } else {
        throw new Error('Failed to fetch media inquiry');
      }
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : 'Failed to fetch media inquiry';
      error.value = errorMessage;
      console.error('Error fetching media inquiry:', err);
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