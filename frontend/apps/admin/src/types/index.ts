export type Role   = 'admin' | 'manager' | 'staff' | 'customer'
export type Status = 'pending'|'confirmed'|'processing'|'shipped'|'delivered'|'cancelled'|'refunded'

export interface Branch {
  id: string; name: string; address: Address
  warehouse_mode: boolean; settings: BranchSettings; is_active: boolean
}
export interface BranchSettings { tax_rate: number; currency: string; open_hours?: string }
export interface Address { street: string; city: string; state: string; zip: string; country: string }

export interface User {
  id: string; email: string; role: Role; branch_id?: string
  first_name: string; last_name: string; mfa_enabled: boolean; is_active: boolean
}

export interface Product {
  id: string; sku: string; name: string; description: string
  base_price: number; branch_price?: number
  category_id?: string; images: ProductImage[]
  tags: string[]; is_active: boolean; stock?: number
}
export interface ProductImage { url: string; alt_text: string; is_main: boolean }

export interface Order {
  id: string; branch_id: string; customer_id: string; status: Status
  items: OrderItem[]; subtotal: number; tax: number; discount: number; total: number
  currency: string; created_at: string
}
export interface OrderItem { product_id: string; sku: string; name: string; unit_price: number; quantity: number; line_total: number }

export interface Page<T> { data: T[]; total: number; page: number; page_size: number; total_pages: number }
export interface ApiError { code: string; message: string; fields?: Record<string,string> }
