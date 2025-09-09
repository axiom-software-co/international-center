#!/usr/bin/env node
/**
 * Contract Test Runner - Main CLI for running contract tests
 */

import { readFile } from 'fs/promises';
import { load } from 'js-yaml';
import { ContractValidator, generateTestReport, type ContractTestResult } from './contract-validator';
import type { OpenAPIV3_1 } from 'openapi-types';

interface TestRunnerConfig {
  publicApiUrl: string;
  adminApiUrl: string;
  timeout?: number;
  verbose?: boolean;
}

class ContractTestRunner {
  private config: TestRunnerConfig;

  constructor(config: TestRunnerConfig) {
    this.config = {
      timeout: 15000,
      verbose: false,
      ...config
    };
  }

  async runAllTests(): Promise<void> {
    console.log('🚀 Starting Contract Test Suite');
    console.log('=================================');

    const results: ContractTestResult[] = [];
    
    try {
      // Test Public API
      console.log('\n📋 Testing Public API...');
      const publicResults = await this.testPublicAPI();
      results.push(...publicResults);

      // Test Admin API  
      console.log('\n🔐 Testing Admin API...');
      const adminResults = await this.testAdminAPI();
      results.push(...adminResults);

      // Generate and display report
      console.log('\n📊 Test Results');
      console.log('================');
      const report = generateTestReport(results);
      console.log(report);

      // Exit with appropriate code
      const failedTests = results.filter(r => !r.passed);
      if (failedTests.length > 0) {
        console.error(`❌ ${failedTests.length} test(s) failed`);
        process.exit(1);
      } else {
        console.log('✅ All tests passed!');
        process.exit(0);
      }
    } catch (error) {
      console.error('💥 Test runner failed:', error);
      process.exit(1);
    }
  }

  private async testPublicAPI(): Promise<ContractTestResult[]> {
    const spec = await this.loadSpec('../openapi/public-api.yaml');
    const validator = new ContractValidator(spec, {
      baseUrl: this.config.publicApiUrl,
      timeout: this.config.timeout
    });

    const results: ContractTestResult[] = [];

    // Test health endpoint
    if (this.config.verbose) console.log('  - Testing health endpoint');
    results.push(await validator.testHealthEndpoint());

    // Test core endpoints with sample data
    const sampleData = {
      mediaInquiry: this.generateSampleData('mediaInquiry'),
      businessInquiry: this.generateSampleData('businessInquiry'),
      donationInquiry: this.generateSampleData('donationInquiry'),
      volunteerInquiry: this.generateSampleData('volunteerInquiry')
    };

    // Test services endpoints
    if (this.config.verbose) console.log('  - Testing services endpoints');
    results.push(await validator.testEndpoint('/api/v1/services', 'GET', { page: 1, limit: 10 }));
    results.push(await validator.testEndpoint('/api/v1/services/featured', 'GET'));

    // Test news endpoints
    if (this.config.verbose) console.log('  - Testing news endpoints');
    results.push(await validator.testEndpoint('/api/v1/news', 'GET', { page: 1, limit: 10 }));
    results.push(await validator.testEndpoint('/api/v1/news/featured', 'GET'));

    // Test research endpoints
    if (this.config.verbose) console.log('  - Testing research endpoints');
    results.push(await validator.testEndpoint('/api/v1/research', 'GET', { page: 1, limit: 10 }));

    // Test events endpoints
    if (this.config.verbose) console.log('  - Testing events endpoints');
    results.push(await validator.testEndpoint('/api/v1/events', 'GET', { page: 1, limit: 10 }));

    // Test inquiry submissions
    if (this.config.verbose) console.log('  - Testing inquiry submissions');
    results.push(await validator.testEndpoint('/api/v1/inquiries/media', 'POST', sampleData.mediaInquiry, 201));
    results.push(await validator.testEndpoint('/api/v1/inquiries/business', 'POST', sampleData.businessInquiry, 201));

    return results;
  }

  private async testAdminAPI(): Promise<ContractTestResult[]> {
    const spec = await this.loadSpec('../openapi/admin-api.yaml');
    const validator = new ContractValidator(spec, {
      baseUrl: this.config.adminApiUrl,
      timeout: this.config.timeout
    });

    const results: ContractTestResult[] = [];

    // Test health endpoint
    if (this.config.verbose) console.log('  - Testing health endpoint');
    results.push(await validator.testHealthEndpoint());

    // Test authentication endpoints (without valid credentials)
    if (this.config.verbose) console.log('  - Testing authentication endpoints');
    results.push(await validator.testEndpoint('/api/v1/auth/login', 'POST', {
      email: 'test@example.com',
      password: 'testpassword'
    }, 401)); // Expect 401 for invalid credentials

    // Test protected endpoints (should all return 401)
    if (this.config.verbose) console.log('  - Testing protected endpoints');
    results.push(await validator.testEndpoint('/api/v1/users', 'GET', undefined, 401));
    results.push(await validator.testEndpoint('/api/v1/news', 'POST', {
      title: 'Test Article',
      summary: 'Test summary',
      content: 'Test content'
    }, 401));

    return results;
  }

  private async loadSpec(specPath: string): Promise<OpenAPIV3_1.Document> {
    try {
      const specContent = await readFile(specPath, 'utf-8');
      return load(specContent) as OpenAPIV3_1.Document;
    } catch (error) {
      throw new Error(`Failed to load OpenAPI spec from ${specPath}: ${error}`);
    }
  }

  private generateSampleData(type: string): any {
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
            subject: 'Contract Test Media Inquiry',
            message: 'This is a test media inquiry for contract validation.',
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

      default:
        return {};
    }
  }
}

// CLI entry point
if (require.main === module) {
  const config: TestRunnerConfig = {
    publicApiUrl: process.env.PUBLIC_API_BASE_URL || 'http://localhost:4000',
    adminApiUrl: process.env.ADMIN_API_BASE_URL || 'http://localhost:4001',
    timeout: parseInt(process.env.TEST_TIMEOUT || '15000'),
    verbose: process.env.VERBOSE === 'true' || process.argv.includes('--verbose')
  };

  const runner = new ContractTestRunner(config);
  runner.runAllTests().catch(console.error);
}