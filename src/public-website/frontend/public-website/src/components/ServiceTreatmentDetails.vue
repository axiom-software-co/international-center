<template>
  <div :class="cn('rounded border bg-card p-6', className)">
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
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue';
import { Separator } from '@/components/vue-ui';
import { cn } from '@/lib/utils';
import { CheckCircle, Info } from 'lucide-vue-next';


interface TreatmentDetail {
  icon: any;
  label: string;
  value: string;
}

interface ServiceTreatmentDetailsProps {
  duration: string;
  availability: string;
  safety: string;
  recovery: string;
  className?: string;
}

const props = defineProps<ServiceTreatmentDetailsProps>();

const details = computed<TreatmentDetail[]>(() => [
  {
    icon: Clock,
    label: 'Duration',
    value: props.duration,
  },
  {
    icon: Users,
    label: 'Availability',
    value: props.availability,
  },
  {
    icon: Shield,
    label: 'Safety',
    value: props.safety,
  },
  {
    icon: Calendar,
    label: 'Recovery',
    value: props.recovery,
  },
]);
</script>
