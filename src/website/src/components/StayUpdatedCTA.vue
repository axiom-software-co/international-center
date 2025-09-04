<template>
  <a
    :href="href"
    :class="
      cn(
        'block rounded border p-6 transition-all focus:outline-none focus:ring-2 focus:ring-offset-2',
        bgClass,
        hoverBgClass,
        borderClass,
        hoverBorderClass,
        `focus:ring-${getColorClasses(theme, 'dark')}`,
        className
      )
    "
  >
    <div class="text-center">
      <h3 :class="cn('text-lg font-semibold mb-2', titleClass)">{{ title }}</h3>
      <p :class="cn('text-sm mb-4', textClass)">{{ description }}</p>
      <div
        :class="
          cn(
            'inline-flex items-center justify-center rounded px-6 py-2 text-sm font-medium text-white transition-colors',
            buttonBgClass,
            buttonHoverClass
          )
        "
      >
        {{ buttonText }}
      </div>
    </div>
  </a>
</template>

<script setup lang="ts">
import { computed } from 'vue';
import { cn } from '@/lib/utils';
import type { ColorTheme } from '@/lib/theme';
import { getColorClasses } from '@/lib/theme';

interface StayUpdatedCTAProps {
  title: string;
  description: string;
  buttonText: string;
  href?: string;
  theme?: ColorTheme;
  className?: string;
}

const props = withDefaults(defineProps<StayUpdatedCTAProps>(), {
  href: '/company/contact',
  theme: 'blue',
  className: '',
});

const bgClass = computed(() => `bg-${getColorClasses(props.theme, 'light')}`);
const hoverBgClass = computed(() => `hover:bg-${getColorClasses(props.theme, 'bg')}`);
const borderClass = computed(() => 'border');
const hoverBorderClass = computed(() => `hover:border-${getColorClasses(props.theme, 'border')}`);
const titleClass = computed(() => `text-${getColorClasses(props.theme, 'dark')}`);
const textClass = computed(() => `text-${getColorClasses(props.theme, 'text')}`);
const buttonBgClass = computed(() => `bg-${getColorClasses(props.theme, 'dark')}`);
const buttonHoverClass = computed(() => `hover:bg-${getColorClasses(props.theme, 'medium')}`);
</script>
