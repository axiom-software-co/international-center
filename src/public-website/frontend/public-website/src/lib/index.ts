// Main Library Exports
// Clean exports for the new domain-driven architecture

// Re-export all clients and types
export * from './clients';

// Re-export all services
export * from './services';

// Re-export all React hooks
export * from '../hooks';

// Re-export commonly used utilities (keeping existing ones that don't conflict)
export { cn } from './utils';
export { validateEmail, validatePhone } from './validation';

// Theme utilities
export { getTheme, setTheme, toggleTheme, initializeTheme } from './theme';
