// Basic API client without gRPC for initial testing
export interface Model {
  name: string
  app: string
  verbose_name: string
  verbose_name_plural: string
  list_display: string[]
  search_fields: string[]
  list_filter: string[]
  permissions: {
    can_add: boolean
    can_change: boolean
    can_delete: boolean
    can_view: boolean
  }
}

export interface Site {
  name: string
  header_title: string
  index_title: string
}

export interface ModelsResponse {
  models: Record<string, Model>
  site: Site
}

class AdminAPI {
  private baseURL = '/admin/api'

  async getModels(): Promise<ModelsResponse> {
    const response = await fetch(`${this.baseURL}/models/`)
    if (!response.ok) {
      throw new Error('Failed to fetch models')
    }
    return response.json()
  }
}

export const adminAPI = new AdminAPI()