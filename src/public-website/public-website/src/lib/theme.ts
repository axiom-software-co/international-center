import {
  HiOutlineHeart,
  HiOutlineUserGroup,
  HiOutlineShieldCheck,
  HiOutlineDocumentText,
} from 'react-icons/hi2';
import { FaMicroscope, FaPills } from 'react-icons/fa6';
import { MdOutlineWaterDrop } from 'react-icons/md';
import { IoFlashOutline } from 'react-icons/io5';
import {
  Clock,
  Users,
  Shield,
  Calendar,
  CheckCircle2,
  Heart,
  MapPin,
  Home,
  Building,
  Download,
} from 'lucide-react';

export const colors = {
  primary: {
    light: 'emerald-50',
    medium: 'emerald-500',
    dark: 'emerald-600',
    border: 'emerald-200',
    text: 'emerald-700',
    bg: 'emerald-100',
  },
  blue: {
    light: 'blue-50',
    medium: 'blue-500',
    dark: 'blue-600',
    border: 'blue-200',
    text: 'blue-700',
    bg: 'blue-100',
  },
  green: {
    light: 'green-50',
    medium: 'green-500',
    dark: 'green-600',
    border: 'green-200',
    text: 'green-700',
    bg: 'green-100',
  },
  purple: {
    light: 'purple-50',
    medium: 'purple-500',
    dark: 'purple-600',
    border: 'purple-200',
    text: 'purple-700',
    bg: 'purple-100',
  },
  orange: {
    light: 'orange-50',
    medium: 'orange-500',
    dark: 'orange-600',
    border: 'orange-200',
    text: 'orange-700',
    bg: 'orange-100',
  },
  gray: {
    light: 'gray-50',
    medium: 'gray-500',
    dark: 'gray-600',
    border: 'gray-200',
    text: 'gray-700',
    bg: 'gray-100',
  },
} as const;

export const icons = {
  // Service icons
  heart: HiOutlineHeart,
  microscope: FaMicroscope,
  pills: FaPills,
  waterDrop: MdOutlineWaterDrop,
  flash: IoFlashOutline,
  userGroup: HiOutlineUserGroup,
  shieldCheck: HiOutlineShieldCheck,

  // General icons
  clock: Clock,
  users: Users,
  shield: Shield,
  calendar: Calendar,
  checkCircle: CheckCircle2,
  heartLucide: Heart,
  mapPin: MapPin,
  home: Home,
  building: Building,
  document: HiOutlineDocumentText,
  download: Download,
} as const;

export type ColorTheme = keyof typeof colors;
export type IconName = keyof typeof icons;

export interface ThemeConfig {
  primary: ColorTheme;
  secondary: ColorTheme;
  accent: ColorTheme;
}

export const defaultTheme: ThemeConfig = {
  primary: 'primary',
  secondary: 'blue',
  accent: 'green',
};

export function getColorClasses(
  theme: ColorTheme,
  variant: 'light' | 'medium' | 'dark' | 'border' | 'text' | 'bg' = 'medium'
) {
  return colors[theme][variant];
}

export function getIcon(iconName: IconName) {
  return icons[iconName];
}
