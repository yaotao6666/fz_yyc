import type { UserAddress } from '@types'

export const STORE_SELECTED_ADDRESS_ID_KEY = 'store:selectedAddressId'
export const STORE_ADDRESS_LIST_REFRESH_KEY = 'store:addressListRefresh'
export const STORE_EDIT_ADDRESS_KEY = 'store:editAddress'

export interface WechatChooseAddressResult {
  userName: string
  telNumber: string
  provinceName: string
  cityName: string
  countyName: string
  detailInfo: string
}

export function formatUserAddress(address?: Partial<UserAddress> | null) {
  if (!address) return ''
  return [address.province, address.city, address.district, address.address].filter(Boolean).join('')
}

export function buildUserAddressPayload(address: Partial<UserAddress>): Partial<UserAddress> {
  const payload: Partial<UserAddress> = {
    name: (address.name || '').trim(),
    phone: (address.phone || '').trim(),
    province: (address.province || '').trim(),
    city: (address.city || '').trim(),
    district: (address.district || '').trim(),
    address: (address.address || '').trim(),
    is_default: Boolean(address.is_default)
  }

  if (typeof address.lat === 'number') {
    payload.lat = address.lat
  }
  if (typeof address.lng === 'number') {
    payload.lng = address.lng
  }

  return payload
}

export function mapWechatAddressToUserAddress(result: WechatChooseAddressResult): Partial<UserAddress> {
  return {
    name: result.userName || '',
    phone: result.telNumber || '',
    province: result.provinceName || '',
    city: result.cityName || '',
    district: result.countyName || '',
    address: result.detailInfo || ''
  }
}

export function pickAddressById(addresses: UserAddress[], addressId?: number | null) {
  if (!addressId) return null
  return addresses.find((item) => item.id === addressId) || null
}

export function getPreferredAddress(addresses: UserAddress[], preferredAddressId?: number | null) {
  return (
    pickAddressById(addresses, preferredAddressId) ||
    addresses.find((item) => item.is_default) ||
    addresses[0] ||
    null
  )
}

export function chooseWechatAddress() {
  return new Promise<WechatChooseAddressResult>((resolve, reject) => {
    const chooseAddress = (uni as any).chooseAddress
    if (typeof chooseAddress !== 'function') {
      reject(new Error('当前环境不支持微信地址导入'))
      return
    }

    chooseAddress({
      success: (res: WechatChooseAddressResult) => resolve(res),
      fail: (err: any) => reject(err)
    })
  })
}

export function isChooseAddressCancel(error: any) {
  const message = String(error?.errMsg || error?.message || '')
  return message.includes('cancel')
}
