// Generated protobuf client using Connect-ES
import { createClient } from "@connectrpc/connect"
import { createConnectTransport } from "@connectrpc/connect-web"
import { AdminService } from "../gen/admin_connect"

// Create the transport
const transport = createConnectTransport({
  baseUrl: "/admin",
  useBinaryFormat: false,
})

// Create the client
const client = createClient(AdminService, transport)

// Export the client with a cleaner interface
export const connectClient = {
  listModels: () => client.listModels({}),
  
  listObjects: (params: {
    app: string
    model: string
    page?: number
    pageSize?: number
    search?: string
    filters?: Record<string, string>
  }) => client.listObjects({
    app: params.app,
    model: params.model,
    page: params.page || 1,
    pageSize: params.pageSize || 25,
    search: params.search || "",
    filters: params.filters || {},
  }),
  
  getObject: (app: string, model: string, id: string) => 
    client.getObject({ app, model, id }),
  
  createObject: (app: string, model: string, data: Record<string, any>) =>
    client.createObject({ app, model, data }),
  
  updateObject: (app: string, model: string, id: string, data: Record<string, any>) =>
    client.updateObject({ app, model, id, data }),
  
  deleteObject: (app: string, model: string, id: string) =>
    client.deleteObject({ app, model, id }),
  
  executeAction: (app: string, model: string, action: string, selectedIds: string[]) =>
    client.executeAction({ app, model, action, objectIds: selectedIds }),
}

// Re-export generated types
export type { 
  ModelInfo, 
  ObjectData, 
  ListObjectsRequest, 
  ListObjectsResponse,
  ListModelsResponse 
} from '../gen/admin_pb'