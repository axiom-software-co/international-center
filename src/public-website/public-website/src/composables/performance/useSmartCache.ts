// Smart Cache Composable
// Advanced caching system with TTL, invalidation, and memory management

import { ref, reactive, onUnmounted, computed } from 'vue';

export interface CacheEntry<T = any> {
  data: T;
  timestamp: number;
  ttl: number;
  accessCount: number;
  lastAccessed: number;
  tags?: string[];
}

export interface CacheOptions {
  ttl?: number; // Time to live in milliseconds
  maxSize?: number; // Maximum number of entries
  tags?: string[]; // Tags for grouped invalidation
  priority?: number; // Higher priority items are evicted last
}

export interface UseSmartCacheOptions {
  defaultTtl?: number;
  maxSize?: number;
  cleanupInterval?: number;
  enablePersistence?: boolean;
  storageKey?: string;
}

export interface UseSmartCacheResult {
  // Core cache operations
  get: <T>(key: string) => T | null;
  set: <T>(key: string, data: T, options?: CacheOptions) => void;
  has: (key: string) => boolean;
  delete: (key: string) => boolean;
  clear: () => void;
  
  // Advanced operations
  invalidateByTag: (tag: string) => number;
  invalidateByPattern: (pattern: RegExp) => number;
  prune: () => number;
  
  // Cache statistics
  stats: {
    size: number;
    hitRate: number;
    missRate: number;
    totalRequests: number;
  };
  
  // Persistence
  persist: () => void;
  restore: () => void;
}

export function useSmartCache(options: UseSmartCacheOptions = {}): UseSmartCacheResult {
  const {
    defaultTtl = 5 * 60 * 1000, // 5 minutes
    maxSize = 100,
    cleanupInterval = 60 * 1000, // 1 minute
    enablePersistence = false,
    storageKey = 'smart_cache',
  } = options;

  const cache = reactive<Map<string, CacheEntry>>(new Map());
  const stats = reactive({
    hits: 0,
    misses: 0,
  });

  // Computed statistics
  const cacheStats = computed(() => ({
    size: cache.size,
    hitRate: stats.hits / Math.max(1, stats.hits + stats.misses),
    missRate: stats.misses / Math.max(1, stats.hits + stats.misses),
    totalRequests: stats.hits + stats.misses,
  }));

  // Check if an entry is expired
  const isExpired = (entry: CacheEntry): boolean => {
    return Date.now() - entry.timestamp > entry.ttl;
  };

  // Get value from cache
  const get = <T>(key: string): T | null => {
    const entry = cache.get(key);
    
    if (!entry) {
      stats.misses++;
      return null;
    }
    
    if (isExpired(entry)) {
      cache.delete(key);
      stats.misses++;
      return null;
    }
    
    // Update access statistics
    entry.accessCount++;
    entry.lastAccessed = Date.now();
    stats.hits++;
    
    return entry.data as T;
  };

  // Set value in cache
  const set = <T>(key: string, data: T, cacheOptions: CacheOptions = {}): void => {
    const now = Date.now();
    const entry: CacheEntry<T> = {
      data,
      timestamp: now,
      ttl: cacheOptions.ttl ?? defaultTtl,
      accessCount: 0,
      lastAccessed: now,
      tags: cacheOptions.tags,
    };

    // Check if we need to make room
    if (cache.size >= maxSize) {
      evictLeastValuable();
    }

    cache.set(key, entry);
  };

  // Check if key exists in cache (and is not expired)
  const has = (key: string): boolean => {
    const entry = cache.get(key);
    return entry !== undefined && !isExpired(entry);
  };

  // Delete specific key
  const deleteKey = (key: string): boolean => {
    return cache.delete(key);
  };

  // Clear entire cache
  const clear = (): void => {
    cache.clear();
    stats.hits = 0;
    stats.misses = 0;
  };

  // Evict the least valuable entry based on LRU + access count + TTL
  const evictLeastValuable = (): void => {
    let worstKey: string | null = null;
    let worstScore = Infinity;

    for (const [key, entry] of cache.entries()) {
      // Calculate eviction score (lower is worse)
      // Factors: recency, access count, TTL remaining
      const recency = Date.now() - entry.lastAccessed;
      const popularity = entry.accessCount;
      const timeLeft = entry.ttl - (Date.now() - entry.timestamp);
      
      const score = (popularity * 1000) - recency + Math.max(0, timeLeft);
      
      if (score < worstScore) {
        worstScore = score;
        worstKey = key;
      }
    }

    if (worstKey) {
      cache.delete(worstKey);
    }
  };

  // Invalidate entries by tag
  const invalidateByTag = (tag: string): number => {
    let invalidated = 0;
    
    for (const [key, entry] of cache.entries()) {
      if (entry.tags?.includes(tag)) {
        cache.delete(key);
        invalidated++;
      }
    }
    
    return invalidated;
  };

  // Invalidate entries by pattern
  const invalidateByPattern = (pattern: RegExp): number => {
    let invalidated = 0;
    
    for (const key of cache.keys()) {
      if (pattern.test(key)) {
        cache.delete(key);
        invalidated++;
      }
    }
    
    return invalidated;
  };

  // Remove expired entries
  const prune = (): number => {
    let pruned = 0;
    const now = Date.now();
    
    for (const [key, entry] of cache.entries()) {
      if (now - entry.timestamp > entry.ttl) {
        cache.delete(key);
        pruned++;
      }
    }
    
    return pruned;
  };

  // Persist cache to storage
  const persist = (): void => {
    if (!enablePersistence || typeof localStorage === 'undefined') return;

    try {
      const serializable = Array.from(cache.entries()).map(([key, entry]) => ({
        key,
        data: entry.data,
        timestamp: entry.timestamp,
        ttl: entry.ttl,
        tags: entry.tags,
      }));

      localStorage.setItem(storageKey, JSON.stringify({
        entries: serializable,
        stats: {
          hits: stats.hits,
          misses: stats.misses,
        },
      }));
    } catch (error) {
      console.warn('Failed to persist cache:', error);
    }
  };

  // Restore cache from storage
  const restore = (): void => {
    if (!enablePersistence || typeof localStorage === 'undefined') return;

    try {
      const stored = localStorage.getItem(storageKey);
      if (!stored) return;

      const parsed = JSON.parse(stored);
      
      // Restore entries
      if (parsed.entries) {
        for (const item of parsed.entries) {
          const entry: CacheEntry = {
            ...item,
            accessCount: 0,
            lastAccessed: Date.now(),
          };
          
          // Only restore non-expired entries
          if (!isExpired(entry)) {
            cache.set(item.key, entry);
          }
        }
      }

      // Restore stats
      if (parsed.stats) {
        stats.hits = parsed.stats.hits || 0;
        stats.misses = parsed.stats.misses || 0;
      }
    } catch (error) {
      console.warn('Failed to restore cache:', error);
    }
  };

  // Automatic cleanup interval
  let cleanupTimer: number | null = null;
  
  if (typeof window !== 'undefined') {
    cleanupTimer = window.setInterval(() => {
      prune();
      if (enablePersistence) {
        persist();
      }
    }, cleanupInterval);
  }

  // Restore cache on mount
  if (enablePersistence) {
    restore();
  }

  // Cleanup on unmount
  onUnmounted(() => {
    if (cleanupTimer) {
      clearInterval(cleanupTimer);
    }
    if (enablePersistence) {
      persist();
    }
  });

  // Persist before page unload
  if (enablePersistence && typeof window !== 'undefined') {
    window.addEventListener('beforeunload', persist);
    
    onUnmounted(() => {
      window.removeEventListener('beforeunload', persist);
    });
  }

  return {
    get,
    set,
    has,
    delete: deleteKey,
    clear,
    invalidateByTag,
    invalidateByPattern,
    prune,
    stats: cacheStats,
    persist,
    restore,
  };
}

// Helper function to create a cached version of an async function
export function withSmartCache<T extends (...args: any[]) => Promise<any>>(
  fn: T,
  keyGenerator: (...args: Parameters<T>) => string,
  cache: UseSmartCacheResult,
  options?: CacheOptions
): T {
  return (async (...args: Parameters<T>): Promise<ReturnType<T>> => {
    const key = keyGenerator(...args);
    
    // Try to get from cache first
    const cached = cache.get<ReturnType<T>>(key);
    if (cached !== null) {
      return cached;
    }
    
    // Execute function and cache result
    try {
      const result = await fn(...args);
      cache.set(key, result, options);
      return result;
    } catch (error) {
      // Don't cache errors, but still throw them
      throw error;
    }
  }) as T;
}