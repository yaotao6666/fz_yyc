/**
 * 订单列表共享逻辑
 * 承载实物订单(my-orders-goods)与服务订单(my-orders-service)两页共用的
 * 列表加载、分页、状态筛选、卡片操作逻辑。category 固定传入：1=实物 2=服务。
 */
import { ref } from 'vue'
import { onShow } from '@dcloudio/uni-app'
import { getMyOrders, cancelMyOrder } from '@api'
import { OrderStatus, OrderStatusText } from '@types'
import type { Order } from '@types'
import { BrandAsset } from '../../../utils/constants'
import { useAuth } from '../../../utils/useAuth'

export function useOrderList(category: number) {
  const currentStatus = ref(0)
  const orders = ref<Order[]>([])
  const loading = ref(false)
  const noMore = ref(false)
  const page = ref(1)
  const pageSize = 10

  const statusTabs = [
    { label: '全部', value: 0 },
    { label: '待支付', value: OrderStatus.PENDING_PAYMENT },
    { label: '已支付', value: OrderStatus.PAID },
    { label: '已完成', value: OrderStatus.COMPLETED }
  ]

  onShow(async () => {
    const pages = getCurrentPages()
    const currentPage = pages[pages.length - 1] as any
    const status = currentPage?.options?.status

    const { ensureAuth } = useAuth()
    await ensureAuth()

    if (status) {
      const parsed = Number(status)
      if (Number.isFinite(parsed)) {
        currentStatus.value = parsed
      } else if (status === 'paid') {
        currentStatus.value = OrderStatus.PAID
      } else if (status === 'pending') {
        currentStatus.value = OrderStatus.PENDING_PAYMENT
      } else {
        currentStatus.value = 0
      }
    }

    loadOrders(true)
  })

  async function loadOrders(reset = false) {
    if (reset) {
      page.value = 1
      noMore.value = false
      orders.value = []
    }

    if (noMore.value || loading.value) return

    loading.value = true

    try {
      const params: any = {
        page: page.value,
        page_size: pageSize
      }

      if (currentStatus.value !== 0) {
        params.status = currentStatus.value
      }

      if (category !== 0) {
        params.category = category
      }

      const res = await getMyOrders(params)

      if (reset) {
        orders.value = res.list
      } else {
        orders.value.push(...res.list)
      }

      if (res.list.length < pageSize) {
        noMore.value = true
      } else {
        page.value++
      }
    } catch (error) {
      console.error('加载订单失败:', error)
    } finally {
      loading.value = false
    }
  }

  function loadMore() {
    loadOrders()
  }

  function changeStatus(status: number) {
    currentStatus.value = status
    loadOrders(true)
  }

  function getStatusText(status: number): string {
    return OrderStatusText[status] || '未知'
  }

  function getStatusClass(status: number): string {
    const classMap: Record<number, string> = {
      [OrderStatus.PENDING_PAYMENT]: 'pending',
      [OrderStatus.PAID]: 'paid',
      [OrderStatus.COMPLETED]: 'completed',
      [OrderStatus.CANCELLED]: 'cancelled',
      [OrderStatus.REFUNDING]: 'refunding',
      [OrderStatus.REFUNDED]: 'refunded'
    }
    return classMap[status] || ''
  }

  function formatTime(time: string): string {
    const date = new Date(time)
    return `${date.getMonth() + 1}-${date.getDate()} ${date.getHours()}:${String(date.getMinutes()).padStart(2, '0')}`
  }

  function getOrderItemImage(item: any) {
    if (typeof item?.image === 'string' && item.image.trim()) {
      return item.image
    }

    if (Array.isArray(item?.images) && typeof item.images[0] === 'string' && item.images[0].trim()) {
      return item.images[0]
    }

    if (typeof item?.product_image === 'string' && item.product_image.trim()) {
      return item.product_image
    }

    return BrandAsset.DEFAULT_PRODUCT_IMAGE
  }

  function getDeliveryAddressText(order: Order): string {
    return order.delivery_info?.address || order.delivery_address || ''
  }

  function goDetail(orderId: number) {
    uni.navigateTo({ url: `/pages/store/order-detail?id=${orderId}` })
  }

  function goReview(order: Order) {
    uni.navigateTo({ url: `/pages/store/review-submit?id=${order.id}` })
  }

  function goShopping() {
    uni.switchTab({ url: `/pages/store/home` })
  }

  function cancelOrder(order: Order) {
    uni.showModal({
      title: '确认取消',
      content: '确定要取消该订单吗？',
      success: async (res) => {
        if (res.confirm) {
          try {
            await cancelMyOrder(order.id)
            const index = orders.value.findIndex(o => o.id === order.id)
            if (index !== -1) {
              orders.value[index].status = OrderStatus.CANCELLED
            }
            uni.showToast({ title: '订单已取消', icon: 'success' })
          } catch (error: any) {
            uni.showToast({ title: error.message || '取消失败', icon: 'none' })
          }
        }
      }
    })
  }

  function contactMerchantForRefund(order: Order) {
    const phone = order.merchant?.phone?.trim()
    const merchantName = order.merchant?.name || '商家'

    if (!phone) {
      uni.showModal({
        title: '联系商家退款',
        content: `请联系${merchantName}协助处理退款。`,
        showCancel: false,
        confirmText: '我知道了'
      })
      return
    }

    uni.showModal({
      title: '联系商家退款',
      content: `请联系${merchantName}退款\n联系电话：${phone}`,
      confirmText: '拨打电话',
      cancelText: '取消',
      success: (res) => {
        if (res.confirm) {
          uni.makePhoneCall({ phoneNumber: phone })
        }
      }
    })
  }

  return {
    statusTabs,
    currentStatus,
    orders,
    loading,
    noMore,
    loadMore,
    changeStatus,
    getStatusText,
    getStatusClass,
    formatTime,
    getOrderItemImage,
    getDeliveryAddressText,
    goDetail,
    goReview,
    goShopping,
    cancelOrder,
    contactMerchantForRefund
  }
}