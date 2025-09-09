<template>
  <div :class="cn('rounded border bg-card p-6', props.class)">
    <h3 class="mb-4 text-lg font-semibold">Treatment Details</h3>
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

    <!-- Booking Buttons -->
    <div v-if="!isComingSoon" class="mt-6 pt-6 border-t border-gray-200">
      <div class="space-y-3">
        <a
          href="/appointment"
          :class="cn(buttonVariants({ size: 'lg' }), 'w-full')"
        >
          Book Appointment
        </a>

        <a
          href="/services"
          :class="cn(buttonVariants({ variant: 'outline', size: 'lg' }), 'w-full')"
        >
          View All Services
        </a>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue';
import { Clock, Calendar, MapPin } from 'lucide-vue-next';
import { cva } from 'class-variance-authority';
import { cn } from '@/lib/vue-utils';
import Separator from '@/components/vue-ui/Separator.vue';

// Import the button variants directly
const buttonVariants = cva(
  'inline-flex items-center justify-center whitespace-nowrap rounded-md text-sm font-medium ring-offset-background transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:pointer-events-none disabled:opacity-50',
  {
    variants: {
      variant: {
        default: 'bg-primary text-primary-foreground hover:bg-primary/90',
        destructive: 'bg-destructive text-destructive-foreground hover:bg-destructive/90',
        outline: 'border border-input bg-background hover:bg-accent hover:text-accent-foreground',
        secondary: 'bg-secondary text-secondary-foreground hover:bg-secondary/80',
        ghost: 'hover:bg-accent hover:text-accent-foreground',
        link: 'text-primary underline-offset-4 hover:underline',
      },
      size: {
        default: 'h-10 px-4 py-2',
        sm: 'h-9 rounded-md px-3',
        lg: 'h-11 rounded-md px-8',
        icon: 'h-10 w-10',
      },
    },
    defaultVariants: {
      variant: 'default',
      size: 'default',
    },
  }
);

interface Props {
  duration?: string;
  recovery?: string;
  deliveryModes?: string[];
  class?: string;
  isComingSoon?: boolean;
}

const props = withDefaults(defineProps<Props>(), {
  duration: '45-90 minutes',
  recovery: 'Minimal to no downtime',
  deliveryModes: () => ['outpatient'],
  isComingSoon: false,
});

const formatDeliveryModes = (modes: string[]): string => {
  return modes.map(mode => mode.charAt(0).toUpperCase() + mode.slice(1)).join(', ');
};

const details = computed(() => [
  {
    icon: MapPin,
    label: 'Delivery Options',
    value: formatDeliveryModes(props.deliveryModes),
  },
  {
    icon: Clock,
    label: 'Duration',
    value: props.duration,
  },
  {
    icon: Calendar,
    label: 'Recovery',
    value: props.recovery,
  },
]);
</script>