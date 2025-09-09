// Environment Configuration Manager
// Detects current environment and provides appropriate configuration

import { localConfig } from './local';
import { stagingConfig } from './staging';
import { tojkuvStagingConfig } from './tojkuv-staging';
import { productionConfig } from './production';

// Environment detection function
function detectEnvironment(): 'local' | 'staging' | 'tojkuv-staging' | 'production' {
  // Check build-time environment variables first
  const nodeEnv = import.meta.env.NODE_ENV;
  const mode = import.meta.env.MODE;

  // Local development detection
  if (mode === 'development' || nodeEnv === 'development') {
    return 'local';
  }

  // Production detection
  if (nodeEnv === 'production') {
    // Check if we're on tojkuv-staging domain
    if (typeof window !== 'undefined') {
      const hostname = window.location.hostname;
      if (hostname.includes('tojkuv-staging') || hostname.includes('tojkuv-stg')) {
        return 'tojkuv-staging';
      }
      if (hostname.includes('staging') || hostname.includes('dev-')) {
        return 'staging';
      }
    }
    return 'production';
  }

  // Fallback to tojkuv-staging for our current setup
  return 'tojkuv-staging';
}

// Get configuration for current environment
function getEnvironmentConfig() {
  const env = detectEnvironment();

  switch (env) {
    case 'local':
      return localConfig;
    case 'staging':
      return stagingConfig;
    case 'tojkuv-staging':
      return tojkuvStagingConfig;
    case 'production':
      return productionConfig;
    default:
      console.warn(`Unknown environment: ${env}, falling back to tojkuv-staging`);
      return tojkuvStagingConfig;
  }
}

// Export current configuration
export const config = getEnvironmentConfig();

// Export individual configs for testing
export { localConfig, stagingConfig, tojkuvStagingConfig, productionConfig };

// Export type for configuration
export type EnvironmentConfig = typeof config;
export type Environment = 'local' | 'staging' | 'tojkuv-staging' | 'production';

// Utility function to check current environment
export function isLocal(): boolean {
  return config.environment === 'local';
}

export function isStaging(): boolean {
  return config.environment === 'staging';
}

export function isTojkuvStaging(): boolean {
  return config.environment === 'tojkuv-staging';
}

export function isProduction(): boolean {
  return config.environment === 'production';
}

// Development-only logging
if (config.environment === 'local') {
  console.log('🔧 Environment Configuration:', {
    environment: config.environment,
    domains: Object.keys(config.domains),
    features: config.features,
  });
}
