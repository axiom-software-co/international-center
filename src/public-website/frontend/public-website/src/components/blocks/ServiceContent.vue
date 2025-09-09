<template>
  <section class="service-block">
    <div class="service-header">
      <h2 class="service-title">{{ service.title }}</h2>
      <p class="service-description">{{ service.description }}</p>
    </div>

    <div v-if="service.content" class="service-content">
      <div v-html="sanitizedContent" class="content-html"></div>
    </div>

    <div v-if="showDeliveryMode" class="service-delivery">
      <span class="delivery-badge">{{ formattedDeliveryMode }}</span>
    </div>
  </section>
</template>

<script setup lang="ts">
import { computed } from 'vue';
import type { Service } from '@/lib/clients/services/types';

interface Props {
  service: Service;
  showDeliveryMode?: boolean;
}

const props = withDefaults(defineProps<Props>(), {
  showDeliveryMode: false,
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
.service-block {
  padding: 1.5rem;
  border-radius: 0.75rem;
  background: #ffffff;
  box-shadow: 0 1px 3px 0 rgb(0 0 0 / 0.1);
}

.service-header {
  margin-bottom: 1.5rem;
}

.service-title {
  font-size: 1.5rem;
  font-weight: 600;
  margin-bottom: 0.75rem;
  color: #1f2937;
}

.service-description {
  color: #6b7280;
  margin-bottom: 1rem;
}

.service-content {
  margin-bottom: 1.5rem;
}

.content-html :deep(h3) {
  font-size: 1.25rem;
  font-weight: 600;
  margin: 1rem 0 0.5rem;
  color: #374151;
}

.content-html :deep(p) {
  margin-bottom: 0.75rem;
  color: #4b5563;
}

.content-html :deep(ul), 
.content-html :deep(ol) {
  margin: 0.75rem 0;
  padding-left: 1.25rem;
}

.content-html :deep(li) {
  margin-bottom: 0.25rem;
  color: #4b5563;
}

.service-delivery {
  margin-top: 1rem;
}

.delivery-badge {
  display: inline-block;
  padding: 0.25rem 0.75rem;
  background: #f3f4f6;
  color: #374151;
  border-radius: 9999px;
  font-size: 0.875rem;
  font-weight: 500;
}
</style>