// Pinia Plugin for Astro + Vue Integration
// Configures Pinia state management for production use

import { createPinia } from 'pinia';
import type { App } from 'vue';

// Create and configure Pinia instance
export const pinia = createPinia();

// Vue plugin function for Astro integration
export function installPinia(app: App) {
  app.use(pinia);
}

// Export for direct use in components/composables
export default pinia;