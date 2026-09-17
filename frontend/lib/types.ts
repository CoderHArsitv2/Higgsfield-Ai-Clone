export type Modality = "image" | "video" | "audio";
export type Access = "server" | "byok" | "locked";

export interface Option {
  value: string;
  label: string;
}

export interface ParamSpec {
  key: string;
  label: string;
  kind: "select" | "number" | "text" | "bool";
  options?: Option[];
  default?: string | number | boolean;
  min?: number;
  max?: number;
  step?: number;
  optional?: boolean;
  help?: string;
}

export interface Model {
  id: string;
  provider_id: string;
  provider_name: string;
  name: string;
  modality: Modality;
  description: string;
  tags?: string[];
  credit_cost: number;
  params?: ParamSpec[];
  ref_images: number;
  featured?: boolean;
  access: Access;
  enabled: boolean;
  lock_reason?: string;
  key_env_name?: string;
  docs_url?: string;
}

export interface ProviderInfo {
  id: string;
  name: string;
  env_key: string;
  docs_url: string;
  has_server_key: boolean;
  accepts_byok: boolean;
  model_count: number;
  enabled_models: number;
}

export interface Asset {
  id: string;
  kind: string;
  url: string;
  thumbnail_url?: string;
  width?: number;
  height?: number;
  duration_ms?: number;
}

export type GenerationStatus =
  "queued" | "running" | "succeeded" | "failed" | "canceled";

export interface Generation {
  id: string;
  model_id: string;
  provider_id: string;
  modality: Modality;
  prompt: string;
  params: Record<string, unknown>;
  status: GenerationStatus;
  error?: string;
  used_own_key: boolean;
  created_at: string;
  completed_at?: string;
  assets: Asset[] | null;
}

export interface User {
  id: string;
  email: string;
  name: string;
  picture: string;
  credits: number;
}

export interface StoredKey {
  id: string;
  provider_id: string;
  preview: string;
  last_used_at?: string;
}

/** Shape returned by the public, unauthenticated catalogue endpoint. */
export interface PublicModel {
  id: string;
  name: string;
  modality: Modality;
  provider_name: string;
  description: string;
  enabled: boolean;
  featured: boolean;
  tags?: string[];
}
