import { describe, it, expect } from 'vitest';
import { 
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
  MediaType
} from './types';

describe('Inquiry Types', () => {
  describe('BaseInquiry', () => {
    it('should define required base fields for all inquiries', () => {
      const baseInquiry: BaseInquiry = {
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

      expect(baseInquiry.inquiry_id).toBe('123e4567-e89b-12d3-a456-426614174000');
      expect(baseInquiry.contact_name).toBe('John Doe');
      expect(baseInquiry.email).toBe('john.doe@example.com');
      expect(baseInquiry.status).toBe('new');
      expect(baseInquiry.priority).toBe('medium');
    });

    it('should support optional fields', () => {
      const baseInquiry: BaseInquiry = {
        inquiry_id: '123e4567-e89b-12d3-a456-426614174000',
        contact_name: 'John Doe',
        email: 'john.doe@example.com',
        phone: '+1-555-123-4567',
        status: 'new',
        priority: 'high',
        source: 'website',
        ip_address: '192.168.1.1',
        user_agent: 'Mozilla/5.0...',
        created_at: '2024-03-15T10:00:00Z',
        updated_at: '2024-03-15T10:00:00Z',
        created_by: 'system',
        updated_by: 'system',
        is_deleted: false,
        deleted_at: '2024-03-15T11:00:00Z'
      };

      expect(baseInquiry.phone).toBe('+1-555-123-4567');
      expect(baseInquiry.ip_address).toBe('192.168.1.1');
      expect(baseInquiry.user_agent).toBe('Mozilla/5.0...');
      expect(baseInquiry.deleted_at).toBe('2024-03-15T11:00:00Z');
    });
  });

  describe('BusinessInquiry', () => {
    it('should extend BaseInquiry with business-specific fields', () => {
      const businessInquiry: BusinessInquiry = {
        inquiry_id: '123e4567-e89b-12d3-a456-426614174000',
        contact_name: 'John Smith',
        email: 'john.smith@company.com',
        organization_name: 'Acme Corporation',
        title: 'Director of Partnerships',
        inquiry_type: 'partnership',
        message: 'We are interested in exploring partnership opportunities.',
        status: 'new',
        priority: 'medium',
        source: 'website',
        created_at: '2024-03-15T10:00:00Z',
        updated_at: '2024-03-15T10:00:00Z',
        created_by: 'system',
        updated_by: 'system',
        is_deleted: false
      };

      expect(businessInquiry.organization_name).toBe('Acme Corporation');
      expect(businessInquiry.title).toBe('Director of Partnerships');
      expect(businessInquiry.inquiry_type).toBe('partnership');
      expect(businessInquiry.message).toContain('partnership opportunities');
    });

    it('should support optional industry field', () => {
      const businessInquiry: BusinessInquiry = {
        inquiry_id: '123e4567-e89b-12d3-a456-426614174000',
        contact_name: 'Jane Doe',
        email: 'jane.doe@biotech.com',
        organization_name: 'BioTech Solutions',
        title: 'CEO',
        inquiry_type: 'research',
        industry: 'Biotechnology',
        message: 'Looking for research collaboration opportunities.',
        status: 'new',
        priority: 'high',
        source: 'website',
        created_at: '2024-03-15T10:00:00Z',
        updated_at: '2024-03-15T10:00:00Z',
        created_by: 'system',
        updated_by: 'system',
        is_deleted: false
      };

      expect(businessInquiry.industry).toBe('Biotechnology');
    });
  });

  describe('DonationsInquiry', () => {
    it('should extend BaseInquiry with donations-specific fields', () => {
      const donationsInquiry: DonationsInquiry = {
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

      expect(donationsInquiry.donor_type).toBe('individual');
      expect(donationsInquiry.message).toContain('donation to support');
    });

    it('should support corporate donors with organization field', () => {
      const corporateDonation: DonationsInquiry = {
        inquiry_id: '123e4567-e89b-12d3-a456-426614174000',
        contact_name: 'Robert Wilson',
        email: 'robert.wilson@foundation.org',
        organization: 'Wilson Foundation',
        donor_type: 'foundation',
        interest_area: 'research-funding',
        preferred_amount_range: '25000-100000',
        donation_frequency: 'annually',
        message: 'Our foundation is interested in funding medical research projects.',
        status: 'new',
        priority: 'high',
        source: 'website',
        created_at: '2024-03-15T10:00:00Z',
        updated_at: '2024-03-15T10:00:00Z',
        created_by: 'system',
        updated_by: 'system',
        is_deleted: false
      };

      expect(corporateDonation.organization).toBe('Wilson Foundation');
      expect(corporateDonation.donor_type).toBe('foundation');
      expect(corporateDonation.interest_area).toBe('research-funding');
      expect(corporateDonation.preferred_amount_range).toBe('25000-100000');
      expect(corporateDonation.donation_frequency).toBe('annually');
    });
  });

  describe('MediaInquiry', () => {
    it('should extend BaseInquiry with media-specific fields', () => {
      const mediaInquiry: MediaInquiry = {
        inquiry_id: '123e4567-e89b-12d3-a456-426614174000',
        contact_name: 'Sarah Reporter',
        email: 'sarah.reporter@newsnetwork.com',
        outlet: 'Medical News Network',
        title: 'Senior Medical Reporter',
        phone: '+1-555-987-6543',
        subject: 'Request for interview regarding new treatment protocol',
        status: 'new',
        priority: 'medium',
        urgency: 'medium',
        source: 'website',
        created_at: '2024-03-15T10:00:00Z',
        updated_at: '2024-03-15T10:00:00Z',
        created_by: 'system',
        updated_by: 'system',
        is_deleted: false
      };

      expect(mediaInquiry.outlet).toBe('Medical News Network');
      expect(mediaInquiry.title).toBe('Senior Medical Reporter');
      expect(mediaInquiry.phone).toBe('+1-555-987-6543');
      expect(mediaInquiry.subject).toContain('interview regarding');
      expect(mediaInquiry.urgency).toBe('medium');
    });

    it('should support optional deadline and media type', () => {
      const urgentMediaInquiry: MediaInquiry = {
        inquiry_id: '123e4567-e89b-12d3-a456-426614174000',
        contact_name: 'Tom Journalist',
        email: 'tom.journalist@tv.com',
        outlet: 'Health TV',
        title: 'Health Correspondent',
        phone: '+1-555-111-2222',
        media_type: 'television',
        deadline: '2024-03-16',
        urgency: 'high',
        subject: 'Breaking: New FDA approval for innovative treatment',
        status: 'new',
        priority: 'urgent',
        source: 'website',
        created_at: '2024-03-15T10:00:00Z',
        updated_at: '2024-03-15T10:00:00Z',
        created_by: 'system',
        updated_by: 'system',
        is_deleted: false
      };

      expect(urgentMediaInquiry.media_type).toBe('television');
      expect(urgentMediaInquiry.deadline).toBe('2024-03-16');
      expect(urgentMediaInquiry.urgency).toBe('high');
    });
  });

  describe('Enum Types', () => {
    it('should define InquiryStatus enum values', () => {
      const statuses: InquiryStatus[] = ['new', 'acknowledged', 'in_progress', 'resolved', 'closed'];
      
      statuses.forEach(status => {
        expect(['new', 'acknowledged', 'in_progress', 'resolved', 'closed']).toContain(status);
      });
    });

    it('should define InquiryPriority enum values', () => {
      const priorities: InquiryPriority[] = ['low', 'medium', 'high', 'urgent'];
      
      priorities.forEach(priority => {
        expect(['low', 'medium', 'high', 'urgent']).toContain(priority);
      });
    });

    it('should define BusinessInquiryType enum values', () => {
      const types: BusinessInquiryType[] = ['partnership', 'licensing', 'research', 'technology', 'regulatory', 'other'];
      
      types.forEach(type => {
        expect(['partnership', 'licensing', 'research', 'technology', 'regulatory', 'other']).toContain(type);
      });
    });

    it('should define DonorType enum values', () => {
      const types: DonorType[] = ['individual', 'corporate', 'foundation', 'estate', 'other'];
      
      types.forEach(type => {
        expect(['individual', 'corporate', 'foundation', 'estate', 'other']).toContain(type);
      });
    });

    it('should define InterestArea enum values', () => {
      const areas: InterestArea[] = ['clinic-development', 'research-funding', 'patient-care', 'equipment', 'general-support', 'other'];
      
      areas.forEach(area => {
        expect(['clinic-development', 'research-funding', 'patient-care', 'equipment', 'general-support', 'other']).toContain(area);
      });
    });

    it('should define AmountRange enum values', () => {
      const ranges: AmountRange[] = ['under-1000', '1000-5000', '5000-25000', '25000-100000', 'over-100000', 'undisclosed'];
      
      ranges.forEach(range => {
        expect(['under-1000', '1000-5000', '5000-25000', '25000-100000', 'over-100000', 'undisclosed']).toContain(range);
      });
    });

    it('should define DonationFrequency enum values', () => {
      const frequencies: DonationFrequency[] = ['one-time', 'monthly', 'quarterly', 'annually', 'other'];
      
      frequencies.forEach(frequency => {
        expect(['one-time', 'monthly', 'quarterly', 'annually', 'other']).toContain(frequency);
      });
    });

    it('should define MediaType enum values', () => {
      const types: MediaType[] = ['print', 'digital', 'television', 'radio', 'podcast', 'medical-journal', 'other'];
      
      types.forEach(type => {
        expect(['print', 'digital', 'television', 'radio', 'podcast', 'medical-journal', 'other']).toContain(type);
      });
    });
  });

  describe('Type Validation', () => {
    it('should validate message length constraints for business inquiries', () => {
      const shortMessage = 'Hi';
      const validMessage = 'We are interested in exploring partnership opportunities with your organization.';
      const longMessage = 'A'.repeat(1501);

      expect(shortMessage.length).toBeLessThan(20);
      expect(validMessage.length).toBeGreaterThanOrEqual(20);
      expect(validMessage.length).toBeLessThanOrEqual(1500);
      expect(longMessage.length).toBeGreaterThan(1500);
    });

    it('should validate message length constraints for donations inquiries', () => {
      const shortMessage = 'Hi';
      const validMessage = 'I would like to make a donation to support your research initiatives and patient care programs.';
      const longMessage = 'A'.repeat(2001);

      expect(shortMessage.length).toBeLessThan(20);
      expect(validMessage.length).toBeGreaterThanOrEqual(20);
      expect(validMessage.length).toBeLessThanOrEqual(2000);
      expect(longMessage.length).toBeGreaterThan(2000);
    });

    it('should validate subject length constraints for media inquiries', () => {
      const shortSubject = 'Hi';
      const validSubject = 'Request for interview regarding new treatment protocol and FDA approval process';
      const longSubject = 'A'.repeat(1501);

      expect(shortSubject.length).toBeLessThan(20);
      expect(validSubject.length).toBeGreaterThanOrEqual(20);
      expect(validSubject.length).toBeLessThanOrEqual(1500);
      expect(longSubject.length).toBeGreaterThan(1500);
    });

    it('should validate contact name length constraints', () => {
      const shortName = 'A';
      const validName = 'John Doe';
      const longName = 'A'.repeat(51);

      expect(shortName.length).toBeLessThan(2);
      expect(validName.length).toBeGreaterThanOrEqual(2);
      expect(validName.length).toBeLessThanOrEqual(50);
      expect(longName.length).toBeGreaterThan(50);
    });

    it('should validate email format constraints', () => {
      const validEmail = 'john.doe@example.com';
      const longEmail = 'a'.repeat(250) + '@example.com';
      
      expect(validEmail.length).toBeLessThanOrEqual(254);
      expect(longEmail.length).toBeGreaterThan(254);
    });
  });
});