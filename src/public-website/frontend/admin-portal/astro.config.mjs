// @ts-check
import { defineConfig } from 'astro/config';
import vue from '@astrojs/vue';
import tailwind from '@astrojs/tailwind';
import node from '@astrojs/node';

// https://astro.build/config
export default defineConfig({
  output: 'server', // Enable server-side rendering with dynamic routes
  adapter: node({
    mode: 'standalone'
  }),
  integrations: [
    vue({
      appEntrypoint: '/src/plugins/vue-app.ts'
    }),
    tailwind({
      applyBaseStyles: false, // We'll handle base styles ourselves for shadcn/ui
    }),
  ],
  build: {
    inlineStylesheets: 'never', // Keep CSS external for better caching
    assets: '_astro',
    assetsPrefix: '/_astro/', // Explicit assets prefix for CDN
  },
  compressHTML: true,
  prefetch: {
    prefetchAll: false,
    defaultStrategy: 'hover',
  },
  vite: {
    build: {
      cssCodeSplit: true,
      minify: 'esbuild',
      target: 'es2020', // Modern target for better optimization
      rollupOptions: {
        output: {
          // Enhanced chunking strategy for optimal CDN caching
          manualChunks: {
            'vue-vendor': ['vue', '@vueuse/core', 'pinia'],
            'ui-vendor': ['lucide-vue-next', 'radix-vue'],
            utils: ['clsx', 'tailwind-merge'],
          },
          // Generate consistent hashed filenames for immutable caching
          entryFileNames: 'assets/[name].[hash].js',
          chunkFileNames: 'assets/[name].[hash].js',
          assetFileNames: 'assets/[name].[hash].[ext]',
        },
      },
      // Optimize for CDN compression
      reportCompressedSize: true,
      chunkSizeWarningLimit: 1000, // 1MB warning limit
    },
    server: {
      host: '0.0.0.0', // Expose to local network for mobile device testing
      port: 3001, // Different port for admin portal
    },
    // Optimize dependencies for better CDN caching
    optimizeDeps: {
      include: ['vue', '@vueuse/core', 'pinia', 'lucide-vue-next', 'radix-vue'],
      exclude: [],
    },
  },
  // CDN-optimized image settings
  image: {
    service: {
      entrypoint: 'astro/assets/services/sharp',
      config: {
        limitInputPixels: false,
      },
    },
  },
});