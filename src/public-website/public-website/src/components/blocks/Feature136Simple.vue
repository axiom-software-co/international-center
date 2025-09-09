<template>
  <div>
    <section class="bg-blue-900 text-white p-8">
      <h1 class="text-4xl font-bold mb-4">Excellence in Regenerative Medicine</h1>
      <p>Testing - Primary Category: {{ primaryCategoryName }}</p>
      <p>Testing - Secondary Category: {{ secondaryCategoryName }}</p>
      <p>Primary services count: {{ primaryCareServices.length }}</p>
      <p>Secondary services count: {{ regenerativeTherapies.length }}</p>
    </section>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue';

interface Service {
  title: string;
  url: string;
}

interface HeroServicesData {
  primaryCareServices: Service[];
  regenerativeTherapies: Service[];
  primaryCategoryName: string;
  secondaryCategoryName: string;
}

interface Props {
  heroServicesData?: HeroServicesData;
}

const props = withDefaults(defineProps<Props>(), {
  heroServicesData: () => ({
    primaryCareServices: [],
    regenerativeTherapies: [],
    primaryCategoryName: 'Featured Category 1',
    secondaryCategoryName: 'Featured Category 2',
  }),
});

// Use database-driven services if available
const primaryCareServices = computed(() => props.heroServicesData?.primaryCareServices || []);
const regenerativeTherapies = computed(() => props.heroServicesData?.regenerativeTherapies || []);
const primaryCategoryName = computed(
  () => props.heroServicesData?.primaryCategoryName || 'Featured Category 1'
);
const secondaryCategoryName = computed(
  () => props.heroServicesData?.secondaryCategoryName || 'Featured Category 2'
);
</script>
