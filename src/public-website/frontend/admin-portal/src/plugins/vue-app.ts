import type { App } from 'vue';
import { createPinia } from 'pinia';

export default function (app: App) {
  const pinia = createPinia();
  app.use(pinia);
}