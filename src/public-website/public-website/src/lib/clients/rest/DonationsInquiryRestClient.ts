import { BaseRestClient } from './BaseRestClient';
import { config } from '../../../environments';
import type { 
  DonationsInquirySubmission,
  InquirySubmissionResponse,
  InquiryGetResponse
} from '../inquiries/types';

export class DonationsInquiryRestClient extends BaseRestClient {
  constructor() {
    super({
      baseUrl: config.domains.services.baseUrl, // Use services domain for inquiries
      timeout: config.domains.services.timeout,
      retryAttempts: config.domains.services.retryAttempts,
    });
  }

  /**
   * Submit donations inquiry
   * Maps to POST /api/inquiries/donations endpoint through Public Gateway
   * Returns backend format with correlation_id for tracking
   */
  async submitDonationsInquiry(submission: DonationsInquirySubmission): Promise<InquirySubmissionResponse> {
    return this.post<InquirySubmissionResponse>('/api/inquiries/donations', submission);
  }

  /**
   * Get donations inquiry by ID
   * Maps to GET /api/inquiries/donations/{id} endpoint through Public Gateway
   * Returns single donations inquiry with metadata
   */
  async getDonationsInquiry(inquiryId: string): Promise<InquiryGetResponse> {
    return this.get<InquiryGetResponse>(`/api/inquiries/donations/${inquiryId}`);
  }
}