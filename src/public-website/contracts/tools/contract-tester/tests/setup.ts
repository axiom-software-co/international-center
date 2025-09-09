/**
 * Test setup for contract testing framework
 */

import { beforeAll, afterAll, vi } from 'vitest';

// Global test configuration
beforeAll(async () => {
  // Set test environment variables
  process.env.NODE_ENV = 'test';
  process.env.PUBLIC_API_BASE_URL = process.env.PUBLIC_API_BASE_URL || 'http://localhost:4000';
  process.env.ADMIN_API_BASE_URL = process.env.ADMIN_API_BASE_URL || 'http://localhost:4001';
  
  // Mock console.log in tests to reduce noise
  vi.spyOn(console, 'log').mockImplementation(() => {});
});

afterAll(() => {
  vi.restoreAllMocks();
});

// Global test utilities
declare global {
  var testUtils: {
    waitForService: (url: string, timeout?: number) => Promise<boolean>;
    generateTestData: (type: string) => any;
  };
}

globalThis.testUtils = {
  /**
   * Wait for a service to be available
   */
  async waitForService(url: string, timeout: number = 30000): Promise<boolean> {
    const startTime = Date.now();
    
    while (Date.now() - startTime < timeout) {
      try {
        const response = await fetch(`${url}/health`);
        if (response.ok) {
          return true;
        }
      } catch (error) {
        // Service not ready yet
      }
      
      // Wait 1 second before next attempt
      await new Promise(resolve => setTimeout(resolve, 1000));
    }
    
    return false;
  },
  
  /**
   * Generate test data for different types
   */
  generateTestData(type: string): any {
    switch (type) {
      case 'mediaInquiry':
        return {
          contactInfo: {
            name: 'Test Reporter',
            email: 'reporter@test.com',
            phone: '+1-555-0123',
            organization: 'Test News'
          },
          inquiryDetails: {
            subject: 'Test Media Inquiry',
            message: 'This is a test media inquiry for contract testing.',
            deadline: new Date(Date.now() + 7 * 24 * 60 * 60 * 1000).toISOString(),
            preferredContactMethod: 'email'
          }
        };
        
      case 'businessInquiry':
        return {
          contactInfo: {
            name: 'Test Business Owner',
            email: 'business@test.com',
            phone: '+1-555-0124',
            organization: 'Test Corp'
          },
          inquiryDetails: {
            subject: 'Partnership Inquiry',
            message: 'Interested in business partnership opportunities.',
            serviceInterest: 'consultation',
            budget: 'under-10k'
          }
        };
        
      case 'donationInquiry':
        return {
          donorInfo: {
            name: 'Test Donor',
            email: 'donor@test.com',
            phone: '+1-555-0125'
          },
          donationDetails: {
            amount: 100.00,
            frequency: 'one-time',
            cause: 'general',
            isAnonymous: false
          }
        };
        
      case 'volunteerInquiry':
        return {
          personalInfo: {
            name: 'Test Volunteer',
            email: 'volunteer@test.com',
            phone: '+1-555-0126',
            dateOfBirth: '1990-01-01',
            address: {
              street: '123 Test St',
              city: 'Test City',
              state: 'TS',
              postalCode: '12345',
              country: 'US'
            }
          },
          availability: {
            daysAvailable: ['monday', 'wednesday', 'friday'],
            timesAvailable: ['morning', 'afternoon'],
            hoursPerWeek: 10
          },
          interests: ['healthcare', 'education'],
          experience: 'Some volunteer experience with local organizations.',
          motivation: 'Want to help the community and gain experience.'
        };
        
      case 'eventRegistration':
        return {
          participantInfo: {
            name: 'Test Participant',
            email: 'participant@test.com',
            phone: '+1-555-0127'
          },
          registrationDetails: {
            numberOfAttendees: 1,
            specialRequirements: 'No special requirements',
            dietaryRestrictions: 'none'
          }
        };
        
      default:
        return {};
    }
  }
};