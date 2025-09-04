// Contacts Domain Types - Standardized API responses

import {
  BaseEntity,
  PaginationParams,
  FilterParams,
  StandardResponse,
  SingleResponse,
} from '../shared/types';

export interface Contact extends BaseEntity {
  name: string;
  email: string;
  phone?: string;
  subject?: string;
  message: string;
  status: 'submitted' | 'processing' | 'completed' | 'archived';
  ip_address?: string;
  user_agent?: string;
}

export interface ContactSubmission {
  name: string;
  email: string;
  phone?: string;
  subject?: string;
  message: string;
}

export interface ContactSubmissionResponse {
  id: string;
  message: string;
  timestamp: string;
  status: string;
}

// Standardized response types
export type ContactsResponse = StandardResponse<Contact>;
export type ContactResponse = SingleResponse<Contact>;
export type ContactSubmissionResult = SingleResponse<ContactSubmissionResponse>;

export interface GetContactsParams extends PaginationParams, FilterParams {
  status?: string;
}

export interface SearchContactsParams extends PaginationParams {
  q: string;
  sortBy?: string;
}

// Legacy support - will be deprecated
export interface LegacyContactsResponse {
  contacts: Contact[];
  total: number;
  page: number;
  pageSize: number;
  totalPages: number;
}
