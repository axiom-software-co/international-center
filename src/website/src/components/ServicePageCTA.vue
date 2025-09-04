<template>
  <div class="rounded border bg-card p-6">
    <template v-if="isComingSoon">
      <h3 class="text-lg font-semibold mb-3">Coming Soon</h3>
      <p class="text-sm text-muted-foreground">
        {{ serviceName }} will be available at our future clinic location. We're working to bring
        you the most advanced treatments.
      </p>
    </template>

    <template v-else>
      <h3 class="text-lg font-semibold mb-3">Ready to Get Started?</h3>

      <p class="text-sm text-muted-foreground mb-4">
        Schedule your appointment for {{ serviceName }} and begin your journey to optimal health.
      </p>

      <div class="space-y-3">
        <Button class="w-full" size="lg" as-child>
          <a href="/appointment">Book Appointment</a>
        </Button>

        <Button variant="outline" class="w-full" size="lg" as-child>
          <a href="/services">View All Services</a>
        </Button>
      </div>
    </template>
  </div>
</template>

<script setup lang="ts">
import { Button } from '@/components/vue-ui';

interface ServicePageCTAProps {
  serviceName?: string;
  isComingSoon?: boolean;
}

withDefaults(defineProps<ServicePageCTAProps>(), {
  serviceName: 'this treatment',
  isComingSoon: false,
});
</script>
