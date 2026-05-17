/**
 * 用户行为埋点工具
 * 用于记录用户访问、点击等行为事件
 */

import { trackStoreBehaviorEvent } from '../api/store'
import { post } from './request'
import { useAuth } from './useAuth'

interface VisitParams {
  merchant_id: number
  source?: string
}

interface TrackEventParams {
  event: 'page_view' | 'product_view' | 'submit_order' | 'pay_success'
  merchant_id: number
  page?: string
  product_id?: number
  order_id?: number
  source?: string
  data?: Record<string, any>
}

export function useAnalytics() {
  const { ensureAuth, getOpenid } = useAuth()

  let authPromise: Promise<boolean> | null = null

  async function ensureAuthed(): Promise<boolean> {
    if (!authPromise) {
      authPromise = ensureAuth().finally(() => {
        authPromise = null
      })
    }
    return await authPromise
  }

  const trackVisit = async (params: VisitParams): Promise<boolean> => {
    try {
      const authed = await ensureAuthed()
      if (!authed) {
        return false
      }

      const res = await post<{ user_id: number; visit_count: number }>(
        `/api/v1/store/${params.merchant_id}/visit`,
        {
          openid: getOpenid(),
          source: params.source || 'scan'
        },
        { loading: false, showErrorToast: false }
      )

      return !!res
    } catch (error) {
      console.error('Analytics: 访问埋点异常', error)
      return false
    }
  }

  const trackEvent = async (params: TrackEventParams): Promise<boolean> => {
    try {
      const authed = await ensureAuthed()
      if (!authed) {
        return false
      }

      await trackStoreBehaviorEvent(params.merchant_id, {
        openid: getOpenid(),
        event_type: params.event,
        page: params.page,
        product_id: params.product_id,
        order_id: params.order_id,
        source: params.source || 'store',
        payload: params.data
      })
      return true
    } catch (error) {
      console.error('Analytics: 事件埋点异常', error)
      return false
    }
  }

  const trackPageView = async (page: string, merchantId: number, source = 'store') => {
    return await trackEvent({
      event: 'page_view',
      merchant_id: merchantId,
      page,
      source,
      data: { page }
    })
  }

  const trackProductView = async (merchantId: number, productId: number) => {
    return await trackEvent({
      event: 'product_view',
      merchant_id: merchantId,
      product_id: productId,
      page: 'store_product',
      data: { product_id: productId }
    })
  }

  const trackCheckout = async (merchantId: number, amount: number) => {
    return await trackEvent({
      event: 'submit_order',
      merchant_id: merchantId,
      page: 'store_confirm',
      data: { amount }
    })
  }

  const trackPayment = async (merchantId: number, orderId: number, amount: number) => {
    return await trackEvent({
      event: 'pay_success',
      merchant_id: merchantId,
      order_id: orderId,
      page: 'store_payment_result',
      data: { order_id: orderId, amount }
    })
  }

  return {
    trackVisit,
    trackEvent,
    trackPageView,
    trackProductView,
    trackCheckout,
    trackPayment
  }
}

export default useAnalytics
