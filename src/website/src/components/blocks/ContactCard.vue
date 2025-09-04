<template>
  <div :class="cn('rounded border bg-card p-6', props.class)">
    <h3 class="mb-4 text-lg font-semibold">Media Inquiries?</h3>
    <div class="space-y-3">
      <Button variant="outline" class="w-full rounded-sm" asChild>
        <a href="/company/contact" class="flex items-center justify-center">
          Contact Us
        </a>
      </Button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { cn } from '@/lib/vue-utils';
import Button from '@/components/vue-ui/Button.vue';

interface Props {
  class?: string;
}

const props = defineProps<Props>();
</script>