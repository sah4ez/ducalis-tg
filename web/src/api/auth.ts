import { api } from './client';
import type { User } from './types';

export interface AuthResponse {
  user: User;
  token: string;
}

export async function login(email: string, password: string): Promise<AuthResponse> {
  return api.post('/api/v1/auth/login', { email, password });
}

export async function register(name: string, email: string, password: string): Promise<AuthResponse> {
  return api.post('/api/v1/auth/register', { req: { name, email, password } });
}

export async function getMe(userID: string): Promise<User> {
  return api.get(`/api/v1/auth/me?userID=${encodeURIComponent(userID)}`);
}
