/**
 * 用户行为埋点工具
 * 用于记录用户访问、点击等行为事件
 */

import { useAuth } from './useAuth'

interface VisitParams {
  merchant_id: number
  source?: string
}

interface TrackEventParams {
  event: string
  merchant_id?: number
  data?: Record<string, any>
}

export function useAnalytics() {
  const { getOpenid } = useAuth()

  const trackVisit = async (params: VisitParams): Promise<boolean> => {
    try {
      const openid = getOpenid()

      if (!openid) {
        console.log('Analytics: 用户未登录，跳过访问埋点')
        return false
      }

      const res = await uni.request({
        url: `http://localhost:8080/api/v1/store/${params.merchant_id}/visit`,
        method: 'POST',
        data: {
          openid,
          source: params.source || 'scan'
        },
        header: {
          'Content-Type': 'application/json'
        }
      }) as { data: { code: number; data: { user_id: number; visit_count: number } } }

      if (res.data?.code === 0) {
        console.log('Analytics: 访问埋点已记录', {
          merchant_id: params.merchant_id,
          user_id: res.data.data.user_id,
          visit_count: res.data.data.visit_count
        })
        return true
      }

      console.error('Analytics: 访问埋点记录失败', res.data)
      return false
    } catch (error) {
      console.error('Analytics: 访问埋点异常', error)
      return false
    }
  }

  const trackEvent = async (params: TrackEventParams): Promise<boolean> => {
    try {
      const openid = getOpenid()

      if (!openid) {
        console.log('Analytics: 用户未登录，跳过事件埋点')
        return false
      }

      console.log('Analytics: 事件埋点', {
        event: params.event,
        merchant_id: params.merchant_id,
        data: params.data
      })

      return true
    } catch (error) {
      console.error('Analytics: 事件埋点异常', error)
      return false
    }
  }

  const trackPageView = async (page: string, merchantId?: number) => {
    return await trackEvent({
      event: 'page_view',
      merchant_id: merchantId,
      data: { page }
    })
  }

  const trackAddToCart = async (merchantId: number, productId: number, quantity: number) => {
    return await trackEvent({
      event: 'add_to_cart',
      merchant_id: merchantId,
      data: { product_id: productId, quantity }
    })
  }

  const trackCheckout = async (merchantId: number, amount: number) => {
    return await trackEvent({
      event: 'checkout',
      merchant_id: merchantId,
      data: { amount }
    })
  }

  const trackPayment = async (merchantId: number, orderId: number, amount: number) => {
    return await trackEvent({
      event: 'payment',
      merchant_id: merchantId,
      data: { order_id: orderId, amount }
    })
  }

  return {
    trackVisit,
    trackEvent,
    trackPageView,
    trackAddToCart,
    trackCheckout,
    trackPayment
  }
}

export default useAnalytics
