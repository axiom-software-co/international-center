// Vue App Entrypoint for Astro + Pinia Integration
// Configures Vue app with Pinia state management

import type { App } from 'vue';
import { installPinia } from './pinia';

export default function (app: App) {
  // Install Pinia state management
  installPinia(app);
  
  // Additional Vue plugins can be installed here
  
  return app;
}