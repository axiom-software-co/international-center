// Newsletter Domain Types - Matches actual backend implementation
// Only public-facing types, no admin functionality

// API request/response types - match backend exactly
export interface NewsletterSubscriptionData {
  email: string;
  source?: string;
  contentType?: string;
}

export interface NewsletterSubscriptionResult {
  success: boolean;
  message: string;
  subscription_id?: string;
  confirmation_required: boolean;
}

export interface NewsletterUnsubscribeData {
  email: string;
  subscription_id?: string;
  reason?: string;
}

export interface NewsletterUnsubscribeResult {
  success: boolean;
  message: string;
}
