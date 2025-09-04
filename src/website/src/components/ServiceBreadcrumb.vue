<template>
  <nav :class="cn('flex items-center gap-2 text-sm text-muted-foreground mb-4', className)">
    <template v-for="(item, index) in items" :key="index">
      <span v-if="index > 0">/</span>
      <a
        v-if="item.href"
        :href="item.href"
        class="flex items-center gap-1 hover:text-foreground transition-colors"
      >
        <component v-if="index === 0" :is="HomeIcon" class="h-4 w-4" />
        {{ item.label }}
      </a>
      <span v-else class="flex items-center gap-1">
        <component v-if="index === 0" :is="HomeIcon" class="h-4 w-4" />
        {{ item.label }}
      </span>
    </template>
    <span>/</span>
    <span class="text-foreground">{{ serviceName }}</span>
  </nav>
</template>

<script setup lang="ts">
import { cn } from '@/lib/utils';
import { ChevronRight, Home } from 'lucide-vue-next';


interface BreadcrumbItem {
  label: string;
  href?: string;
}

interface ServiceBreadcrumbProps {
  serviceName: string;
  items?: BreadcrumbItem[];
  className?: string;
}

withDefaults(defineProps<ServiceBreadcrumbProps>(), {
  items: () => [
    { label: 'Home', href: '/' },
    { label: 'Services', href: '/services' },
  ],
});
</script>
