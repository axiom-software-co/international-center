<template>
  <div :class="cn('p-3', props.class)" v-bind="$attrs">
    <!-- Header -->
    <div class="flex items-center justify-between mb-4">
      <Button variant="outline" size="sm" class="h-7 w-7 p-0" @click="navigateMonth(-1)">
        <ChevronLeft class="h-4 w-4" />
      </Button>

      <div class="text-sm font-medium">{{ MONTHS[viewMonth] }} {{ viewYear }}</div>

      <Button variant="outline" size="sm" class="h-7 w-7 p-0" @click="navigateMonth(1)">
        <ChevronRight class="h-4 w-4" />
      </Button>
    </div>

    <!-- Days of week header -->
    <div class="grid grid-cols-7 mb-2">
      <div
        v-for="day in DAYS"
        :key="day"
        class="text-center text-sm font-medium text-muted-foreground p-2"
      >
        {{ day }}
      </div>
    </div>

    <!-- Calendar grid -->
    <div class="grid grid-cols-7 gap-1">
      <!-- Empty cells for days before month starts -->
      <div v-for="i in firstDay" :key="`empty-${i}`" class="p-2" />

      <!-- Days of the month -->
      <Button
        v-for="day in daysInMonth"
        :key="day"
        variant="ghost"
        size="sm"
        :class="
          cn(
            'h-8 w-8 p-0 font-normal',
            isToday(new Date(viewYear, viewMonth, day)) && 'bg-accent text-accent-foreground',
            isSelected(new Date(viewYear, viewMonth, day)) &&
              'bg-primary text-primary-foreground hover:bg-primary hover:text-primary-foreground',
            isDateDisabled(new Date(viewYear, viewMonth, day)) &&
              'text-muted-foreground opacity-50 cursor-not-allowed'
          )
        "
        :disabled="isDateDisabled(new Date(viewYear, viewMonth, day))"
        @click="handleDateClick(day)"
      >
        {{ day }}
      </Button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue';
import Button from './Button.vue';
import { ChevronLeft, ChevronRight } from 'lucide-vue-next';
import { cn } from '@/lib/utils';

export interface CalendarProps {
  class?: string;
  selected?: Date;
  disabled?: (date: Date) => boolean;
  minDate?: Date;
  maxDate?: Date;
}

const props = defineProps<CalendarProps>();

const emit = defineEmits<{
  select: [date: Date | undefined];
}>();

const DAYS = ['Su', 'Mo', 'Tu', 'We', 'Th', 'Fr', 'Sa'];
const MONTHS = [
  'January',
  'February',
  'March',
  'April',
  'May',
  'June',
  'July',
  'August',
  'September',
  'October',
  'November',
  'December',
];

const currentDate = ref(props.selected || new Date());
const viewMonth = ref(currentDate.value.getMonth());
const viewYear = ref(currentDate.value.getFullYear());

const getDaysInMonth = (month: number, year: number) => {
  return new Date(year, month + 1, 0).getDate();
};

const getFirstDayOfMonth = (month: number, year: number) => {
  return new Date(year, month, 1).getDay();
};

const daysInMonth = computed(() => getDaysInMonth(viewMonth.value, viewYear.value));
const firstDay = computed(() => getFirstDayOfMonth(viewMonth.value, viewYear.value));

const isDateDisabled = (date: Date) => {
  if (props.disabled && props.disabled(date)) return true;
  if (props.minDate && date < props.minDate) return true;
  if (props.maxDate && date > props.maxDate) return true;
  return false;
};

const isToday = (date: Date) => {
  const today = new Date();
  return date.toDateString() === today.toDateString();
};

const isSelected = (date: Date) => {
  return props.selected && date.toDateString() === props.selected.toDateString();
};

const handleDateClick = (day: number) => {
  const date = new Date(viewYear.value, viewMonth.value, day);
  if (!isDateDisabled(date)) {
    emit('select', date);
  }
};

const navigateMonth = (direction: number) => {
  const newMonth = viewMonth.value + direction;
  if (newMonth < 0) {
    viewMonth.value = 11;
    viewYear.value = viewYear.value - 1;
  } else if (newMonth > 11) {
    viewMonth.value = 0;
    viewYear.value = viewYear.value + 1;
  } else {
    viewMonth.value = newMonth;
  }
};

watch(
  () => props.selected,
  newSelected => {
    if (newSelected) {
      currentDate.value = newSelected;
      viewMonth.value = newSelected.getMonth();
      viewYear.value = newSelected.getFullYear();
    }
  }
);
</script>
