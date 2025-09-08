// Admin API Types

export interface Site {
  name: string
  header_title: string
  index_title: string
  site_url: string
}

export interface Model {
  name: string
  app: string
  verbose_name: string
  verbose_name_plural: string
  list_display: string[]
  search_fields: string[]
  list_filter: string[]
  actions: string[]
}

export interface ModelsResponse {
  site: Site
  models: Record<string, Model>
}

export interface AdminObject {
  id: number | string
  [key: string]: any
}

export interface ListResponse<T = AdminObject> {
  objects: T[]
  count: number
  page: number
  page_size: number
  has_next: boolean
  has_previous: boolean
}

export interface ActionResponse {
  success: boolean
  message: string
  count?: number
}