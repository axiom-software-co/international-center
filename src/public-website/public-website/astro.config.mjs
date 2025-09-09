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
    // split: true, // Removed - not a valid option in newer Astro versions
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
          assetFileNames: assetInfo => {
            // Use newer API for file extensions
            const ext =
              assetInfo.type === 'asset' && assetInfo.source
                ? assetInfo.fileName
                  ? assetInfo.fileName.split('.').pop()
                  : null
                : null;

            if (ext && /^(png|jpe?g|svg|gif|tiff|bmp|ico)$/i.test(ext)) {
              return `assets/images/[name].[hash].${ext}`;
            } else if (ext && /^(woff|woff2|eot|ttf|otf)$/i.test(ext)) {
              return `assets/fonts/[name].[hash].${ext}`;
            } else if (ext && /^css$/i.test(ext)) {
              return `assets/styles/[name].[hash].${ext}`;
            }
            return `assets/[name].[hash].[ext]`;
          },
        },
      },
      // Optimize for CDN compression
      reportCompressedSize: true,
      chunkSizeWarningLimit: 1000, // 1MB warning limit
    },
    server: {
      host: '0.0.0.0', // Expose to local network for mobile device testing
      port: 3000, // Match test expectations
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
  // Remove experimental features for now
  // experimental: {},
});
