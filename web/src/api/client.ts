// REST client для ducalis-tg (tg v3 transport).
//
// Форматы (сверено с E2E):
//   - Login:      POST /api/v1/auth/login       body {email, password}    → {user, token}
//   - Register:   POST /api/v1/auth/register    body {req: {name,email,password}} → {user, token}
//   - Struct params: body обёрнут в {req: {...}} когда параметр контракта — структура.
//   - enableInlineSingle методы (Get, Create, Vote...): ответ — сам объект без конверта.
//   - List: конверт {tasks: [...], total: N} / {workspaces: [...], total: N}.
//   - GetRanked: {result: {tasks: [...]}}.

const BASE = '';

export class ApiError extends Error {
  status: number;
  constructor(message: string, status: number) {
    super(message);
    this.status = status;
    this.name = 'ApiError';
  }
}

let authToken: string | null = localStorage.getItem('token');
export function setToken(t: string | null) {
  authToken = t;
  if (t) localStorage.setItem('token', t);
  else localStorage.removeItem('token');
}
export function getToken(): string | null {
  return authToken;
}

async function req<T>(method: string, path: string, body?: unknown): Promise<T> {
  const headers: Record<string, string> = { 'Content-Type': 'application/json' };
  if (authToken) headers['Authorization'] = `Bearer ${authToken}`;

  const res = await fetch(BASE + path, {
    method,
    headers,
    body: body !== undefined ? JSON.stringify(body) : undefined,
  });

  if (res.status === 204) return undefined as T;

  const text = await res.text();
  let json: any;
  try {
    json = text ? JSON.parse(text) : {};
  } catch {
    throw new ApiError(text || res.statusText, res.status);
  }

  if (!res.ok) {
    // Postgres/pgx ошибки приходят как {Message: "..."}
    const msg = json.Message || json.error?.message || json.message || res.statusText;
    throw new ApiError(String(msg), res.status);
  }
  return json as T;
}

export const api = {
  get: <T>(path: string) => req<T>('GET', path),
  post: <T>(path: string, body?: unknown) => req<T>('POST', path, body),
  put: <T>(path: string, body?: unknown) => req<T>('PUT', path, body),
  patch: <T>(path: string, body?: unknown) => req<T>('PATCH', path, body),
  del: <T>(path: string, body?: unknown) => req<T>('DELETE', path, body),
};
