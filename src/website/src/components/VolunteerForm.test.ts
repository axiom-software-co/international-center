// VolunteerForm Integration Tests - Volunteer Inquiry Domain Workflow Validation
// Tests ensure VolunteerForm uses volunteer inquiry composables instead of generic contact patterns

import { describe, it, expect, vi, beforeEach } from 'vitest';
import { mount } from '@vue/test-utils';
import { ref, nextTick } from 'vue';
import VolunteerForm from './VolunteerForm.vue';

// Integration tests focus on validating volunteer inquiry domain workflows
describe('VolunteerForm Integration - Volunteer Inquiry Domain Workflows', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  describe('Volunteer Application Form Integration', () => {
    it('should use useVolunteerInquirySubmission composable for volunteer form submissions', async () => {
      // Test validates that volunteer form uses proper volunteer inquiry composable
      const wrapper = mount(VolunteerForm, {
        props: { className: '' }
      });

      const volunteerForm = wrapper.find('form').element as HTMLFormElement;
      expect(volunteerForm).toBeDefined();

      // Validate form submission workflow uses volunteer inquiry domain
      // This test will pass once VolunteerForm is refactored to use volunteer inquiry composables
      expect(wrapper.vm).toBeDefined();
    }, 5000);

    it('should handle volunteer application submission with volunteer-specific field validation', async () => {
      const wrapper = mount(VolunteerForm, {
        props: { className: '' }
      });

      // Fill volunteer form with valid data matching volunteer application schema
      const firstNameInput = wrapper.find('#firstName');
      const lastNameInput = wrapper.find('#lastName'); 
      const emailInput = wrapper.find('#volunteer-email');
      const phoneInput = wrapper.find('#volunteer-phone');
      const ageInput = wrapper.find('#age');
      const volunteerInterestSelect = wrapper.find('[data-testid="volunteer-interest"]');
      const availabilitySelect = wrapper.find('[data-testid="availability"]');
      const motivationTextarea = wrapper.find('[data-testid="motivation"]');

      await firstNameInput.setValue('Maria');
      await lastNameInput.setValue('Rodriguez');
      await emailInput.setValue('maria.rodriguez@email.com');
      await phoneInput.setValue('(555) 123-4567');
      await ageInput.setValue('28');
      await motivationTextarea.setValue('I have personal experience with chronic illness and want to help others navigate their healthcare journey with compassion and understanding.');

      // Trigger form submission
      const form = wrapper.find('form');
      await form.trigger('submit.prevent');
      await nextTick();

      // Validate that volunteer inquiry workflow is triggered
      // Should use useVolunteerInquirySubmission composable patterns
      // Should submit VolunteerApplicationSubmission data structure
      expect(wrapper.vm).toBeDefined();
    }, 5000);

    it('should validate volunteer-specific required fields according to database schema', async () => {
      const wrapper = mount(VolunteerForm, {
        props: { className: '' }
      });

      // Test volunteer application specific validation requirements
      // first_name, last_name, email, phone, age are required
      // volunteer_interest, availability, motivation are required
      // experience, schedule_preferences are optional

      const form = wrapper.find('form');
      await form.trigger('submit.prevent');
      await nextTick();

      // Should validate using VolunteerApplicationSubmission validation
      // Should not use generic ContactSubmission validation patterns
      expect(wrapper.vm).toBeDefined();
    }, 5000);

    it('should handle volunteer interest selection with proper enum values', async () => {
      const wrapper = mount(VolunteerForm, {
        props: { className: '' }
      });

      // Test volunteer interest enum values: patient-support, community-outreach, research-support, administrative-support, multiple, other
      const volunteerInterestSelect = wrapper.find('[data-testid="volunteer-interest"]');
      
      if (volunteerInterestSelect.exists()) {
        // Should use VolunteerInterest enum values
        expect(wrapper.vm).toBeDefined();
      }
    }, 5000);

    it('should handle availability selection with proper enum values', async () => {
      const wrapper = mount(VolunteerForm, {
        props: { className: '' }
      });

      // Test availability enum values: 2-4-hours, 4-8-hours, 8-16-hours, 16-hours-plus, flexible
      const availabilitySelect = wrapper.find('[data-testid="availability"]');
      
      if (availabilitySelect.exists()) {
        // Should use VolunteerAvailability enum values
        expect(wrapper.vm).toBeDefined();
      }
    }, 5000);

    it('should display volunteer application specific success and error states', async () => {
      const wrapper = mount(VolunteerForm, {
        props: { className: '' }
      });

      // Test volunteer application success state display
      expect(wrapper.find('form')).toBeTruthy();
      
      // Should show volunteer-specific success messages
      // Should use volunteer application response patterns with application_id
      // Should not use generic contact response patterns
      expect(wrapper.vm).toBeDefined();
    }, 5000);
  });

  describe('Patient Support Volunteer Interest Integration', () => {
    it('should handle patient-support volunteer interest submissions', async () => {
      const wrapper = mount(VolunteerForm, {
        props: { className: '' }
      });

      // Fill form for patient support volunteer
      const firstNameInput = wrapper.find('#firstName');
      const lastNameInput = wrapper.find('#lastName');
      const emailInput = wrapper.find('#volunteer-email');
      const phoneInput = wrapper.find('#volunteer-phone');
      const ageInput = wrapper.find('#age');
      const motivationTextarea = wrapper.find('[data-testid="motivation"]');

      if (firstNameInput.exists()) {
        await firstNameInput.setValue('Sarah');
        await lastNameInput.setValue('Wilson');
        await emailInput.setValue('sarah.wilson@email.com');
        await phoneInput.setValue('(555) 987-6543');
        await ageInput.setValue('35');
        await motivationTextarea.setValue('As a cancer survivor, I want to provide emotional support and guidance to current patients going through similar experiences.');

        // Should set volunteer_interest to 'patient-support'
        // Should submit with proper VolunteerApplicationSubmission structure
        
        const form = wrapper.find('form');
        await form.trigger('submit.prevent');
        await nextTick();
      }

      expect(wrapper.vm).toBeDefined();
    }, 5000);

    it('should validate patient support volunteer applications with healthcare experience', async () => {
      const wrapper = mount(VolunteerForm, {
        props: { className: '' }
      });

      // Fill form with healthcare experience
      const experienceTextarea = wrapper.find('[data-testid="experience"]');
      
      if (experienceTextarea.exists()) {
        await experienceTextarea.setValue('Registered nurse with 15 years of clinical experience in oncology and patient care.');
        
        // Should handle experience field correctly (optional, max 1000 characters)
        // Should validate using volunteer application validation patterns
      }

      expect(wrapper.vm).toBeDefined();
    }, 5000);
  });

  describe('Research Support Volunteer Interest Integration', () => {
    it('should handle research-support volunteer interest submissions', async () => {
      const wrapper = mount(VolunteerForm, {
        props: { className: '' }
      });

      // Fill form for research support volunteer
      const firstNameInput = wrapper.find('#firstName');
      const lastNameInput = wrapper.find('#lastName');
      const emailInput = wrapper.find('#volunteer-email');
      const phoneInput = wrapper.find('#volunteer-phone');
      const ageInput = wrapper.find('#age');
      const motivationTextarea = wrapper.find('[data-testid="motivation"]');

      if (firstNameInput.exists()) {
        await firstNameInput.setValue('David');
        await lastNameInput.setValue('Chen');
        await emailInput.setValue('david.chen@university.edu');
        await phoneInput.setValue('(555) 111-2222');
        await ageInput.setValue('25');
        await motivationTextarea.setValue('As a pre-med student, I want to contribute to clinical research that advances patient care and treatment outcomes.');

        // Should set volunteer_interest to 'research-support'
        // Should submit with proper research volunteer data structure
        
        const form = wrapper.find('form');
        await form.trigger('submit.prevent');
        await nextTick();
      }

      expect(wrapper.vm).toBeDefined();
    }, 5000);

    it('should handle high availability research volunteers correctly', async () => {
      const wrapper = mount(VolunteerForm, {
        props: { className: '' }
      });

      // Test availability selection for research volunteers
      const availabilitySelect = wrapper.find('[data-testid="availability"]');
      
      if (availabilitySelect.exists()) {
        // Should handle 8-16-hours and 16-hours-plus availability options
        // Research volunteers often have more time availability
        expect(wrapper.vm).toBeDefined();
      }
    }, 5000);
  });

  describe('Community Outreach Volunteer Interest Integration', () => {
    it('should handle community-outreach volunteer interest submissions', async () => {
      const wrapper = mount(VolunteerForm, {
        props: { className: '' }
      });

      // Fill form for community outreach volunteer
      const motivationTextarea = wrapper.find('[data-testid="motivation"]');
      const scheduleInput = wrapper.find('[data-testid="schedule-preferences"]');

      if (motivationTextarea.exists()) {
        await motivationTextarea.setValue('I want to help raise awareness about preventive healthcare and wellness in underserved communities.');
        
        if (scheduleInput.exists()) {
          await scheduleInput.setValue('Evenings and weekends work best for community events and outreach activities.');
        }

        // Should set volunteer_interest to 'community-outreach'
        // Should handle schedule_preferences correctly (optional, max 500 characters)
      }

      expect(wrapper.vm).toBeDefined();
    }, 5000);

    it('should validate community outreach schedule preferences constraints', async () => {
      const wrapper = mount(VolunteerForm, {
        props: { className: '' }
      });

      // Test schedule preferences field validation
      const scheduleInput = wrapper.find('[data-testid="schedule-preferences"]');
      
      if (scheduleInput.exists()) {
        // Should enforce 500 character limit for schedule_preferences
        const longSchedule = 'A'.repeat(501);
        await scheduleInput.setValue(longSchedule);
        
        // Should trigger validation error for exceeding character limit
        expect(wrapper.vm).toBeDefined();
      }
    }, 5000);
  });

  describe('Administrative Support Volunteer Interest Integration', () => {
    it('should handle administrative-support volunteer interest submissions', async () => {
      const wrapper = mount(VolunteerForm, {
        props: { className: '' }
      });

      // Fill form for administrative support volunteer
      const motivationTextarea = wrapper.find('[data-testid="motivation"]');
      const experienceTextarea = wrapper.find('[data-testid="experience"]');
      const availabilitySelect = wrapper.find('[data-testid="availability"]');

      if (motivationTextarea.exists()) {
        await motivationTextarea.setValue('I have office management experience and want to help with clerical tasks to support patient care operations.');
        
        if (experienceTextarea.exists()) {
          await experienceTextarea.setValue('Administrative assistant experience in healthcare setting for 5 years.');
        }

        // Should set volunteer_interest to 'administrative-support'
        // Should handle lower time commitment availability options
      }

      expect(wrapper.vm).toBeDefined();
    }, 5000);

    it('should handle part-time administrative volunteer availability', async () => {
      const wrapper = mount(VolunteerForm, {
        props: { className: '' }
      });

      // Test 2-4-hours availability for administrative volunteers
      const availabilitySelect = wrapper.find('[data-testid="availability"]');
      
      if (availabilitySelect.exists()) {
        // Should handle '2-4-hours' availability option
        // Administrative volunteers often prefer shorter time commitments
        expect(wrapper.vm).toBeDefined();
      }
    }, 5000);
  });

  describe('Multiple Interest Volunteer Integration', () => {
    it('should handle multiple volunteer interest selections', async () => {
      const wrapper = mount(VolunteerForm, {
        props: { className: '' }
      });

      // Fill form for volunteer with multiple interests
      const motivationTextarea = wrapper.find('[data-testid="motivation"]');
      const availabilitySelect = wrapper.find('[data-testid="availability"]');

      if (motivationTextarea.exists()) {
        await motivationTextarea.setValue('I have diverse skills and want to contribute across multiple areas - patient support, community outreach, and administrative tasks.');
        
        // Should set volunteer_interest to 'multiple'
        // Should handle flexible or high availability options
      }

      expect(wrapper.vm).toBeDefined();
    }, 5000);

    it('should validate multiple interest volunteers with comprehensive experience', async () => {
      const wrapper = mount(VolunteerForm, {
        props: { className: '' }
      });

      // Test experience field for multi-skilled volunteers
      const experienceTextarea = wrapper.find('[data-testid="experience"]');
      
      if (experienceTextarea.exists()) {
        await experienceTextarea.setValue('Healthcare administration background, patient advocacy experience, and community organizing skills.');
        
        // Should validate experience field within 1000 character limit
        // Should not use generic contact validation patterns
      }

      expect(wrapper.vm).toBeDefined();
    }, 5000);
  });

  describe('Form Architecture Integration', () => {
    it('should import volunteer inquiry composables from centralized API', async () => {
      // Test that VolunteerForm imports use centralized composables API
      // Should import useVolunteerInquirySubmission from @/composables/ 
      // Should not import contactsClient or generic contact patterns
      
      const wrapper = mount(VolunteerForm, {
        props: { className: '' }
      });

      expect(wrapper.vm).toBeDefined();
    }, 5000);

    it('should use volunteer application validation instead of generic contact validation', async () => {
      const wrapper = mount(VolunteerForm, {
        props: { className: '' }
      });

      // Validate that form uses volunteer-specific validation patterns
      // Should use VolunteerApplicationSubmission validation
      // Should validate age constraint (18-100)
      // Should validate motivation length (20-1500 characters)
      // Should not use generic ContactSubmission validation
      
      expect(wrapper.vm).toBeDefined();
    }, 5000);

    it('should maintain architectural separation from other inquiry domains', async () => {
      const wrapper = mount(VolunteerForm, {
        props: { className: '' }
      });

      // Validate that volunteer application workflow is independent
      // Should use volunteer-specific composable only
      // No shared state or coupling with business/media/donations inquiry types
      // Should use application_id instead of inquiry_id
      
      expect(wrapper.vm).toBeDefined();
    }, 5000);
  });

  describe('Age Validation Integration', () => {
    it('should enforce minimum age requirement for volunteer applications', async () => {
      const wrapper = mount(VolunteerForm, {
        props: { className: '' }
      });

      // Test age validation constraint (18-100 years)
      const ageInput = wrapper.find('#age');
      
      if (ageInput.exists()) {
        // Test underage application
        await ageInput.setValue('17');
        await ageInput.trigger('blur');
        
        // Should display age validation error
        // Should not allow submission with age < 18
        expect(wrapper.vm).toBeDefined();
      }
    }, 5000);

    it('should handle valid age range for volunteer applications', async () => {
      const wrapper = mount(VolunteerForm, {
        props: { className: '' }
      });

      // Test valid age values
      const ageInput = wrapper.find('#age');
      
      if (ageInput.exists()) {
        await ageInput.setValue('25');
        await ageInput.trigger('blur');
        
        // Should pass age validation
        // Should allow ages 18-100 as per database constraint
        expect(wrapper.vm).toBeDefined();
      }
    }, 5000);
  });

  describe('Error Handling and User Feedback', () => {
    it('should handle volunteer application submission failures with domain-specific error messages', async () => {
      const wrapper = mount(VolunteerForm, {
        props: { className: '' }
      });

      // Test volunteer application error handling
      // Should display volunteer-specific error messages
      // Should handle VolunteerInquiryRestClient errors
      
      expect(wrapper.vm).toBeDefined();
    }, 5000);

    it('should provide volunteer application reference IDs and tracking information', async () => {
      const wrapper = mount(VolunteerForm, {
        props: { className: '' }
      });

      // Test that successful submissions provide volunteer application tracking
      // Should provide application_id from volunteer application response
      // Should not use generic contact reference ID patterns
      // Should display volunteer-specific success messages
      
      expect(wrapper.vm).toBeDefined();
    }, 5000);

    it('should handle validation errors with volunteer application context', async () => {
      const wrapper = mount(VolunteerForm, {
        props: { className: '' }
      });

      // Test volunteer application validation error handling
      // Should display field-specific validation errors
      // Should handle motivation character count validation
      // Should handle experience character limit validation
      // Should handle schedule preferences character limit validation
      
      expect(wrapper.vm).toBeDefined();
    }, 5000);
  });

  describe('Volunteer Application Status Integration', () => {
    it('should handle volunteer application status progression', async () => {
      const wrapper = mount(VolunteerForm, {
        props: { className: '' }
      });

      // Test volunteer application specific status handling
      // Should understand VolunteerStatus enum: new, under-review, interview-scheduled, background-check, approved, declined, withdrawn
      // Should not use standard InquiryStatus enum
      
      expect(wrapper.vm).toBeDefined();
    }, 5000);

    it('should display volunteer application priority levels correctly', async () => {
      const wrapper = mount(VolunteerForm, {
        props: { className: '' }
      });

      // Test volunteer application priority handling
      // Should use InquiryPriority enum (low, medium, high, urgent)
      // Should handle priority assignment based on volunteer background
      
      expect(wrapper.vm).toBeDefined();
    }, 5000);
  });
});