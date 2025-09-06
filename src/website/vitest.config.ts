import { defineConfig } from 'vitest/config';
import { resolve } from 'path';
import vue from '@vitejs/plugin-vue';

export default defineConfig({
  plugins: [vue()],
  test: {
    globals: true,
    environment: process.env.TEST_INTEGRATION ? 'node' : 'jsdom',
    setupFiles: process.env.TEST_INTEGRATION 
      ? ['./src/test/integration/setup.ts']
      : ['./src/test/setup.ts'],
    include: process.env.TEST_INTEGRATION 
      ? ['src/test/integration/**/*.{test,spec}.{js,mjs,cjs,ts,mts,cts,jsx,tsx}']
      : ['src/**/*.{test,spec}.{js,mjs,cjs,ts,mts,cts,jsx,tsx}'],
    exclude: process.env.TEST_INTEGRATION
      ? ['node_modules/**', 'dist/**', '.cache/**', 'e2e/**']
      : ['node_modules/**', 'dist/**', '.cache/**', 'e2e/**', 'src/test/integration/**'],
    
    // Resource management - limit memory usage and processes
    pool: 'threads',
    poolOptions: {
      threads: {
        maxThreads: 4,
        minThreads: 1,
        isolate: true,
        singleThread: false,
      }
    },
    
    // Timeout controls - different for integration vs unit tests
    testTimeout: process.env.TEST_INTEGRATION ? 30000 : 5000,   // 30s for integration, 5s for unit tests
    hookTimeout: process.env.TEST_INTEGRATION ? 10000 : 2000,   // 10s for integration, 2s for unit tests
    teardownTimeout: process.env.TEST_INTEGRATION ? 5000 : 1000, // 5s for integration, 1s for unit tests
    
    // Memory and performance optimization
    sequence: {
      concurrent: true,
      shuffle: false,
    },
    
    // Test isolation and cleanup
    isolate: true,
    clearMocks: true,
    restoreMocks: true,
    unstubEnvs: true,
    unstubGlobals: true,
    
    // Optimize for CI/memory constrained environments
    maxConcurrency: 4,
    bail: 10, // Stop after 10 failures to prevent resource waste
    
    // Coverage configuration (disable for performance)
    coverage: {
      enabled: false, // Only enable when needed
    },
  },
  resolve: {
    alias: {
      '@': resolve(__dirname, './src'),
      '@/components': resolve(__dirname, './src/components'),
      '@/lib': resolve(__dirname, './src/lib'),
      '@/data': resolve(__dirname, './src/data'),
    },
  },
  define: {
    'import.meta.env.VITEST': true,
  },
});
