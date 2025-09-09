<template>
  <div :class="cn('rounded border bg-card p-6', props.class)">
    <h3 class="mb-4 text-lg font-semibold">Article Details</h3>
    <div class="space-y-3">
      <template v-for="(detail, index) in details" :key="detail.label">
        <div class="flex items-center gap-3">
          <component :is="detail.icon" class="h-5 w-5 text-muted-foreground" />
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
import { Calendar, User, Clock } from 'lucide-vue-next';
import { cn } from '@/lib/vue-utils';
import Separator from '@/components/vue-ui/Separator.vue';

interface Props {
  publishedAt?: string;
  author?: string;
  readTime?: string;
  class?: string;
}

const props = withDefaults(defineProps<Props>(), {
  publishedAt: 'Recently',
  author: 'International Center Team',
  readTime: '5 min read',
});

const details = computed(() => [
  {
    icon: Calendar,
    label: 'Published',
    value: props.publishedAt,
  },
  {
    icon: User,
    label: 'Author',
    value: props.author,
  },
  {
    icon: Clock,
    label: 'Read Time',
    value: props.readTime,
  },
]);
</script>