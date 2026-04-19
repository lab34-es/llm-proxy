const API_BASE = '/admin'

function getToken(): string {
  return localStorage.getItem('llmp_admin_token') || ''
}

async function request<T>(path: string, options: RequestInit = {}): Promise<T> {
  const token = getToken()
  const headers: Record<string, string> = {
    ...(options.headers as Record<string, string> || {}),
  }
  if (token) {
    headers['Authorization'] = `Bearer ${token}`
  }
  if (options.body && typeof options.body === 'string') {
    headers['Content-Type'] = 'application/json'
  }

  const resp = await fetch(`${API_BASE}${path}`, { ...options, headers })

  if (resp.status === 401) {
    localStorage.removeItem('llmp_admin_token')
    window.location.href = '/dashboard/login'
    throw new Error('Unauthorized')
  }

  if (resp.status === 204) {
    return undefined as T
  }

  if (!resp.ok) {
    const body = await resp.json().catch(() => ({ message: resp.statusText }))
    throw new Error(body.message || resp.statusText)
  }

  return resp.json()
}

// ── Providers ──────────────────────────────────────────────────────────

export interface Provider {
  id: string
  name: string
  base_url: string
  created_at: string
  updated_at: string
}

export function listProviders(): Promise<Provider[]> {
  return request('/providers')
}

export function createProvider(data: { name: string; base_url: string; api_key: string }): Promise<Provider> {
  return request('/providers', { method: 'POST', body: JSON.stringify(data) })
}

export function deleteProvider(id: string): Promise<void> {
  return request(`/providers/${id}`, { method: 'DELETE' })
}

// ── API Keys ───────────────────────────────────────────────────────────

export interface APIKey {
  id: string
  name: string
  provider_id: string
  rate_limit_rpm: number
  created_at: string
  revoked_at?: string | null
}

export interface CreateKeyResponse {
  id: string
  name: string
  key: string
  provider_id: string
  rate_limit_rpm: number
  created_at: string
}

export function listKeys(): Promise<APIKey[]> {
  return request('/keys')
}

export function createKey(data: { name: string; provider_id: string; rate_limit_rpm: number }): Promise<CreateKeyResponse> {
  return request('/keys', { method: 'POST', body: JSON.stringify(data) })
}

export function revokeKey(id: string): Promise<void> {
  return request(`/keys/${id}`, { method: 'DELETE' })
}

// ── Usage ──────────────────────────────────────────────────────────────

export interface UsageRecord {
  id: string
  api_key_id: string
  provider_id: string
  model: string
  prompt_tokens: number
  completion_tokens: number
  total_tokens: number
  created_at: string
}

export interface UsageResult {
  records: UsageRecord[] | null
  total: number
}

export function queryUsage(params: Record<string, string>): Promise<UsageResult> {
  const qs = new URLSearchParams(params).toString()
  return request(`/usage?${qs}`)
}

// ── Guardrails ─────────────────────────────────────────────────────────

export interface Guardrail {
  id: string
  pattern: string
  mode: string
  replace_by?: string
  created_at: string
  updated_at: string
}

export function listGuardrails(): Promise<Guardrail[]> {
  return request('/guardrails')
}

export function createGuardrail(data: { pattern: string; mode: string; replace_by?: string }): Promise<Guardrail> {
  return request('/guardrails', { method: 'POST', body: JSON.stringify(data) })
}

export function deleteGuardrail(id: string): Promise<void> {
  return request(`/guardrails/${id}`, { method: 'DELETE' })
}

// ── Guardrail Events ───────────────────────────────────────────────────

export interface GuardrailEvent {
  id: string
  guardrail_id: string
  api_key_id: string
  pattern: string
  mode: string
  input_text: string
  created_at: string
}

export interface GuardrailEventResult {
  records: GuardrailEvent[] | null
  total: number
}

export function listGuardrailEvents(params?: Record<string, string>): Promise<GuardrailEventResult> {
  const qs = params ? new URLSearchParams(params).toString() : ''
  return request(`/guardrail-events${qs ? '?' + qs : ''}`)
}

export function deleteGuardrailEvent(id: string): Promise<void> {
  return request(`/guardrail-events/${id}`, { method: 'DELETE' })
}
