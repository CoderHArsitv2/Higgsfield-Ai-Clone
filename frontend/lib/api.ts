"use client";

import type { Generation, Model, ProviderInfo, StoredKey, User } from "./types";

export class ApiError extends Error {
  constructor(
    message: string,
    readonly code: string,
    readonly status: number,
  ) {
    super(message);
  }
}

async function call<T>(path: string, init?: RequestInit): Promise<T> {
  const res = await fetch(`/api/proxy/${path}`, {
    ...init,
    headers: { "Content-Type": "application/json", ...init?.headers },
  });
  if (res.status === 204) return undefined as T;

  const body = await res.json().catch(() => null);
  if (!res.ok) {
    const err = body?.error;
    throw new ApiError(
      err?.message ?? "Something went wrong",
      err?.code ?? "unknown",
      res.status,
    );
  }
  return body as T;
}

export const api = {
  me: () => call<{ user: User; byok_providers: string[] }>("me"),

  models: () =>
    call<{
      models: Model[];
      providers: ProviderInfo[];
      counts: Record<string, number>;
    }>("models"),

  generations: (params?: Record<string, string>) => {
    const q = params ? `?${new URLSearchParams(params)}` : "";
    return call<{ generations: Generation[]; total: number }>(
      `generations${q}`,
    );
  },

  generation: (id: string) =>
    call<{ generation: Generation }>(`generations/${id}`),

  create: (input: {
    model_id: string;
    prompt: string;
    params?: Record<string, unknown>;
  }) =>
    call<{ generation: Generation; credits_remaining: number }>("generations", {
      method: "POST",
      body: JSON.stringify(input),
    }),

  cancel: (id: string) =>
    call<{ generation: Generation }>(`generations/${id}/cancel`, {
      method: "POST",
    }),

  keys: () => call<{ keys: StoredKey[] }>("keys"),

  saveKey: (provider_id: string, api_key: string) =>
    call<{ key: StoredKey }>("keys", {
      method: "PUT",
      body: JSON.stringify({ provider_id, api_key }),
    }),

  deleteKey: (provider_id: string) =>
    call<void>(`keys/${provider_id}`, { method: "DELETE" }),
};
