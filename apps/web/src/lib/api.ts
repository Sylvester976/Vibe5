import type { GenresResponse, HeatmapResponse, Period, StoryResponse, TopArtistsResponse } from './types'

// Same-origin: the Vite dev proxy (vite.config.ts) and the production
// Caddyfile both route /api, /auth, /widget to the Go backend, so the
// session cookie is always first-party.
export class ApiError extends Error {
  status: number
  constructor(status: number, message: string) {
    super(message)
    this.status = status
  }
}

async function request<T>(path: string, init?: RequestInit): Promise<T> {
  const res = await fetch(path, {
    credentials: 'include',
    headers: { 'Content-Type': 'application/json', ...(init?.headers ?? {}) },
    ...init,
  })
  if (!res.ok) {
    const text = await res.text().catch(() => '')
    throw new ApiError(res.status, text || res.statusText)
  }
  if ((res.headers.get('content-type') ?? '').includes('application/json')) {
    return (await res.json()) as T
  }
  return undefined as T
}

const api = {
  get: <T>(path: string) => request<T>(path),
  post: <T>(path: string) => request<T>(path, { method: 'POST' }),
}

export const meApi = {
  get: () => api.get<{ display_name: string }>('/api/me'),
}

export const statsApi = {
  topArtists: (period: Period) => api.get<TopArtistsResponse>(`/api/stats/top-artists?period=${period}`),
  genres: (period: Period) => api.get<GenresResponse>(`/api/stats/genres?period=${period}`),
  heatmap: () => api.get<HeatmapResponse>('/api/stats/heatmap'),
  story: () => api.get<StoryResponse>('/api/story'),
}

export const widgetApi = {
  issueToken: () => api.post<{ widget_url: string }>('/api/widget/token'),
}
