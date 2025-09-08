// Re-export the mock client for now (Connect client has build issues)
export { adminClient } from './adminClient'
export type { 
  ModelInfo, 
  ObjectData, 
  ListObjectsRequest, 
  ListObjectsResponse,
  ListModelsResponse,
  AdminAction 
} from './adminClient'