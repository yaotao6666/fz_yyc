export interface ServiceProviderLoginRequest {
  username: string
  password: string
}

export interface ServiceProviderInfo {
  id: number
  name: string
  sp_name: string
}

export interface ServiceProviderLoginResponse {
  token: string
  service_provider: ServiceProviderInfo
}

export interface DashboardData {
  total_merchants: number
  pending_merchants?: number
  today_orders: number
  today_revenue: number
  distribution?: Array<{ category: string; count: number }>
  trend?: Array<{ date: string; orders: number }>
}

export interface SpSettings {
  service_provider_id?: number
  name: string
  sp_name: string
  contact_phone: string
  contact_email?: string
  created_at?: string
}

export interface MerchantListItem {
  id: number
  name: string
  logo?: string
  cover_image?: string
  contact_name?: string
  contact_phone?: string
  contact_email?: string
  address?: string
  business_category?: string
  business_hours?: string
  announcement?: string
  sub_mch_id?: string
  profit_sharing_enabled?: boolean
  profit_sharing_ratio?: number
  payment_config_status?: number
  status: number
  created_at?: string
  total_users: number
  total_orders: number
  total_amount: number
}

export interface MerchantDetail extends MerchantListItem {
  qrcode_url?: string
  admin_staff_id?: number
  admin_username?: string
  admin_name?: string
  admin_phone?: string
  admin_status?: number
}

export interface MerchantListResponse {
  list: MerchantListItem[]
  pagination: {
    total: number
    page: number
    page_size: number
  }
}

export interface SpMerchantFormData {
  name: string
  contact_name?: string
  contact_phone?: string
  contact_email?: string
  address?: string
  business_category?: string
  business_hours?: string
  announcement?: string
  username: string
  password: string
  staff_name?: string
  staff_phone?: string
  sub_mch_id?: string
  profit_sharing_enabled: boolean
  profit_sharing_ratio: number
}

export interface UpdateSpMerchantFormData {
  name?: string
  contact_name?: string
  contact_phone?: string
  contact_email?: string
  address?: string
  business_category?: string
  business_hours?: string
  announcement?: string
  status?: number
}

export interface MerchantPaymentConfigFormData {
  sub_mch_id: string
  profit_sharing_enabled: boolean
  profit_sharing_ratio: number
}

export interface ProfitSharingRecord {
  id: number
  service_provider_id: number
  merchant_id: number
  order_id: number
  order_no: string
  transaction_id?: string
  profit_sharing_order_no: string
  profit_sharing_date: string
  pay_amount: number
  profit_sharing_ratio: number
  profit_sharing_amount: number
  merchant_received_amount: number
  status: number
  error_message?: string
  created_at: string
}

export interface ProfitSharingRecordListResponse {
  list: ProfitSharingRecord[]
  pagination: {
    total: number
    page: number
    page_size: number
  }
}

export interface SpMerchantConversionItem {
  merchant_id: number
  merchant_name: string
  merchant_logo?: string
  visit_users: number
  order_users: number
  paid_orders: number
  order_amount: number
  avg_order_amount: number
  visit_rate: number
  order_rate: number
}

export interface MerchantDistributionData {
  merchants: SpMerchantConversionItem[]
  totals: {
    merchant_count: number
    visit_users: number
    order_users: number
    paid_orders: number
    order_amount: number
  }
  pagination?: {
    total: number
    page: number
    page_size: number
  }
}

export interface OrderAnalyticsBucket {
  label: string
  order_count: number
}

export interface OrderAnalyticsData {
  day: OrderAnalyticsBucket[]
  week: OrderAnalyticsBucket[]
  month: OrderAnalyticsBucket[]
  year: OrderAnalyticsBucket[]
}

export interface AmountTrendData {
  trends: Array<{
    date: string
    amount: number
  }>
}

export interface TopMerchantRanking {
  rank: number
  metric: string
  merchant_id: number
  merchant_name: string
  merchant_logo?: string
  visit_rate: number
  order_rate: number
  order_amount: number
  avg_order_amount: number
  visit_users: number
  order_users: number
  paid_orders: number
}

export interface UploadTokenResponse {
  token: string
  domain: string
  prefix: string
  upload_url?: string
}

export interface AnnouncementItem {
  id: number
  service_provider_id: number
  title: string
  content: string
  status: number
  created_at: string
  updated_at: string
}

export interface AnnouncementListResponse {
  list: AnnouncementItem[]
  pagination: {
    total: number
    page: number
    page_size: number
  }
}

export interface AnnouncementFormData {
  title: string
  content: string
  status: number
}
