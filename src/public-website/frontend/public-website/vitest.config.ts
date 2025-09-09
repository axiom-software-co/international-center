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
    
    // Unit test timeouts - fast failure for proper unit test isolation
    testTimeout: 5000,     // 5s maximum for unit tests
    hookTimeout: 2000,     // 2s for setup/teardown hooks
    teardownTimeout: 1000, // 1s for cleanup
    
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
    unstubGlobals: false,
    
    // Optimize for unit test performance
    maxConcurrency: 4,
    bail: 10, // Stop after 10 failures
    
    // Coverage configuration (disable by default for faster unit tests)
    coverage: {
      enabled: false,
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
