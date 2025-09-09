<template>
  <div>
    <section class="relative h-[480px] md:h-[420px] lg:h-[580px] overflow-hidden">
      <!-- Container wrapper to constrain background on desktop -->
      <div class="container relative h-full">
        <!-- Solid Blue Background with Twinkling Shapes - Constrained to container -->
        <div class="absolute inset-0 bg-blue-900">
          <!-- Twinkling Shapes Animation -->
          <div class="absolute inset-0 overflow-hidden">
            <!-- Circles -->
            <div
              v-for="(circle, i) in twinklingCircles"
              :key="`circle-${i}`"
              class="w-1.5 h-1.5 bg-white rounded-full absolute"
              :style="circle.style"
            />

            <!-- Triangles -->
            <div
              v-for="(triangle, i) in twinklingTriangles"
              :key="`triangle-${i}`"
              class="w-0 h-0 absolute"
              :style="triangle.style"
            />

            <!-- Diamonds (Rotated Squares) -->
            <div
              v-for="(diamond, i) in twinklingDiamonds"
              :key="`diamond-${i}`"
              class="w-1.5 h-1.5 bg-white transform rotate-45 absolute"
              :style="diamond.style"
            />

            <!-- Squares -->
            <div
              v-for="(square, i) in twinklingSquares"
              :key="`square-${i}`"
              class="w-1.5 h-1.5 bg-white absolute"
              :style="square.style"
            />
          </div>
        </div>

        <!-- Dark Blue overlay for mobile/tablet - darker than gradient background -->
        <div class="absolute inset-0 bg-blue-900 lg:hidden" />

        <!-- Glassmorphism Background for Desktop - Positioned at container level for full height -->
        <div
          class="absolute inset-y-0 left-[4%] w-[52%] bg-black/10 backdrop-blur-3xl hidden lg:block"
        />

        <!-- Content wrapper - Full height with padding for vertical centering -->
        <div class="relative h-full flex items-center pb-16 md:pb-24 lg:pb-32">
          <!-- Main Content Area - Full coverage on mobile/tablet, slightly offset on desktop -->
          <div class="relative h-full w-full lg:w-[52%] lg:ml-[4%] flex items-center">
            <!-- Content - Centered accounting for card overlap -->
            <div class="relative px-8 pt-2 pb-8 lg:px-12 lg:pt-2 lg:pb-12 max-w-2xl">
              <h1 class="text-4xl lg:text-6xl font-bold text-white mb-8">
                Excellence in Regenerative Medicine
              </h1>
              <div class="flex flex-col sm:flex-row gap-4 relative z-10">
                <a
                  v-if="primaryCareServices.length > 0 || regenerativeTherapies.length > 0"
                  href="/services"
                  class="inline-block px-8 py-3 text-black bg-white hover:bg-gray-100 font-semibold rounded transition-colors text-center border border-gray-300"
                >
                  View All Services
                </a>
                <a
                  href="/appointment"
                  class="inline-block px-8 py-3 text-white bg-black hover:bg-gray-800 font-semibold rounded transition-colors text-center border border-black hover:border-gray-800"
                >
                  Book Appointment
                </a>
              </div>
            </div>
          </div>
        </div>
      </div>
    </section>

    <!-- Service Cards - Positioned to Overlap Hero Section -->
    <section class="relative -mt-16 md:-mt-24 lg:-mt-32 pb-20">
      <div class="container">
        <div class="grid grid-cols-1 md:grid-cols-2 gap-8 max-w-4xl mx-auto">
          <!-- Primary Category Card -->
          <div class="relative bg-white rounded-sm overflow-hidden border border-gray-200">
            <div class="p-4">
              <h3 class="text-lg font-bold text-gray-900 mb-3">{{ primaryCategoryName }}</h3>

              <div class="space-y-1">
                <template v-if="primaryCareServices.length > 0">
                  <a
                    v-for="(service, index) in primaryCareServices"
                    :key="index"
                    :href="service.url"
                    class="relative flex items-center rounded-sm px-4 py-3 transition-all duration-200 hover:bg-blue-50 group"
                  >
                    <!-- Vertical colored line -->
                    <div
                      class="absolute left-0 top-0 bottom-0 w-1 bg-transparent group-hover:bg-blue-600 transition-colors rounded-l-sm"
                    />

                    <!-- Text -->
                    <div class="ml-2">
                      <div
                        class="text-base font-medium text-gray-700 group-hover:text-gray-900 transition-colors"
                      >
                        {{ service.title }}
                      </div>
                    </div>
                  </a>
                </template>
                <div v-else class="text-center py-4">
                  <p class="text-gray-400 text-sm">Services temporarily unavailable</p>
                </div>
              </div>
            </div>
          </div>

          <!-- Secondary Category Card -->
          <div class="relative bg-white rounded-sm overflow-hidden border border-gray-200">
            <div class="p-4">
              <h3 class="text-lg font-bold text-gray-900 mb-3">{{ secondaryCategoryName }}</h3>

              <div class="space-y-1">
                <template v-if="regenerativeTherapies.length > 0">
                  <a
                    v-for="(service, index) in regenerativeTherapies"
                    :key="index"
                    :href="service.url"
                    class="relative flex items-center rounded-sm px-4 py-3 transition-all duration-200 hover:bg-blue-50 group"
                  >
                    <!-- Vertical colored line -->
                    <div
                      class="absolute left-0 top-0 bottom-0 w-1 bg-transparent group-hover:bg-blue-600 transition-colors rounded-l-sm"
                    />

                    <!-- Text -->
                    <div class="ml-2">
                      <div
                        class="text-base font-medium text-gray-700 group-hover:text-gray-900 transition-colors"
                      >
                        {{ service.title }}
                      </div>
                    </div>
                  </a>
                </template>
                <div v-else class="text-center py-4">
                  <p class="text-gray-400 text-sm">Services temporarily unavailable</p>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </section>
  </div>
</template>

<script setup lang="ts">
import { computed, ref, onMounted } from 'vue';

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

// Reactive state for hero data that can be updated on client side
const heroData = ref<HeroServicesData>({
  primaryCareServices: props.heroServicesData?.primaryCareServices || [],
  regenerativeTherapies: props.heroServicesData?.regenerativeTherapies || [],
  primaryCategoryName: props.heroServicesData?.primaryCategoryName || 'Featured Category 1',
  secondaryCategoryName: props.heroServicesData?.secondaryCategoryName || 'Featured Category 2',
});

// Use reactive data for computed properties
const primaryCareServices = computed(() => heroData.value.primaryCareServices);
const regenerativeTherapies = computed(() => heroData.value.regenerativeTherapies);
const primaryCategoryName = computed(() => heroData.value.primaryCategoryName);
const secondaryCategoryName = computed(() => heroData.value.secondaryCategoryName);

// Twinkling animation configuration
const animationConfigs = [
  { name: 'twinkleRare', initialOpacity: 0.08, weight: 40 },
  { name: 'twinkleDormant', initialOpacity: 0.05, weight: 25 },
  { name: 'twinkleOccasional', initialOpacity: 0.12, weight: 20 },
  { name: 'twinkleRegular', initialOpacity: 0.15, weight: 10 },
  { name: 'twinkleConstant', initialOpacity: 0.2, weight: 5 },
];

const durations = ['8s', '12s', '15s', '18s', '22s', '25s'];

// Function to create twinkling animation with natural distribution
const getTwinkleAnimation = () => {
  const totalWeight = animationConfigs.reduce((sum, config) => sum + config.weight, 0);
  let random = Math.random() * totalWeight;
  let selectedConfig = animationConfigs[0];

  for (const config of animationConfigs) {
    if (random < config.weight) {
      selectedConfig = config;
      break;
    }
    random -= config.weight;
  }

  const duration = durations[Math.floor(Math.random() * durations.length)];
  const durationMs = parseInt(duration) * 1000;

  // Random start point in animation cycle
  const randomStartPoint = Math.random();
  const negativeDelay = -Math.floor(randomStartPoint * durationMs);
  const delayString = `${negativeDelay}ms`;

  // Calculate start opacity based on cycle position
  let startOpacity = selectedConfig.initialOpacity;
  if (selectedConfig.name !== 'twinkleDormant') {
    const cyclePosition = randomStartPoint;
    if (selectedConfig.name === 'twinkleRare') {
      if (cyclePosition >= 0.88 && cyclePosition <= 0.95) {
        startOpacity = 0.6;
      }
    } else if (selectedConfig.name === 'twinkleOccasional') {
      if (cyclePosition >= 0.75 && cyclePosition <= 0.85) {
        startOpacity = 0.5;
      }
    } else if (selectedConfig.name === 'twinkleRegular') {
      if (cyclePosition >= 0.65 && cyclePosition <= 0.8) {
        startOpacity = 0.45;
      }
    } else if (selectedConfig.name === 'twinkleConstant') {
      if (cyclePosition >= 0.25 && cyclePosition <= 0.75) {
        startOpacity = 0.35;
      }
    }
  }

  return {
    opacity: startOpacity,
    animation: `${selectedConfig.name} ${duration} infinite ease-in-out both`,
    animationDelay: delayString,
    left: `${Math.random() * 95}%`,
    top: `${Math.random() * 95}%`,
  };
};

// Create twinkling elements
const twinklingCircles = ref<Array<{ style: any }>>([]);
const twinklingTriangles = ref<Array<{ style: any }>>([]);
const twinklingDiamonds = ref<Array<{ style: any }>>([]);
const twinklingSquares = ref<Array<{ style: any }>>([]);

onMounted(async () => {
  // If no meaningful data provided (client:only or empty data), load data on client side
  if (
    heroData.value.primaryCareServices.length === 0 &&
    heroData.value.regenerativeTherapies.length === 0
  ) {
    try {
      const { loadHeroServicesData } = await import('../../lib/navigation-data');
      const clientData = await loadHeroServicesData();
      console.log('🔍 [Feature136] Client-side data loaded:', clientData);

      // Update reactive state with new data
      heroData.value = clientData;
    } catch (error) {
      console.error('❌ [Feature136] Client-side data loading failed:', error);
    }
  }

  // Add CSS keyframes
  const style = document.createElement('style');
  style.textContent = `
    @keyframes twinkleRare {
      0%, 85% { opacity: 0.08; }
      88%, 95% { opacity: 0.6; }
      100% { opacity: 0.08; }
    }
    
    @keyframes twinkleOccasional {
      0%, 70% { opacity: 0.12; }
      75%, 85% { opacity: 0.5; }
      100% { opacity: 0.12; }
    }
    
    @keyframes twinkleRegular {
      0%, 60% { opacity: 0.15; }
      65%, 80% { opacity: 0.45; }
      100% { opacity: 0.15; }
    }
    
    @keyframes twinkleConstant {
      0%, 100% { opacity: 0.2; }
      50% { opacity: 0.35; }
    }
    
    @keyframes twinkleDormant {
      0%, 100% { opacity: 0.05; }
    }
  `;
  document.head.appendChild(style);

  // Generate circles
  twinklingCircles.value = Array.from({ length: 15 }, () => ({
    style: getTwinkleAnimation(),
  }));

  // Generate triangles
  twinklingTriangles.value = Array.from({ length: 12 }, () => {
    const animation = getTwinkleAnimation();
    return {
      style: {
        ...animation,
        borderLeft: '4px solid transparent',
        borderRight: '4px solid transparent',
        borderBottom: '6px solid rgb(255, 255, 255)',
      },
    };
  });

  // Generate diamonds
  twinklingDiamonds.value = Array.from({ length: 10 }, () => ({
    style: getTwinkleAnimation(),
  }));

  // Generate squares
  twinklingSquares.value = Array.from({ length: 8 }, () => ({
    style: getTwinkleAnimation(),
  }));
});
</script>
