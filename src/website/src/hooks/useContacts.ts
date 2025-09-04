// Contacts Hook - React hook for contacts data
// Provides clean interface for components to interact with contacts domain

import { useState, useCallback } from 'react';
import { contactsClient } from '../lib/clients';
import type {
  Contact,
  ContactSubmission,
  ContactSubmissionResult,
  GetContactsParams,
  SearchContactsParams,
} from '../lib/clients';

export interface UseContactSubmissionResult {
  submitting: boolean;
  error: string | null;
  success: boolean;
  submitContact: (contactData: ContactSubmission) => Promise<ContactSubmissionResult | null>;
  reset: () => void;
}

export function useContactSubmission(): UseContactSubmissionResult {
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [success, setSuccess] = useState(false);

  const submitContact = useCallback(
    async (contactData: ContactSubmission): Promise<ContactSubmissionResult | null> => {
      try {
        setSubmitting(true);
        setError(null);
        setSuccess(false);

        const response = await contactsClient.submitContact(contactData);
        setSuccess(true);
        return response;
      } catch (err) {
        const errorMessage = err instanceof Error ? err.message : 'Failed to submit contact form';
        setError(errorMessage);
        console.error('Error submitting contact form:', err);
        return null;
      } finally {
        setSubmitting(false);
      }
    },
    []
  );

  const reset = useCallback(() => {
    setError(null);
    setSuccess(false);
    setSubmitting(false);
  }, []);

  return {
    submitting,
    error,
    success,
    submitContact,
    reset,
  };
}

// Admin-only hooks for managing contacts

export interface UseContactsResult {
  contacts: Contact[];
  loading: boolean;
  error: string | null;
  total: number;
  page: number;
  pageSize: number;
  totalPages: number;
  refetch: () => Promise<void>;
}

export interface UseContactsOptions extends GetContactsParams {
  enabled?: boolean;
}

export function useContacts(options: UseContactsOptions = {}): UseContactsResult {
  const { enabled = true, ...params } = options;

  const [contacts, setContacts] = useState<Contact[]>([]);
  const [loading, setLoading] = useState(enabled);
  const [error, setError] = useState<string | null>(null);
  const [pagination, setPagination] = useState({
    total: 0,
    page: 1,
    pageSize: 10,
    totalPages: 0,
  });

  const fetchContacts = useCallback(async () => {
    if (!enabled) return;

    try {
      setLoading(true);
      setError(null);

      const response = await contactsClient.getContacts(params);

      setContacts(response.contacts || []);
      setPagination({
        total: response.pagination?.total || 0,
        page: response.pagination?.page || 1,
        pageSize: response.pagination?.pageSize || 10,
        totalPages: response.pagination?.totalPages || 0,
      });
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : 'Failed to fetch contacts';
      setError(errorMessage);
      console.error('Error fetching contacts:', err);
    } finally {
      setLoading(false);
    }
  }, [enabled, JSON.stringify(params)]);

  return {
    contacts,
    loading,
    error,
    ...pagination,
    refetch: fetchContacts,
  };
}

export interface UseContactResult {
  contact: Contact | null;
  loading: boolean;
  error: string | null;
  refetch: () => Promise<void>;
}

export function useContact(id: string | null): UseContactResult {
  const [contact, setContact] = useState<Contact | null>(null);
  const [loading, setLoading] = useState(!!id);
  const [error, setError] = useState<string | null>(null);

  const fetchContact = useCallback(async () => {
    if (!id) {
      setContact(null);
      setLoading(false);
      return;
    }

    try {
      setLoading(true);
      setError(null);

      const response = await contactsClient.getContactById(id);
      setContact(response);
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : 'Failed to fetch contact';
      setError(errorMessage);
      console.error('Error fetching contact:', err);
    } finally {
      setLoading(false);
    }
  }, [id]);

  return {
    contact,
    loading,
    error,
    refetch: fetchContact,
  };
}

export interface UseContactSearchResult {
  contacts: Contact[];
  loading: boolean;
  error: string | null;
  total: number;
  page: number;
  pageSize: number;
  totalPages: number;
  search: (query: string, params?: Omit<SearchContactsParams, 'q'>) => Promise<void>;
  clearResults: () => void;
}

export function useContactSearch(): UseContactSearchResult {
  const [contacts, setContacts] = useState<Contact[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [pagination, setPagination] = useState({
    total: 0,
    page: 1,
    pageSize: 10,
    totalPages: 0,
  });

  const search = useCallback(
    async (query: string, params: Omit<SearchContactsParams, 'q'> = {}) => {
      if (!query.trim()) {
        clearResults();
        return;
      }

      try {
        setLoading(true);
        setError(null);

        const response = await contactsClient.searchContacts({ q: query, ...params });

        setContacts(response.contacts || []);
        setPagination({
          total: response.pagination?.total || 0,
          page: response.pagination?.page || 1,
          pageSize: response.pagination?.pageSize || 10,
          totalPages: response.pagination?.totalPages || 0,
        });
      } catch (err) {
        const errorMessage = err instanceof Error ? err.message : 'Failed to search contacts';
        setError(errorMessage);
        console.error('Error searching contacts:', err);
      } finally {
        setLoading(false);
      }
    },
    []
  );

  const clearResults = useCallback(() => {
    setContacts([]);
    setPagination({ total: 0, page: 1, pageSize: 10, totalPages: 0 });
    setError(null);
  }, []);

  return {
    contacts,
    loading,
    error,
    ...pagination,
    search,
    clearResults,
  };
}

export interface UseRecentContactsResult {
  contacts: Contact[];
  loading: boolean;
  error: string | null;
  refetch: () => Promise<void>;
}

export function useRecentContacts(limit: number = 10): UseRecentContactsResult {
  const [contacts, setContacts] = useState<Contact[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const fetchRecentContacts = useCallback(async () => {
    try {
      setLoading(true);
      setError(null);

      const response = await contactsClient.getRecentContacts(limit);
      setContacts(response);
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : 'Failed to fetch recent contacts';
      setError(errorMessage);
      console.error('Error fetching recent contacts:', err);
    } finally {
      setLoading(false);
    }
  }, [limit]);

  return {
    contacts,
    loading,
    error,
    refetch: fetchRecentContacts,
  };
}
