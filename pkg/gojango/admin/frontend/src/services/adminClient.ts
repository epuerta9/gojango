// TypeScript Connect client for Gojango Admin
// This is a hand-written client that will be replaced by generated code

// Types matching our protobuf definitions
export interface ModelInfo {
  app: string
  name: string
  verboseName: string
  verboseNamePlural: string
  listDisplay: string[]
  searchFields: string[]
  listFilter: string[]
  readonlyFields: string[]
  exclude: string[]
  permissions: ModelPermissions
  actions: AdminAction[]
  listPerPage: number
  ordering: string
  showFullResultCount: boolean
}

export interface ModelPermissions {
  add: boolean
  change: boolean
  delete: boolean
  view: boolean
}

export interface AdminAction {
  name: string
  description: string
  confirmationRequired: boolean
  permissions: string[]
}

export interface ObjectData {
  id: string
  fields: Record<string, any>
  strRepresentation: string
  createdAt?: Date
  updatedAt?: Date
}

export interface ListObjectsRequest {
  app: string
  model: string
  page?: number
  pageSize?: number
  ordering?: string
  filters?: Record<string, string>
  search?: string
}

export interface ListObjectsResponse {
  objects: ObjectData[]
  totalCount: number
  page: number
  pageSize: number
  hasNext: boolean
  hasPrevious: boolean
  totalPages: number
  displayFields: string[]
}

export interface ListModelsResponse {
  models: Record<string, ModelInfo>
  site: {
    name: string
    headerTitle: string
    indexTitle: string
  }
}


// Admin service client implementation
export class AdminClient {
  private baseUrl = '/admin'

  async listModels(): Promise<ListModelsResponse> {
    // For now, use the existing REST endpoint
    // TODO: Replace with gRPC call once protobuf generation is complete
    const response = await fetch(`${this.baseUrl}/api/models/`)
    if (!response.ok) {
      throw new Error(`Failed to fetch models: ${response.statusText}`)
    }
    return response.json()
  }

  async listObjects(request: ListObjectsRequest): Promise<ListObjectsResponse> {
    // TODO: Implement gRPC call to AdminService.ListObjects
    // For now, return mock data that matches our protobuf structure
    const mockObjects: ObjectData[] = []
    
    const pageSize = request.pageSize || 25
    const page = request.page || 1
    
    // Generate mock data based on model type
    for (let i = 0; i < pageSize; i++) {
      const id = (page - 1) * pageSize + i + 1
      let fields: Record<string, any> = {}
      
      switch (request.model) {
        case 'user':
          fields = {
            id,
            username: `user${id}`,
            email: `user${id}@example.com`,
            firstName: `First${id}`,
            lastName: `Last${id}`,
            isActive: id % 2 === 1,
            isStaff: id % 5 === 0,
            createdAt: new Date(Date.now() - Math.random() * 10000000000),
          }
          break
        case 'post':
          fields = {
            id,
            title: `Sample Post Title ${id}`,
            content: `This is the content for post ${id}. Lorem ipsum dolor sit amet...`,
            status: id % 3 === 0 ? 'draft' : 'published',
            authorId: Math.floor(Math.random() * 10) + 1,
            createdAt: new Date(Date.now() - Math.random() * 10000000000),
          }
          break
        case 'category':
          fields = {
            id,
            name: `Category ${id}`,
            slug: `category-${id}`,
            description: `Description for category ${id}`,
            createdAt: new Date(Date.now() - Math.random() * 10000000000),
          }
          break
      }
      
      mockObjects.push({
        id: id.toString(),
        fields,
        strRepresentation: `${request.model}: ${fields.name || fields.title || fields.username || `Object ${id}`}`,
        createdAt: fields.createdAt,
        updatedAt: fields.updatedAt,
      })
    }

    const totalCount = 150 // Mock total count
    const totalPages = Math.ceil(totalCount / pageSize)

    return {
      objects: mockObjects,
      totalCount,
      page,
      pageSize,
      hasNext: page < totalPages,
      hasPrevious: page > 1,
      totalPages,
      displayFields: this.getDisplayFieldsForModel(request.model),
    }
  }

  private getDisplayFieldsForModel(model: string): string[] {
    switch (model) {
      case 'user':
        return ['id', 'username', 'email', 'isActive', 'createdAt']
      case 'post':
        return ['id', 'title', 'status', 'authorId', 'createdAt']
      case 'category':
        return ['id', 'name', 'slug', 'createdAt']
      default:
        return ['id', 'createdAt']
    }
  }

  async getObject(_app: string, _model: string, _id: string): Promise<ObjectData> {
    // TODO: Implement gRPC call
    throw new Error('GetObject not implemented yet')
  }

  async createObject(_app: string, _model: string, _data: Record<string, any>): Promise<ObjectData> {
    // TODO: Implement gRPC call
    throw new Error('CreateObject not implemented yet')
  }

  async updateObject(_app: string, _model: string, _id: string, _data: Record<string, any>): Promise<ObjectData> {
    // TODO: Implement gRPC call
    throw new Error('UpdateObject not implemented yet')
  }

  async deleteObject(_app: string, _model: string, _id: string): Promise<void> {
    // TODO: Implement gRPC call
    throw new Error('DeleteObject not implemented yet')
  }

  async deleteObjects(_app: string, _model: string, _ids: string[]): Promise<void> {
    // TODO: Implement gRPC call
    throw new Error('DeleteObjects not implemented yet')
  }

  async executeAction(_app: string, _model: string, _action: string, _objectIds: string[], _parameters?: Record<string, any>): Promise<any> {
    // TODO: Implement gRPC call
    throw new Error('ExecuteAction not implemented yet')
  }

  async listActions(_app: string, _model: string): Promise<AdminAction[]> {
    // TODO: Implement gRPC call
    return []
  }

  async searchObjects(_app: string, _model: string, _query: string, _limit?: number): Promise<ObjectData[]> {
    // TODO: Implement gRPC call
    return []
  }
}

// Export singleton instance
export const adminClient = new AdminClient()