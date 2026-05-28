import axios from 'axios';

const API_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080/api/v1';
const API_KEY = process.env.NEXT_PUBLIC_API_KEY || 'test-api-key-123';

export const apiClient = axios.create({
  baseURL: API_URL,
  headers: {
    'Content-Type': 'application/json',
    'X-API-Key': API_KEY,
  },
});

export type NotificationChannel = 'email' | 'sms' | 'push' | 'webhook' | 'app' | 'imessage';

// Types
export interface Notification {
  id: string;
  user_id?: string;
  template_id?: string;
  recipient_email: string;
  subject: string;
  body: string;
  status: 'pending' | 'processing' | 'sent' | 'failed' | 'cancelled';
  temporal_workflow_id?: string;
  temporal_run_id?: string;
  scheduled_at?: string;
  sent_at?: string;
  failed_at?: string;
  error_message?: string;
  metadata?: Record<string, unknown>;
  created_at: string;
  updated_at: string;
}

export interface Template {
  id: string;
  name: string;
  description: string;
  subject: string;
  body: string;
  variables: string[];
  is_active: boolean;
  created_at: string;
  updated_at: string;
}

export interface CreateNotificationRequest {
  channel?: NotificationChannel;
  recipient?: string;
  template_id?: string;
  recipient_email: string;
  subject?: string;
  body?: string;
  variables?: Record<string, string>;
  scheduled_at?: string;
  metadata?: Record<string, unknown>;
}

export interface CreateTemplateRequest {
  name: string;
  description: string;
  subject: string;
  body: string;
  variables: string[];
}

// API functions
export const notificationApi = {
  list: async (): Promise<{ data: Notification[] }> => {
    const response = await apiClient.get('/notifications');
    return response.data;
  },

  get: async (id: string): Promise<Notification> => {
    const response = await apiClient.get(`/notifications/${id}`);
    return response.data;
  },

  create: async (data: CreateNotificationRequest): Promise<Notification> => {
    const response = await apiClient.post('/notifications', data);
    return response.data;
  },

  getStatus: async (id: string) => {
    const response = await apiClient.get(`/notifications/${id}/status`);
    return response.data;
  },
};

export const templateApi = {
  list: async (): Promise<{ data: Template[] }> => {
    const response = await apiClient.get('/templates');
    return response.data;
  },

  get: async (id: string): Promise<Template> => {
    const response = await apiClient.get(`/templates/${id}`);
    return response.data;
  },

  create: async (data: CreateTemplateRequest): Promise<Template> => {
    const response = await apiClient.post('/templates', data);
    return response.data;
  },

  update: async (id: string, data: Partial<CreateTemplateRequest>): Promise<Template> => {
    const response = await apiClient.put(`/templates/${id}`, data);
    return response.data;
  },

  delete: async (id: string): Promise<void> => {
    await apiClient.delete(`/templates/${id}`);
  },
};
