import { BaseRestClient } from './BaseRestClient';
import { config } from '../../../environments';
import type { 
  VolunteerApplicationSubmission,
  InquirySubmissionResponse,
  InquiryGetResponse
} from '../inquiries/types';

export class VolunteerInquiryRestClient extends BaseRestClient {
  constructor() {
    super({
      baseUrl: config.domains.services.baseUrl, // Use services domain for inquiries
      timeout: config.domains.services.timeout,
      retryAttempts: config.domains.services.retryAttempts,
    });
  }

  /**
   * Submit volunteer application
   * Maps to POST /api/v1/inquiries/volunteers endpoint through Public Gateway
   * Returns backend format with correlation_id for tracking
   */
  async submitVolunteerApplication(submission: VolunteerApplicationSubmission): Promise<InquirySubmissionResponse> {
    return this.post<InquirySubmissionResponse>('/api/v1/inquiries/volunteers', submission);
  }

  /**
   * Get volunteer application by ID
   * Maps to GET /api/v1/inquiries/volunteers/{id} endpoint through Public Gateway
   * Returns single volunteer application with metadata
   */
  async getVolunteerApplication(applicationId: string): Promise<InquiryGetResponse> {
    return this.get<InquiryGetResponse>(`/api/v1/inquiries/volunteers/${applicationId}`);
  }
}