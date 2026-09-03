import request, { unwrapApiResponse } from '@/utils/request'

/* ============ 服务过程安全（PRD V2.0 阶段三） ============ */

/* ---- 预警中心 ---- */

export interface AlertEventStaff {
  id: number
  name: string
  phone: string
}

export interface AlertEventOrder {
  id: number
  order_no: string
  address: string
  biz_status: number
  status: number
}

export interface AlertEvent {
  id: number
  order_id: number | null
  staff_id: number
  staff: AlertEventStaff | null
  order: AlertEventOrder | null
  alert_type: number // 1=SOS求助 2=服务超时未结束
  alert_type_cn: string
  summary?: string
  lat: number
  lng: number
  address: string
  status: number // 1=待处理 2=处理中 3=已处理
  status_cn: string
  handler_id: number | null
  handler_name: string
  handle_remark: string
  handled_at: string | null
  created_at: string
}

export interface AlertEventListResponse {
  list: AlertEvent[]
  total: number
  pagination: { page: number; page_size: number }
}

export function getAlertEvents(params?: Record<string, unknown>) {
  return request.get('/api/v1/merchant/alert-events', { params }).then(unwrapApiResponse<AlertEventListResponse>)
}

export function getAlertEventDetail(id: number) {
  return request.get(`/api/v1/merchant/alert-events/${id}`).then(unwrapApiResponse<AlertEvent>)
}

export function handleAlertEvent(id: number, status: number, remark: string) {
  return request.post(`/api/v1/merchant/alert-events/${id}/handle`, { status, remark }).then(unwrapApiResponse<AlertEvent>)
}

export interface AlertSettings {
  enabled: boolean
  goods_unverified_hours: number
  service_unassigned_hours: number
  escort_unfinished_minutes: number
  service_unstarted_minutes: number
  rental_overdue_hours: number
  refund_stuck_hours: number
  service_audio_retain_days: number
}

export function getAlertSettings() {
  return request.get('/api/v1/merchant/alert-settings').then(unwrapApiResponse<AlertSettings>)
}

export function updateAlertSettings(data: AlertSettings) {
  return request.put('/api/v1/merchant/alert-settings', data).then(unwrapApiResponse<{ message: string }>)
}

/* ---- 通用系统配置（system_configs key-value + JSON + 备注） ---- */

export interface SystemConfigEntry {
  id?: number
  config_key: string
  config_value: string
  remark?: string | null
  created_at?: string
  updated_at?: string
}

export function getSystemConfigs() {
  return request.get('/api/v1/merchant/system-configs').then(unwrapApiResponse<SystemConfigEntry[]>)
}

export function updateSystemConfigs(items: { config_key: string; config_value: string; remark?: string }[]) {
  return request.put('/api/v1/merchant/system-configs', { items }).then(unwrapApiResponse<{ message: string }>)
}

export function deleteSystemConfig(configKey: string) {
  return request
    .delete(`/api/v1/merchant/system-configs/${encodeURIComponent(configKey)}`)
    .then(unwrapApiResponse<{ message: string }>)
}

/* ---- 协议管理 ---- */

export interface Agreement {
  id: number
  type: number // 1=用户协议 2=隐私政策 3=录音/定位授权协议
  type_cn: string
  title: string
  content: string
  version: string
  status: number // 1=已发布(当前生效) 0=草稿/停用
  published_at: string | null
  created_at: string
  updated_at: string
}

export interface AgreementListResponse {
  list: Agreement[]
  total: number
  pagination: { page: number; page_size: number }
}

export interface AgreementPayload {
  type: number
  title: string
  content: string
  version: string
}

export function getAgreements(params?: Record<string, unknown>) {
  return request.get('/api/v1/merchant/agreements', { params }).then(unwrapApiResponse<AgreementListResponse>)
}

export function getAgreementDetail(id: number) {
  return request.get(`/api/v1/merchant/agreements/${id}`).then(unwrapApiResponse<Agreement>)
}

export function createAgreement(data: AgreementPayload) {
  return request.post('/api/v1/merchant/agreements', data).then(unwrapApiResponse<{ id: number }>)
}

export function updateAgreement(id: number, data: { title: string; content: string }) {
  return request.put(`/api/v1/merchant/agreements/${id}`, data).then(unwrapApiResponse<{ id: number }>)
}

export function publishAgreement(id: number) {
  return request.post(`/api/v1/merchant/agreements/${id}/publish`).then(unwrapApiResponse<{ id: number }>)
}

/* ---- 订单服务记录 ---- */

export interface ServiceRecordStaff {
  id: number
  name: string
  phone: string
  service_region: string
  quality_score: number
}

export interface ServiceRecord {
  id: number
  staff_id: number
  staff: ServiceRecordStaff | Record<string, never>
  start_time: string | null
  end_time: string | null
  duration_minutes: number
  gps_track_url: string
  audio_url: string
  audio_uploaded_at: string | null
  audio_deleted_at: string | null
  sos_triggered: number
  status: number // 1=正常 2=异常
  created_at: string
}

export interface LocationTrack {
  id: number
  order_id: number
  staff_id: number
  lat: number
  lng: number
  reported_at: string
}

export interface ServiceRecordResponse {
  order_id: number
  service_record: ServiceRecord | null
  tracks: LocationTrack[]
  track_count: number
}

export function getOrderServiceRecord(orderId: number) {
  return request.get(`/api/v1/merchant/orders/${orderId}/service-record`).then(unwrapApiResponse<ServiceRecordResponse>)
}

/* ---- 服务人员区域维护 ---- */

export function updateServiceStaffRegion(id: number, serviceRegion: string) {
  return request.put(`/api/v1/merchant/service-staff/${id}/service-region`, { service_region: serviceRegion }).then(unwrapApiResponse<{ id: number }>)
}
