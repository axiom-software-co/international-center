// Inquiry Types - TypeScript interfaces matching database schemas
// Provides type safety for three-domain inquiry architecture (business/donations/media)

// Enum types matching database constraints
export type InquiryStatus = 'new' | 'acknowledged' | 'in_progress' | 'resolved' | 'closed';
export type InquiryPriority = 'low' | 'medium' | 'high' | 'urgent';

// Business inquiry specific enums
export type BusinessInquiryType = 'partnership' | 'licensing' | 'research' | 'technology' | 'regulatory' | 'other';

// Donations inquiry specific enums
export type DonorType = 'individual' | 'corporate' | 'foundation' | 'estate' | 'other';
export type InterestArea = 'clinic-development' | 'research-funding' | 'patient-care' | 'equipment' | 'general-support' | 'other';
export type AmountRange = 'under-1000' | '1000-5000' | '5000-25000' | '25000-100000' | 'over-100000' | 'undisclosed';
export type DonationFrequency = 'one-time' | 'monthly' | 'quarterly' | 'annually' | 'other';

// Media inquiry specific enums  
export type MediaType = 'print' | 'digital' | 'television' | 'radio' | 'podcast' | 'medical-journal' | 'other';
export type MediaUrgency = 'low' | 'medium' | 'high';

// Volunteer application specific enums
export type VolunteerInterest = 'patient-support' | 'community-outreach' | 'research-support' | 'administrative-support' | 'multiple' | 'other';
export type VolunteerAvailability = '2-4-hours' | '4-8-hours' | '8-16-hours' | '16-hours-plus' | 'flexible';
export type VolunteerStatus = 'new' | 'under-review' | 'interview-scheduled' | 'background-check' | 'approved' | 'declined' | 'withdrawn';

// Base inquiry interface - shared fields across all inquiry types
export interface BaseInquiry {
  // Primary key
  inquiry_id: string;
  
  // Contact information
  contact_name: string;
  email: string;
  phone?: string;
  
  // Status management
  status: InquiryStatus;
  priority: InquiryPriority;
  
  // Metadata
  source?: string;
  ip_address?: string;
  user_agent?: string;
  
  // Audit fields
  created_at: string;
  updated_at: string;
  created_by: string;
  updated_by: string;
  is_deleted: boolean;
  deleted_at?: string;
}

// Business inquiries interface
export interface BusinessInquiry extends BaseInquiry {
  // Business-specific required fields
  organization_name: string;
  title: string;
  inquiry_type: BusinessInquiryType;
  message: string; // 20-1500 characters
  
  // Business-specific optional fields
  industry?: string;
}

// Donations inquiries interface
export interface DonationsInquiry extends BaseInquiry {
  // Donations-specific required fields
  donor_type: DonorType;
  message: string; // 20-2000 characters
  
  // Donations-specific optional fields
  organization?: string; // Optional for individual donors
  interest_area?: InterestArea;
  preferred_amount_range?: AmountRange;
  donation_frequency?: DonationFrequency;
}

// Media inquiries interface
export interface MediaInquiry extends BaseInquiry {
  // Media-specific required fields
  outlet: string;
  title: string;
  phone: string; // Required for media (unlike other domains)
  subject: string; // 20-1500 characters (not 'message')
  urgency: MediaUrgency;
  
  // Media-specific optional fields
  media_type?: MediaType;
  deadline?: string; // ISO date format
}

// Volunteer applications interface - uses application_id instead of inquiry_id and different status enum
export interface VolunteerApplication {
  // Primary key
  application_id: string;
  
  // Personal information
  first_name: string;
  last_name: string;
  email: string;
  phone: string; // Required for volunteers
  age: number; // Must be 18-100
  
  // Volunteer details
  volunteer_interest: VolunteerInterest;
  availability: VolunteerAvailability;
  experience?: string; // Optional, max 1000 characters
  motivation: string; // Required, 20-1500 characters
  schedule_preferences?: string; // Optional, max 500 characters
  
  // Status management (different from other inquiries)
  status: VolunteerStatus;
  priority: InquiryPriority;
  
  // Metadata
  source?: string;
  ip_address?: string;
  user_agent?: string;
  
  // Audit fields
  created_at: string;
  updated_at: string;
  created_by: string;
  updated_by: string;
  is_deleted: boolean;
  deleted_at?: string;
}

// Submission interfaces - data sent from forms (before database processing)
export interface BaseInquirySubmission {
  contact_name: string;
  email: string;
  phone?: string;
}

export interface BusinessInquirySubmission extends BaseInquirySubmission {
  organization_name: string;
  title: string;
  inquiry_type: BusinessInquiryType;
  message: string;
  industry?: string;
}

export interface DonationsInquirySubmission extends BaseInquirySubmission {
  donor_type: DonorType;
  message: string;
  organization?: string;
  interest_area?: InterestArea;
  preferred_amount_range?: AmountRange;
  donation_frequency?: DonationFrequency;
}

export interface MediaInquirySubmission extends BaseInquirySubmission {
  outlet: string;
  title: string;
  phone: string; // Required for media
  subject: string;
  urgency: MediaUrgency;
  media_type?: MediaType;
  deadline?: string;
}

export interface VolunteerApplicationSubmission {
  // Personal information
  first_name: string;
  last_name: string;
  email: string;
  phone: string;
  age: number;
  
  // Volunteer details
  volunteer_interest: VolunteerInterest;
  availability: VolunteerAvailability;
  experience?: string;
  motivation: string;
  schedule_preferences?: string;
}

// API Response interfaces
export interface InquirySubmissionResponse {
  // Domain-specific inquiry data (one will be populated)
  business_inquiry?: BusinessInquiry;
  donations_inquiry?: DonationsInquiry;
  media_inquiry?: MediaInquiry;
  volunteer_application?: VolunteerApplication;
  
  // Response metadata
  correlation_id: string;
  success: boolean;
  message?: string;
  
  // Error information
  error?: string;
  validation_errors?: string[];
  retry_after?: number; // For rate limiting
}

export interface InquiryGetResponse {
  // Domain-specific inquiry data (one will be populated)
  business_inquiry?: BusinessInquiry;
  donations_inquiry?: DonationsInquiry;
  media_inquiry?: MediaInquiry;
  volunteer_application?: VolunteerApplication;
  
  // Response metadata
  correlation_id: string;
  success?: boolean;
  error?: string;
}

// Union types for working with any inquiry type
export type AnyInquiry = BusinessInquiry | DonationsInquiry | MediaInquiry | VolunteerApplication;
export type AnyInquirySubmission = BusinessInquirySubmission | DonationsInquirySubmission | MediaInquirySubmission | VolunteerApplicationSubmission;

// Type guards for inquiry type detection
export function isBusinessInquiry(inquiry: AnyInquiry): inquiry is BusinessInquiry {
  return 'organization_name' in inquiry && 'inquiry_type' in inquiry;
}

export function isDonationsInquiry(inquiry: AnyInquiry): inquiry is DonationsInquiry {
  return 'donor_type' in inquiry && !('outlet' in inquiry);
}

export function isMediaInquiry(inquiry: AnyInquiry): inquiry is MediaInquiry {
  return 'outlet' in inquiry && 'urgency' in inquiry;
}

export function isVolunteerApplication(inquiry: AnyInquiry): inquiry is VolunteerApplication {
  return 'application_id' in inquiry && 'volunteer_interest' in inquiry && 'availability' in inquiry;
}

export function isBusinessInquirySubmission(submission: AnyInquirySubmission): submission is BusinessInquirySubmission {
  return 'organization_name' in submission && 'inquiry_type' in submission;
}

export function isDonationsInquirySubmission(submission: AnyInquirySubmission): submission is DonationsInquirySubmission {
  return 'donor_type' in submission && !('outlet' in submission);
}

export function isMediaInquirySubmission(submission: AnyInquirySubmission): submission is MediaInquirySubmission {
  return 'outlet' in submission && 'urgency' in submission;
}

export function isVolunteerApplicationSubmission(submission: AnyInquirySubmission): submission is VolunteerApplicationSubmission {
  return 'first_name' in submission && 'last_name' in submission && 'volunteer_interest' in submission && 'availability' in submission;
}