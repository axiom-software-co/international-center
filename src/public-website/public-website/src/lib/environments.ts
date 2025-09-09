// Environment Configuration
// Provides environment-specific settings and feature flags

export interface EnvironmentConfig {
  api: {
    baseUrl: string;
  };
}

export type Environment = 'local' | 'staging' | 'production';

export const config: EnvironmentConfig = {
  api: {
    baseUrl: process.env.API_BASE_URL || 'http://localhost:3000'
  }
};

export const isLocal = process.env.NODE_ENV === 'development' || process.env.NODE_ENV === 'test';
export const isStaging = process.env.NODE_ENV === 'staging';
export const isProduction = process.env.NODE_ENV === 'production';