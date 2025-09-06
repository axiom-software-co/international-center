import { BaseRestClient } from './BaseRestClient';
import { config } from '../../../environments';
import type { 
  MediaInquirySubmission,
  InquirySubmissionResponse,
  InquiryGetResponse
} from '../inquiries/types';

export class MediaInquiryRestClient extends BaseRestClient {
  constructor() {
    super({
      baseUrl: config.domains.services.baseUrl, // Use services domain for inquiries
      timeout: config.domains.services.timeout,
      retryAttempts: config.domains.services.retryAttempts,
    });
  }

  /**
   * Submit media inquiry
   * Maps to POST /api/v1/inquiries/media endpoint through Public Gateway
   * Returns backend format with correlation_id for tracking
   */
  async submitMediaInquiry(submission: MediaInquirySubmission): Promise<InquirySubmissionResponse> {
    return this.post<InquirySubmissionResponse>('/api/v1/inquiries/media', submission);
  }

  /**
   * Get media inquiry by ID
   * Maps to GET /api/v1/inquiries/media/{id} endpoint through Public Gateway
   * Returns single media inquiry with metadata
   */
  async getMediaInquiry(inquiryId: string): Promise<InquiryGetResponse> {
    return this.get<InquiryGetResponse>(`/api/v1/inquiries/media/${inquiryId}`);
  }
}