<template>
  <div class="bg-red-500 text-white p-8 text-center">
    <h1 class="text-4xl font-bold mb-4">TEST HERO COMPONENT</h1>
    <p>Hero Services Data Debug:</p>
    <pre class="text-left text-xs bg-black p-4 mt-4 rounded">{{
      JSON.stringify(heroServicesData, null, 2)
    }}</pre>
  </div>
</template>

<script setup lang="ts">
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
</script>
