export interface Product {
  id: string; sku: string; name: string; description: string
  base_price: number; branch_price?: number
  category_id?: string; images: { url: string; alt_text: string; is_main: boolean }[]
  tags: string[]; is_active: boolean; stock?: number
}

export interface Category {
  id: string; name: string; slug: string; parent_id?: string
}

export interface StoreConfig {
  store_name: string; logo_url?: string; currency: string
  tax_rate: number; contact_email: string; support_phone?: string; theme: string
}

export interface Order {
  id: string; status: string; items: OrderItem[]
  subtotal: number; tax: number; discount: number; total: number
  currency: string; created_at: string
  refund_status?: string; refund_reason?: string; refunded_at?: string
}

export interface OrderItem {
  product_id: string; name: string; sku: string
  unit_price: number; quantity: number; line_total: number
}

export interface Branch {
  id: string; name: string; is_active: boolean
}
