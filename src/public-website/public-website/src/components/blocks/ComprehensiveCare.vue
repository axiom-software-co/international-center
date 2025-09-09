<template>
  <section class="pt-16 pb-20">
    <div class="container">
      <div class="text-center mb-16">
        <h2 class="text-3xl font-semibold lg:text-5xl mb-4">Precision Medicine Approach</h2>
      </div>

      <div class="grid gap-12 lg:grid-cols-2 lg:gap-16 items-center">
        <!-- Left Content -->
        <div>
          <!-- Feature List -->
          <div class="space-y-6">
            <div class="flex gap-4">
              <div class="flex-shrink-0">
                <span class="flex size-12 items-center justify-center rounded-sm bg-gray-100">
                  <ClipboardList class="size-6 text-gray-700" />
                </span>
              </div>
              <div>
                <h3 class="text-lg font-semibold mb-2">Comprehensive Assessment</h3>
                <p class="text-gray-600">
                  Advanced diagnostic testing, biomarker analysis, and detailed health evaluations
                  to establish baseline metrics.
                </p>
              </div>
            </div>

            <div class="flex gap-4">
              <div class="flex-shrink-0">
                <span class="flex size-12 items-center justify-center rounded-sm bg-gray-100">
                  <Sparkles class="size-6 text-gray-700" />
                </span>
              </div>
              <div>
                <h3 class="text-lg font-semibold mb-2">Personalized Protocols</h3>
                <p class="text-gray-600">
                  Custom treatment plans designed from individual assessment data, genetic profiles,
                  and specific health goals.
                </p>
              </div>
            </div>

            <div class="flex gap-4">
              <div class="flex-shrink-0">
                <span class="flex size-12 items-center justify-center rounded-sm bg-gray-100">
                  <TestTube class="size-6 text-gray-700" />
                </span>
              </div>
              <div>
                <h3 class="text-lg font-semibold mb-2">Active Interventions</h3>
                <p class="text-gray-600">
                  Targeted regenerative therapies including PRP, exosomes, peptide protocols, and IV
                  nutritional treatments.
                </p>
              </div>
            </div>

            <div class="flex gap-4">
              <div class="flex-shrink-0">
                <span class="flex size-12 items-center justify-center rounded-sm bg-gray-100">
                  <TrendingUp class="size-6 text-gray-700" />
                </span>
              </div>
              <div>
                <h3 class="text-lg font-semibold mb-2">Ongoing Optimization</h3>
                <p class="text-gray-600">
                  Continuous monitoring, protocol adjustments, and long-term wellness maintenance
                  with outcome tracking.
                </p>
              </div>
            </div>
          </div>
        </div>

        <!-- Right Visual Content -->
        <div class="relative">
          <div class="grid grid-cols-1 md:grid-cols-2 gap-4 md:gap-6 items-stretch">
            <!-- Assessment Phase Card -->
            <div
              class="p-4 md:p-6 bg-white rounded-sm border border-gray-200 flex flex-col md:min-h-[200px]"
            >
              <div class="mb-4">
                <h3 class="font-semibold text-gray-900">Assessment Phase</h3>
              </div>
              <div class="space-y-2 text-sm text-gray-600 flex-1">
                <div class="flex items-center">
                  <span class="w-3 h-px bg-black mr-3 mt-0.5 flex-shrink-0"></span>
                  Comprehensive testing
                </div>
                <div class="flex items-center">
                  <span class="w-3 h-px bg-black mr-3 mt-0.5 flex-shrink-0"></span>
                  Biomarker analysis
                </div>
                <div class="flex items-center">
                  <span class="w-3 h-px bg-black mr-3 mt-0.5 flex-shrink-0"></span>
                  Health evaluations
                </div>
                <div class="flex items-center">
                  <span class="w-3 h-px bg-black mr-3 mt-0.5 flex-shrink-0"></span>
                  Personalized planning
                </div>
              </div>
            </div>

            <!-- Treatment Phase Card -->
            <div
              class="p-4 md:p-6 bg-white rounded-sm border border-gray-200 flex flex-col md:min-h-[200px]"
            >
              <div class="mb-4">
                <h3 class="font-semibold text-gray-900">Treatment Phase</h3>
              </div>
              <div class="space-y-2 text-sm text-gray-600 flex-1">
                <div class="flex items-center">
                  <span class="w-3 h-px bg-black mr-3 mt-0.5 flex-shrink-0"></span>
                  Active interventions
                </div>
                <div class="flex items-center">
                  <span class="w-3 h-px bg-black mr-3 mt-0.5 flex-shrink-0"></span>
                  Regenerative therapies
                </div>
                <div class="flex items-center">
                  <span class="w-3 h-px bg-black mr-3 mt-0.5 flex-shrink-0"></span>
                  Progress monitoring
                </div>
                <div class="flex items-center">
                  <span class="w-3 h-px bg-black mr-3 mt-0.5 flex-shrink-0"></span>
                  Protocol optimization
                </div>
              </div>
            </div>

            <!-- Desktop Integration Symbol -->
            <div
              class="absolute top-1/2 left-1/2 transform -translate-x-1/2 -translate-y-1/2 z-10 hidden md:flex"
            >
              <div
                class="w-12 h-12 bg-gray-800 rounded-sm flex items-center justify-center border border-gray-200"
              >
                <Plus class="size-6 text-white" />
              </div>
            </div>

            <!-- Mobile Integration Symbol - Positioned at card junction -->
            <div
              class="absolute right-2 transform -translate-y-1/2 z-10 md:hidden"
              style="top: 37%"
            >
              <div
                class="w-12 h-12 bg-gray-800 rounded-sm flex items-center justify-center border border-gray-200"
              >
                <Plus class="size-6 text-white" />
              </div>
            </div>
          </div>

          <!-- Bottom Result Card -->
          <div
            class="mt-8 p-4 md:p-6 bg-black rounded-sm text-white text-center border border-gray-200"
          >
            <h3 class="font-semibold mb-2">= Clinical Excellence</h3>
            <p class="text-gray-300 text-sm">
              Evidence-based evaluation + individualized protocols = superior therapeutic outcomes
            </p>
          </div>
        </div>
      </div>
    </div>
  </section>
</template>

<script setup lang="ts">
import { ClipboardList, Sparkles, TestTube, TrendingUp, Plus } from 'lucide-vue-next';
</script>
