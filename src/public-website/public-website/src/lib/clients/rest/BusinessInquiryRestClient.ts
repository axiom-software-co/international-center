import { BaseRestClient } from './BaseRestClient';
import { config } from '../../../environments';
import type { 
  BusinessInquirySubmission,
  InquirySubmissionResponse,
  InquiryGetResponse
} from '../inquiries/types';

export class BusinessInquiryRestClient extends BaseRestClient {
  constructor() {
    super({
      baseUrl: config.domains.services.baseUrl, // Use services domain for inquiries
      timeout: config.domains.services.timeout,
      retryAttempts: config.domains.services.retryAttempts,
    });
  }

  /**
   * Submit business inquiry
   * Maps to POST /api/inquiries/business endpoint through Public Gateway
   * Returns backend format with correlation_id for tracking
   */
  async submitBusinessInquiry(submission: BusinessInquirySubmission): Promise<InquirySubmissionResponse> {
    return this.post<InquirySubmissionResponse>('/api/inquiries/business', submission);
  }

  /**
   * Get business inquiry by ID
   * Maps to GET /api/inquiries/business/{id} endpoint through Public Gateway
   * Returns single business inquiry with metadata
   */
  async getBusinessInquiry(inquiryId: string): Promise<InquiryGetResponse> {
    return this.get<InquiryGetResponse>(`/api/inquiries/business/${inquiryId}`);
  }
}