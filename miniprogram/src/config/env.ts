/// <reference types="vite/client" />

/**
 * 小程序运行环境配置
 * 统一收口 HTTP / WS 地址，避免业务代码分散维护。
 * WebSocket 地址优先读取 VITE_WS_BASE_URL，未显式配置时再根据 HTTP 地址自动推导：
 * - https:// -> wss://
 * - http:// -> ws://
 */

const DEFAULT_API_BASE_URL = 'http://192.168.10.6:8080'
const DEFAULT_APP_NAME = '寻梦私域管家'

function trimTrailingSlash(url: string): string {
  return url.replace(/\/+$/, '')
}

function getEnvValue(
  key: 'VITE_API_BASE_URL' | 'VITE_WS_BASE_URL' | 'VITE_APP_NAME',
  fallback: string
): string {
  const value = import.meta.env?.[key]
  return typeof value === 'string' && value.trim() ? value.trim() : fallback
}

function toWsBaseUrl(apiBaseUrl: string): string {
  if (apiBaseUrl.startsWith('https://')) {
    return apiBaseUrl.replace(/^https:\/\//, 'wss://')
  }

  if (apiBaseUrl.startsWith('http://')) {
    return apiBaseUrl.replace(/^http:\/\//, 'ws://')
  }

  return apiBaseUrl
}

export const API_BASE_URL = trimTrailingSlash(
  getEnvValue('VITE_API_BASE_URL', DEFAULT_API_BASE_URL)
)

export const WS_BASE_URL = trimTrailingSlash(
  getEnvValue('VITE_WS_BASE_URL', toWsBaseUrl(API_BASE_URL))
)

export const APP_NAME = getEnvValue('VITE_APP_NAME', DEFAULT_APP_NAME)

export const MERCHANT_SOCKET_URL = `${WS_BASE_URL}/api/v1/ws/merchant`
