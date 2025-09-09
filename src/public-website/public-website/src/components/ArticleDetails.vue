<template>
  <div :class="cn('rounded border bg-card p-6', className)">
    <h3 class="mb-4 text-lg font-semibold">{{ title }}</h3>
    <div class="space-y-3">
      <template v-for="(detail, index) in details" :key="detail.label">
        <div class="flex items-center gap-3">
          <component :is="detail.icon" />
          <div>
            <p class="text-sm font-medium">{{ detail.label }}</p>
            <p class="text-sm text-muted-foreground">{{ detail.value }}</p>
          </div>
        </div>
        <Separator v-if="index < details.length - 1" />
      </template>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue';
import { Calendar, User, Clock, Tag } from 'lucide-vue-next';
import Separator from '@/components/vue-ui/Separator.vue';
import { cn } from '@/lib/utils';



interface ArticleDetail {
  icon: any;
  label: string;
  value: string;
}

interface ArticleDetailsProps {
  title: string;
  items: {
    label: string;
    value: string;
  }[];
  className?: string;
}

const props = defineProps<ArticleDetailsProps>();

// Map label to appropriate icon
const getIcon = (label: string) => {
  const normalizedLabel = label.toLowerCase();
  if (normalizedLabel.includes('published') || normalizedLabel.includes('date')) {
    return Calendar;
  }
  if (normalizedLabel.includes('author')) {
    return User;
  }
  if (normalizedLabel.includes('read') || normalizedLabel.includes('time')) {
    return Clock;
  }
  if (normalizedLabel.includes('category') || normalizedLabel.includes('type')) {
    return FileText;
  }
  // Default icon
  return FileText;
};

const details = computed((): ArticleDetail[] => {
  return props.items.map(item => ({
    icon: getIcon(item.label),
    label: item.label,
    value: item.value,
  }));
});
</script>
