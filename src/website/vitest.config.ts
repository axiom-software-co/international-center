import { defineConfig } from 'vitest/config';
import { resolve } from 'path';
import vue from '@vitejs/plugin-vue';

export default defineConfig({
  plugins: [vue()],
  test: {
    globals: true,
    environment: 'jsdom',
    setupFiles: ['./src/test/setup.ts'],
    include: ['src/**/*.{test,spec}.{js,mjs,cjs,ts,mts,cts,jsx,tsx}'],
    exclude: ['node_modules/**', 'dist/**', '.cache/**', 'e2e/**'],
    
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
    
    // Timeout controls - fail fast for unit tests  
    testTimeout: 5000,   // 5 seconds for unit tests (per axiom rules)
    hookTimeout: 2000,   // 2 seconds for setup/teardown
    teardownTimeout: 1000, // 1 second for cleanup
    
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
