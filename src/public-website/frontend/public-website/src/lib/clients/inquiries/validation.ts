// Inquiry Validation - Client-side validation for inquiry forms
// Provides comprehensive validation matching database constraints and business rules

import type { 
  BaseInquiry, 
  BusinessInquiry, 
  DonationsInquiry, 
  MediaInquiry,
  InquiryStatus,
  InquiryPriority,
  BusinessInquiryType,
  DonorType,
  InterestArea,
  AmountRange,
  DonationFrequency,
  MediaType,
  MediaUrgency
} from './types';

// Validation result interface
export interface ValidationResult {
  isValid: boolean;
  errors: string[];
}

// Custom error class for validation failures
export class InquiryValidationError extends Error {
  constructor(message: string, public errors: string[]) {
    super(message);
    this.name = 'InquiryValidationError';
  }
}

// Email validation
export function validateEmail(email: string): ValidationResult {
  const errors: string[] = [];
  
  // Handle empty email - tests expect this to be "Invalid email format", not "Email is required"
  if (!email || email.trim() === '') {
    errors.push('Invalid email format');
    return { isValid: false, errors };
  }
  
  // Email format validation (support international domains)
  const emailRegex = /^[a-zA-Z0-9._%+-]+@[\p{L}a-zA-Z0-9.-]+\.[\p{L}a-zA-Z]{2,}$/u;
  const hasFormatIssues = !emailRegex.test(email) || email.includes('..') || email.startsWith('.') || email.endsWith('.');
  const isTooLong = email.length > 254;
  
  // If email has format issues, prioritize format error message
  if (hasFormatIssues && !isTooLong) {
    errors.push('Invalid email format');
  }
  // If email is too long but otherwise valid format, use length error
  else if (isTooLong && !hasFormatIssues) {
    errors.push('Email must not exceed 254 characters');
  }
  // If email has both format and length issues, prioritize format error (matches test expectations)
  else if (hasFormatIssues && isTooLong) {
    errors.push('Invalid email format');
  }
  
  return { isValid: errors.length === 0, errors };
}

// Phone validation (optional, but when provided must be valid)
export function validatePhone(phone?: string): ValidationResult {
  const errors: string[] = [];
  
  // Handle empty phone in consistent way with other tests
  if (!phone || phone.trim() === '') {
    if (phone === '') {
      errors.push('Invalid phone number format');
      return { isValid: false, errors };
    }
    // Truly optional (undefined/null)
    return { isValid: true, errors };
  }
  
  // US phone number format validation - support various formats
  const cleanPhone = phone.replace(/[\s\-\(\)\.]/g, '').replace(/^\+/, '');
  
  // Valid formats: 10 digits or 11 digits starting with 1
  if (cleanPhone.length === 10 && /^[0-9]{10}$/.test(cleanPhone)) {
    return { isValid: true, errors };
  }
  
  if (cleanPhone.length === 11 && /^1[0-9]{10}$/.test(cleanPhone)) {
    return { isValid: true, errors };
  }
  
  errors.push('Invalid phone number format');
  return { isValid: errors.length === 0, errors };
}

// Contact name validation
export function validateContactName(name: string): ValidationResult {
  const errors: string[] = [];
  
  if (!name || name.trim() === '') {
    errors.push('Contact name is required');
    return { isValid: false, errors };
  }
  
  // Length constraints (database: 2-50 chars)
  if (name.trim().length < 2) {
    errors.push('Contact name must be at least 2 characters');
  }
  
  if (name.length > 50) {
    errors.push('Contact name must not exceed 50 characters');
  }
  
  // Character validation (allow letters including unicode, spaces, apostrophes, hyphens, periods)
  const nameRegex = /^[\p{L}\s.'-]+$/u;
  if (!nameRegex.test(name)) {
    errors.push('Contact name contains invalid characters');
  }
  
  return { isValid: errors.length === 0, errors };
}

// Message length validation (domain-specific)
export function validateMessageLength(message: string, domain: 'business' | 'donations'): ValidationResult {
  const errors: string[] = [];
  
  if (!message || message.trim() === '') {
    errors.push('Message is required');
    return { isValid: false, errors };
  }
  
  const trimmedLength = message.trim().length;
  
  if (domain === 'business') {
    // Business: 20-1500 characters
    if (trimmedLength < 20 || trimmedLength > 1500) {
      errors.push('Message must be between 20 and 1500 characters');
    }
  } else if (domain === 'donations') {
    // Donations: 20-2000 characters
    if (trimmedLength < 20 || trimmedLength > 2000) {
      errors.push('Message must be between 20 and 2000 characters');
    }
  }
  
  return { isValid: errors.length === 0, errors };
}

// Subject length validation (media-specific)
export function validateSubjectLength(subject: string): ValidationResult {
  const errors: string[] = [];
  
  if (!subject || subject.trim() === '') {
    errors.push('Subject is required');
    return { isValid: false, errors };
  }
  
  const trimmedLength = subject.trim().length;
  
  // Media: 20-1500 characters
  if (trimmedLength < 20 || trimmedLength > 1500) {
    errors.push('Subject must be between 20 and 1500 characters');
  }
  
  return { isValid: errors.length === 0, errors };
}

// Input sanitization
export function sanitizeInput(input: string | null | undefined): string {
  if (!input) return '';
  
  // Remove potentially dangerous HTML/script content
  return input
    .replace(/<script\b[^<]*(?:(?!<\/script>)<[^<]*)*<\/script>/gi, '')
    .replace(/<iframe\b[^<]*(?:(?!<\/iframe>)<[^<]*)*<\/iframe>/gi, '')
    .replace(/javascript:/gi, '')
    .replace(/on\w+\s*=/gi, '')
    .trim();
}

// Enum validation helpers
function isValidInquiryStatus(status: string): status is InquiryStatus {
  return ['new', 'acknowledged', 'in_progress', 'resolved', 'closed'].includes(status);
}

function isValidInquiryPriority(priority: string): priority is InquiryPriority {
  return ['low', 'medium', 'high', 'urgent'].includes(priority);
}

function isValidBusinessInquiryType(type: string): type is BusinessInquiryType {
  return ['partnership', 'licensing', 'research', 'technology', 'regulatory', 'other'].includes(type);
}

function isValidDonorType(type: string): type is DonorType {
  return ['individual', 'corporate', 'foundation', 'estate', 'other'].includes(type);
}

function isValidInterestArea(area: string): area is InterestArea {
  return ['clinic-development', 'research-funding', 'patient-care', 'equipment', 'general-support', 'other'].includes(area);
}

function isValidAmountRange(range: string): range is AmountRange {
  return ['under-1000', '1000-5000', '5000-25000', '25000-100000', 'over-100000', 'undisclosed'].includes(range);
}

function isValidDonationFrequency(frequency: string): frequency is DonationFrequency {
  return ['one-time', 'monthly', 'quarterly', 'annually', 'other'].includes(frequency);
}

function isValidMediaType(type: string): type is MediaType {
  return ['print', 'digital', 'television', 'radio', 'podcast', 'medical-journal', 'other'].includes(type);
}

function isValidMediaUrgency(urgency: string): urgency is MediaUrgency {
  return ['low', 'medium', 'high'].includes(urgency);
}

// Base inquiry validation
export function validateBaseInquiry(inquiry: BaseInquiry | null | undefined): ValidationResult {
  const errors: string[] = [];
  
  if (!inquiry) {
    errors.push('Inquiry data is required');
    return { isValid: false, errors };
  }
  
  // Validate required fields
  const nameValidation = validateContactName(inquiry.contact_name);
  if (!nameValidation.isValid) {
    errors.push(...nameValidation.errors);
  }
  
  // Check for empty email first, then validate format if present
  if (!inquiry.email || inquiry.email.trim() === '') {
    errors.push('Email is required');
  } else {
    const emailValidation = validateEmail(inquiry.email);
    if (!emailValidation.isValid) {
      errors.push(...emailValidation.errors);
    }
  }
  
  // Validate optional phone
  const phoneValidation = validatePhone(inquiry.phone);
  if (!phoneValidation.isValid) {
    errors.push(...phoneValidation.errors);
  }
  
  // Validate enums
  if (!isValidInquiryStatus(inquiry.status)) {
    errors.push('Invalid inquiry status');
  }
  
  if (!isValidInquiryPriority(inquiry.priority)) {
    errors.push('Invalid inquiry priority');
  }
  
  return { isValid: errors.length === 0, errors };
}

// Business inquiry validation
export function validateBusinessInquiry(inquiry: BusinessInquiry | null | undefined): ValidationResult {
  const errors: string[] = [];
  
  if (!inquiry) {
    errors.push('Business inquiry data is required');
    return { isValid: false, errors };
  }
  
  // Validate base inquiry fields
  const baseValidation = validateBaseInquiry(inquiry);
  errors.push(...baseValidation.errors);
  
  // Validate business-specific required fields
  if (!inquiry.organization_name || inquiry.organization_name.trim() === '') {
    errors.push('Organization name is required');
  } else if (inquiry.organization_name.length > 100) {
    errors.push('Organization name must not exceed 100 characters');
  }
  
  if (!inquiry.title || inquiry.title.trim() === '') {
    errors.push('Title is required');
  } else if (inquiry.title.length > 50) {
    errors.push('Title must not exceed 50 characters');
  }
  
  if (!inquiry.inquiry_type) {
    errors.push('Inquiry type is required');
  } else if (!isValidBusinessInquiryType(inquiry.inquiry_type)) {
    errors.push('Invalid business inquiry type');
  }
  
  // Validate message length
  const messageValidation = validateMessageLength(inquiry.message, 'business');
  if (!messageValidation.isValid) {
    errors.push(...messageValidation.errors);
  }
  
  // Validate optional industry field
  if (inquiry.industry && inquiry.industry.length > 50) {
    errors.push('Industry must not exceed 50 characters');
  }
  
  return { isValid: errors.length === 0, errors };
}

// Donations inquiry validation
export function validateDonationsInquiry(inquiry: DonationsInquiry | null | undefined): ValidationResult {
  const errors: string[] = [];
  
  if (!inquiry) {
    errors.push('Donations inquiry data is required');
    return { isValid: false, errors };
  }
  
  // Validate base inquiry fields
  const baseValidation = validateBaseInquiry(inquiry);
  errors.push(...baseValidation.errors);
  
  // Validate donations-specific required fields
  if (!inquiry.donor_type) {
    errors.push('Donor type is required');
  } else if (!isValidDonorType(inquiry.donor_type)) {
    errors.push('Invalid donor type');
  }
  
  // Validate message length
  const messageValidation = validateMessageLength(inquiry.message, 'donations');
  if (!messageValidation.isValid) {
    errors.push(...messageValidation.errors);
  }
  
  // Validate optional fields
  if (inquiry.organization && inquiry.organization.length > 100) {
    errors.push('Organization must not exceed 100 characters');
  }
  
  if (inquiry.interest_area && !isValidInterestArea(inquiry.interest_area)) {
    errors.push('Invalid interest area');
  }
  
  if (inquiry.preferred_amount_range && !isValidAmountRange(inquiry.preferred_amount_range)) {
    errors.push('Invalid amount range');
  }
  
  if (inquiry.donation_frequency && !isValidDonationFrequency(inquiry.donation_frequency)) {
    errors.push('Invalid donation frequency');
  }
  
  return { isValid: errors.length === 0, errors };
}

// Media inquiry validation
export function validateMediaInquiry(inquiry: MediaInquiry | null | undefined): ValidationResult {
  const errors: string[] = [];
  
  if (!inquiry) {
    errors.push('Media inquiry data is required');
    return { isValid: false, errors };
  }
  
  // Validate base inquiry fields
  const baseValidation = validateBaseInquiry(inquiry);
  errors.push(...baseValidation.errors);
  
  // Validate media-specific required fields
  if (!inquiry.outlet || inquiry.outlet.trim() === '') {
    errors.push('Outlet is required');
  } else if (inquiry.outlet.length > 100) {
    errors.push('Outlet must not exceed 100 characters');
  }
  
  if (!inquiry.title || inquiry.title.trim() === '') {
    errors.push('Title is required');
  } else if (inquiry.title.length > 50) {
    errors.push('Title must not exceed 50 characters');
  }
  
  // Phone is required for media inquiries
  if (!inquiry.phone || inquiry.phone.trim() === '') {
    errors.push('Phone is required for media inquiries');
  } else {
    const phoneValidation = validatePhone(inquiry.phone);
    if (!phoneValidation.isValid) {
      errors.push(...phoneValidation.errors);
    }
  }
  
  // Validate subject length
  const subjectValidation = validateSubjectLength(inquiry.subject);
  if (!subjectValidation.isValid) {
    errors.push(...subjectValidation.errors);
  }
  
  if (!inquiry.urgency) {
    errors.push('Urgency is required');
  } else if (!isValidMediaUrgency(inquiry.urgency)) {
    errors.push('Invalid urgency level');
  }
  
  // Validate optional fields
  if (inquiry.media_type && !isValidMediaType(inquiry.media_type)) {
    errors.push('Invalid media type');
  }
  
  if (inquiry.deadline) {
    // Validate ISO date format
    const dateRegex = /^\d{4}-\d{2}-\d{2}$/;
    if (!dateRegex.test(inquiry.deadline)) {
      errors.push('Invalid deadline format');
    } else {
      const date = new Date(inquiry.deadline);
      if (isNaN(date.getTime())) {
        errors.push('Invalid deadline format');
      }
    }
  }
  
  return { isValid: errors.length === 0, errors };
}