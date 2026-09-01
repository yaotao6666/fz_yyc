import request, { unwrapApiResponse } from '@/utils/request'

/* ============ 服务评价管理（PRD V2.0 阶段四） ============ */

export interface ServiceReviewStaff {
  id: number
  name: string
  phone: string
}

export interface ServiceReviewOrder {
  id: number
  order_no: string
}

export interface ServiceReview {
  id: number
  order_id: number
  order: ServiceReviewOrder | null
  staff_id: number
  staff: ServiceReviewStaff | null
  score: number
  attitude_score: number
  professional_score: number
  punctual_score: number
  content: string
  images: string[] | null
  status: number // 1=正常展示 0=已隐藏
  status_cn: string
  created_at: string
}

export interface ServiceReviewListResponse {
  list: ServiceReview[]
  total: number
  pagination: { page: number; page_size: number }
}

export function getServiceReviews(params?: Record<string, unknown>) {
  return request.get('/api/v1/merchant/service-reviews', { params }).then(unwrapApiResponse<ServiceReviewListResponse>)
}

export function hideServiceReview(id: number) {
  return request.post(`/api/v1/merchant/service-reviews/${id}/hide`).then(unwrapApiResponse<{ id: number }>)
}