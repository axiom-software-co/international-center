<template>
  <div :class="cn('mb-6 rounded border p-4', bgClass, borderClass, className)">
    <p :class="cn('text-sm mb-2 font-medium', titleClass)">{{ title }}</p>
    <p :class="cn('text-sm', textClass)">{{ description }}</p>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue';
import { cn } from '@/lib/utils';
import type { ColorTheme } from '@/lib/theme';
import { getColorClasses } from '@/lib/theme';

interface ServiceComingSoonBannerProps {
  title: string;
  description: string;
  theme?: ColorTheme;
  className?: string;
}

const props = withDefaults(defineProps<ServiceComingSoonBannerProps>(), {
  theme: 'blue',
});

const bgClass = computed(() => `bg-${getColorClasses(props.theme, 'light')}`);
const borderClass = computed(() => `border-${getColorClasses(props.theme, 'border')}`);
const titleClass = computed(() => `text-${getColorClasses(props.theme, 'text')}`);
const textClass = computed(() => `text-${getColorClasses(props.theme, 'text')}`);
</script>
