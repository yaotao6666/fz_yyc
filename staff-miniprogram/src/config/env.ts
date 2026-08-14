export const ENV = {
  API_BASE_URL: import.meta.env.VITE_API_BASE_URL as string || 'http://localhost:8080'
}

export function getApiBaseUrl(): string {
  return ENV.API_BASE_URL.replace(/\/+$/, '')
}
