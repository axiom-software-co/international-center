import { describe, it, expect } from 'vitest';
import {
  validateBaseInquiry,
  validateBusinessInquiry,
  validateDonationsInquiry,
  validateMediaInquiry,
  validateEmail,
  validatePhone,
  validateContactName,
  validateMessageLength,
  validateSubjectLength,
  sanitizeInput,
  InquiryValidationError,
  ValidationResult
} from './validation';
import type { BaseInquiry, BusinessInquiry, DonationsInquiry, MediaInquiry } from './types';

describe('Inquiry Validation', () => {
  describe('validateEmail', () => {
    it('should validate correct email formats', () => {
      const validEmails = [
        'test@example.com',
        'user.name@domain.co.uk',
        'first+last@company.org',
        'admin@sub.domain.edu'
      ];

      validEmails.forEach(email => {
        const result = validateEmail(email);
        expect(result.isValid).toBe(true);
        expect(result.errors).toHaveLength(0);
      });
    });

    it('should reject invalid email formats', () => {
      const invalidEmails = [
        'invalid-email',
        '@domain.com',
        'user@',
        'user..name@domain.com',
        'user@domain',
        ''
      ];

      invalidEmails.forEach(email => {
        const result = validateEmail(email);
        expect(result.isValid).toBe(false);
        expect(result.errors).toContain('Invalid email format');
      });
    });

    it('should enforce email length constraints', () => {
      const longEmail = 'a'.repeat(250) + '@example.com';
      const result = validateEmail(longEmail);
      
      expect(result.isValid).toBe(false);
      expect(result.errors).toContain('Email must not exceed 254 characters');
    });
  });

  describe('validatePhone', () => {
    it('should validate US phone number formats', () => {
      const validPhones = [
        '+1-555-123-4567',
        '+1 (555) 123-4567',
        '555-123-4567',
        '(555) 123-4567',
        '5551234567'
      ];

      validPhones.forEach(phone => {
        const result = validatePhone(phone);
        expect(result.isValid).toBe(true);
        expect(result.errors).toHaveLength(0);
      });
    });

    it('should reject invalid phone formats', () => {
      const invalidPhones = [
        '123',
        '555-12-345',
        'abc-def-ghij',
        '+1-555-123-456789',
        ''
      ];

      invalidPhones.forEach(phone => {
        const result = validatePhone(phone);
        expect(result.isValid).toBe(false);
        expect(result.errors).toContain('Invalid phone number format');
      });
    });

    it('should handle optional phone validation', () => {
      const result = validatePhone(undefined);
      expect(result.isValid).toBe(true);
      expect(result.errors).toHaveLength(0);
    });
  });

  describe('validateContactName', () => {
    it('should validate proper contact names', () => {
      const validNames = [
        'John Doe',
        'Mary Jane Smith',
        'Dr. Johnson',
        'Sarah O\'Connor'
      ];

      validNames.forEach(name => {
        const result = validateContactName(name);
        expect(result.isValid).toBe(true);
        expect(result.errors).toHaveLength(0);
      });
    });

    it('should enforce name length constraints', () => {
      const shortName = 'A';
      const longName = 'A'.repeat(51);

      const shortResult = validateContactName(shortName);
      expect(shortResult.isValid).toBe(false);
      expect(shortResult.errors).toContain('Contact name must be at least 2 characters');

      const longResult = validateContactName(longName);
      expect(longResult.isValid).toBe(false);
      expect(longResult.errors).toContain('Contact name must not exceed 50 characters');
    });

    it('should reject invalid characters in names', () => {
      const invalidNames = [
        'John123',
        'Mary@Smith',
        'User<script>',
        'Name with numbers 456'
      ];

      invalidNames.forEach(name => {
        const result = validateContactName(name);
        expect(result.isValid).toBe(false);
        expect(result.errors).toContain('Contact name contains invalid characters');
      });
    });
  });

  describe('validateMessageLength', () => {
    it('should validate business message length (20-1500 chars)', () => {
      const shortMessage = 'Hi there!';
      const validMessage = 'We are interested in exploring partnership opportunities with your organization to advance medical research.';
      const longMessage = 'A'.repeat(1501);

      expect(validateMessageLength(shortMessage, 'business').isValid).toBe(false);
      expect(validateMessageLength(validMessage, 'business').isValid).toBe(true);
      expect(validateMessageLength(longMessage, 'business').isValid).toBe(false);
    });

    it('should validate donations message length (20-2000 chars)', () => {
      const shortMessage = 'Hi there!';
      const validMessage = 'I would like to make a donation to support your research initiatives and help advance medical treatments for patients.';
      const longMessage = 'A'.repeat(2001);

      expect(validateMessageLength(shortMessage, 'donations').isValid).toBe(false);
      expect(validateMessageLength(validMessage, 'donations').isValid).toBe(true);
      expect(validateMessageLength(longMessage, 'donations').isValid).toBe(false);
    });
  });

  describe('validateSubjectLength', () => {
    it('should validate media subject length (20-1500 chars)', () => {
      const shortSubject = 'Interview request';
      const validSubject = 'Request for interview regarding new treatment protocol and FDA approval process for innovative therapies';
      const longSubject = 'A'.repeat(1501);

      expect(validateSubjectLength(shortSubject).isValid).toBe(false);
      expect(validateSubjectLength(validSubject).isValid).toBe(true);
      expect(validateSubjectLength(longSubject).isValid).toBe(false);
    });
  });

  describe('sanitizeInput', () => {
    it('should remove dangerous HTML and script tags', () => {
      const dangerousInput = '<script>alert("xss")</script>Hello World<img src="x" onerror="alert(1)">';
      const sanitized = sanitizeInput(dangerousInput);
      
      expect(sanitized).not.toContain('<script>');
      expect(sanitized).not.toContain('onerror');
      expect(sanitized).toContain('Hello World');
    });

    it('should preserve safe content', () => {
      const safeInput = 'This is a normal message with some punctuation! Email: test@example.com';
      const sanitized = sanitizeInput(safeInput);
      
      expect(sanitized).toBe(safeInput);
    });

    it('should handle empty and null inputs', () => {
      expect(sanitizeInput('')).toBe('');
      expect(sanitizeInput(null)).toBe('');
      expect(sanitizeInput(undefined)).toBe('');
    });
  });

  describe('validateBaseInquiry', () => {
    const validBaseInquiry: BaseInquiry = {
      inquiry_id: '123e4567-e89b-12d3-a456-426614174000',
      contact_name: 'John Doe',
      email: 'john.doe@example.com',
      status: 'new',
      priority: 'medium',
      source: 'website',
      created_at: '2024-03-15T10:00:00Z',
      updated_at: '2024-03-15T10:00:00Z',
      created_by: 'system',
      updated_by: 'system',
      is_deleted: false
    };

    it('should validate a complete base inquiry', () => {
      const result = validateBaseInquiry(validBaseInquiry);
      expect(result.isValid).toBe(true);
      expect(result.errors).toHaveLength(0);
    });

    it('should reject inquiry with missing required fields', () => {
      const incompleteInquiry = {
        ...validBaseInquiry,
        contact_name: '',
        email: ''
      };

      const result = validateBaseInquiry(incompleteInquiry as BaseInquiry);
      expect(result.isValid).toBe(false);
      expect(result.errors).toContain('Contact name is required');
      expect(result.errors).toContain('Email is required');
    });

    it('should validate optional phone when provided', () => {
      const inquiryWithPhone = {
        ...validBaseInquiry,
        phone: 'invalid-phone'
      };

      const result = validateBaseInquiry(inquiryWithPhone);
      expect(result.isValid).toBe(false);
      expect(result.errors).toContain('Invalid phone number format');
    });

    it('should validate status and priority enums', () => {
      const inquiryWithInvalidStatus = {
        ...validBaseInquiry,
        status: 'invalid-status' as any,
        priority: 'invalid-priority' as any
      };

      const result = validateBaseInquiry(inquiryWithInvalidStatus);
      expect(result.isValid).toBe(false);
      expect(result.errors).toContain('Invalid inquiry status');
      expect(result.errors).toContain('Invalid inquiry priority');
    });
  });

  describe('validateBusinessInquiry', () => {
    const validBusinessInquiry: BusinessInquiry = {
      inquiry_id: '123e4567-e89b-12d3-a456-426614174000',
      contact_name: 'John Smith',
      email: 'john.smith@company.com',
      organization_name: 'Acme Corporation',
      title: 'Director of Partnerships',
      inquiry_type: 'partnership',
      message: 'We are interested in exploring partnership opportunities with your organization.',
      status: 'new',
      priority: 'medium',
      source: 'website',
      created_at: '2024-03-15T10:00:00Z',
      updated_at: '2024-03-15T10:00:00Z',
      created_by: 'system',
      updated_by: 'system',
      is_deleted: false
    };

    it('should validate a complete business inquiry', () => {
      const result = validateBusinessInquiry(validBusinessInquiry);
      expect(result.isValid).toBe(true);
      expect(result.errors).toHaveLength(0);
    });

    it('should reject business inquiry with missing required fields', () => {
      const incompleteInquiry = {
        ...validBusinessInquiry,
        organization_name: '',
        title: '',
        inquiry_type: '' as any,
        message: 'Short'
      };

      const result = validateBusinessInquiry(incompleteInquiry);
      expect(result.isValid).toBe(false);
      expect(result.errors).toContain('Organization name is required');
      expect(result.errors).toContain('Title is required');
      expect(result.errors).toContain('Inquiry type is required');
      expect(result.errors).toContain('Message must be between 20 and 1500 characters');
    });

    it('should validate business inquiry type enum', () => {
      const inquiryWithInvalidType = {
        ...validBusinessInquiry,
        inquiry_type: 'invalid-type' as any
      };

      const result = validateBusinessInquiry(inquiryWithInvalidType);
      expect(result.isValid).toBe(false);
      expect(result.errors).toContain('Invalid business inquiry type');
    });
  });

  describe('validateDonationsInquiry', () => {
    const validDonationsInquiry: DonationsInquiry = {
      inquiry_id: '123e4567-e89b-12d3-a456-426614174000',
      contact_name: 'Mary Johnson',
      email: 'mary.johnson@email.com',
      donor_type: 'individual',
      message: 'I would like to make a donation to support your research initiatives.',
      status: 'new',
      priority: 'medium',
      source: 'website',
      created_at: '2024-03-15T10:00:00Z',
      updated_at: '2024-03-15T10:00:00Z',
      created_by: 'system',
      updated_by: 'system',
      is_deleted: false
    };

    it('should validate a complete donations inquiry', () => {
      const result = validateDonationsInquiry(validDonationsInquiry);
      expect(result.isValid).toBe(true);
      expect(result.errors).toHaveLength(0);
    });

    it('should reject donations inquiry with missing required fields', () => {
      const incompleteInquiry = {
        ...validDonationsInquiry,
        donor_type: '' as any,
        message: 'Short'
      };

      const result = validateDonationsInquiry(incompleteInquiry);
      expect(result.isValid).toBe(false);
      expect(result.errors).toContain('Donor type is required');
      expect(result.errors).toContain('Message must be between 20 and 2000 characters');
    });

    it('should validate donor type enum', () => {
      const inquiryWithInvalidDonorType = {
        ...validDonationsInquiry,
        donor_type: 'invalid-donor' as any
      };

      const result = validateDonationsInquiry(inquiryWithInvalidDonorType);
      expect(result.isValid).toBe(false);
      expect(result.errors).toContain('Invalid donor type');
    });

    it('should validate optional enum fields', () => {
      const inquiryWithInvalidEnums = {
        ...validDonationsInquiry,
        interest_area: 'invalid-area' as any,
        preferred_amount_range: 'invalid-range' as any,
        donation_frequency: 'invalid-frequency' as any
      };

      const result = validateDonationsInquiry(inquiryWithInvalidEnums);
      expect(result.isValid).toBe(false);
      expect(result.errors).toContain('Invalid interest area');
      expect(result.errors).toContain('Invalid amount range');
      expect(result.errors).toContain('Invalid donation frequency');
    });
  });

  describe('validateMediaInquiry', () => {
    const validMediaInquiry: MediaInquiry = {
      inquiry_id: '123e4567-e89b-12d3-a456-426614174000',
      contact_name: 'Sarah Reporter',
      email: 'sarah.reporter@newsnetwork.com',
      outlet: 'Medical News Network',
      title: 'Senior Medical Reporter',
      phone: '+1-555-987-6543',
      subject: 'Request for interview regarding new treatment protocol',
      urgency: 'medium',
      status: 'new',
      priority: 'medium',
      source: 'website',
      created_at: '2024-03-15T10:00:00Z',
      updated_at: '2024-03-15T10:00:00Z',
      created_by: 'system',
      updated_by: 'system',
      is_deleted: false
    };

    it('should validate a complete media inquiry', () => {
      const result = validateMediaInquiry(validMediaInquiry);
      expect(result.isValid).toBe(true);
      expect(result.errors).toHaveLength(0);
    });

    it('should reject media inquiry with missing required fields', () => {
      const incompleteInquiry = {
        ...validMediaInquiry,
        outlet: '',
        title: '',
        phone: '',
        subject: 'Short'
      };

      const result = validateMediaInquiry(incompleteInquiry);
      expect(result.isValid).toBe(false);
      expect(result.errors).toContain('Outlet is required');
      expect(result.errors).toContain('Title is required');
      expect(result.errors).toContain('Phone is required for media inquiries');
      expect(result.errors).toContain('Subject must be between 20 and 1500 characters');
    });

    it('should validate urgency enum', () => {
      const inquiryWithInvalidUrgency = {
        ...validMediaInquiry,
        urgency: 'invalid-urgency' as any
      };

      const result = validateMediaInquiry(inquiryWithInvalidUrgency);
      expect(result.isValid).toBe(false);
      expect(result.errors).toContain('Invalid urgency level');
    });

    it('should validate optional media type enum', () => {
      const inquiryWithInvalidMediaType = {
        ...validMediaInquiry,
        media_type: 'invalid-media' as any
      };

      const result = validateMediaInquiry(inquiryWithInvalidMediaType);
      expect(result.isValid).toBe(false);
      expect(result.errors).toContain('Invalid media type');
    });

    it('should validate deadline format when provided', () => {
      const inquiryWithInvalidDeadline = {
        ...validMediaInquiry,
        deadline: 'invalid-date'
      };

      const result = validateMediaInquiry(inquiryWithInvalidDeadline);
      expect(result.isValid).toBe(false);
      expect(result.errors).toContain('Invalid deadline format');
    });
  });

  describe('InquiryValidationError', () => {
    it('should create validation error with proper structure', () => {
      const errors = ['Error 1', 'Error 2'];
      const validationError = new InquiryValidationError('Validation failed', errors);

      expect(validationError.message).toBe('Validation failed');
      expect(validationError.errors).toEqual(errors);
      expect(validationError.name).toBe('InquiryValidationError');
      expect(validationError).toBeInstanceOf(Error);
    });
  });

  describe('Edge Cases', () => {
    it('should handle null and undefined inputs gracefully', () => {
      expect(() => validateBaseInquiry(null as any)).not.toThrow();
      expect(() => validateBaseInquiry(undefined as any)).not.toThrow();
      expect(() => validateBusinessInquiry(null as any)).not.toThrow();
      expect(() => validateDonationsInquiry(null as any)).not.toThrow();
      expect(() => validateMediaInquiry(null as any)).not.toThrow();
    });

    it('should handle unicode and international characters', () => {
      const unicodeInquiry: BaseInquiry = {
        inquiry_id: '123e4567-e89b-12d3-a456-426614174000',
        contact_name: 'José María García',
        email: 'jose.garcia@español.com',
        status: 'new',
        priority: 'medium',
        source: 'website',
        created_at: '2024-03-15T10:00:00Z',
        updated_at: '2024-03-15T10:00:00Z',
        created_by: 'system',
        updated_by: 'system',
        is_deleted: false
      };

      const result = validateBaseInquiry(unicodeInquiry);
      expect(result.isValid).toBe(true);
    });
  });
});