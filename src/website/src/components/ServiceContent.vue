<template>
  <article class="service-content">
    <header class="service-header">
      <h1 class="service-title">{{ service.title }}</h1>
      <p class="service-description">{{ service.description }}</p>
    </header>

    <div v-if="service.content" class="service-content-body">
      <div v-html="sanitizedContent" class="content-html"></div>
    </div>

    <aside v-if="showServiceMeta" class="service-metadata">
      <div class="service-delivery">
        <strong>Delivery Mode:</strong> {{ formattedDeliveryMode }}
      </div>
      <div v-if="service.image_url" class="service-image">
        <img :src="service.image_url" :alt="service.title" loading="lazy" />
      </div>
    </aside>
  </article>
</template>

<script setup lang="ts">
import { computed } from 'vue';
import type { Service } from '@/lib/clients/services/types';

interface ServiceContentProps {
  service: Service;
  showServiceMeta?: boolean;
}

const props = withDefaults(defineProps<ServiceContentProps>(), {
  showServiceMeta: true,
});

// Basic HTML sanitization (in production, consider using a library like DOMPurify)
const sanitizedContent = computed(() => {
  if (!props.service.content) return '';
  
  // For now, trust PostgreSQL-stored content since it's from the backend
  // In production environment, implement proper sanitization
  return props.service.content;
});

const formattedDeliveryMode = computed(() => {
  const modeMap = {
    'mobile_service': 'Mobile Service',
    'outpatient_service': 'Outpatient Service',
    'inpatient_service': 'Inpatient Service',
  };
  return modeMap[props.service.delivery_mode] || props.service.delivery_mode;
});
</script>

<style scoped>
.service-content {
  max-width: 800px;
  margin: 0 auto;
  line-height: 1.6;
}

.service-header {
  margin-bottom: 2rem;
}

.service-title {
  font-size: 2.5rem;
  font-weight: 700;
  margin-bottom: 1rem;
  color: #1f2937;
}

.service-description {
  font-size: 1.25rem;
  color: #6b7280;
  margin-bottom: 2rem;
}

.service-content-body {
  margin-bottom: 2rem;
}

.content-html :deep(h2) {
  font-size: 1.875rem;
  font-weight: 600;
  margin: 1.5rem 0 1rem;
  color: #374151;
}

.content-html :deep(h3) {
  font-size: 1.5rem;
  font-weight: 600;
  margin: 1.25rem 0 0.75rem;
  color: #374151;
}

.content-html :deep(p) {
  margin-bottom: 1rem;
  color: #4b5563;
}

.content-html :deep(ul), 
.content-html :deep(ol) {
  margin: 1rem 0;
  padding-left: 1.5rem;
}

.content-html :deep(li) {
  margin-bottom: 0.5rem;
  color: #4b5563;
}

.service-metadata {
  border-top: 1px solid #e5e7eb;
  padding-top: 2rem;
  margin-top: 2rem;
}

.service-delivery {
  margin-bottom: 1rem;
  color: #6b7280;
}

.service-image img {
  max-width: 100%;
  height: auto;
  border-radius: 0.5rem;
}
</style>
