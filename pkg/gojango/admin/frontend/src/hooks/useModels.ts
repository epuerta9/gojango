import { useQuery } from '@tanstack/react-query'
import { adminClient } from '@/services/client'

export function useModels() {
  return useQuery({
    queryKey: ['models'],
    queryFn: async () => {
      return await adminClient.listModels()
    },
  })
}

export function useModelSchema(app: string, model: string) {
  return useQuery({
    queryKey: ['model-schema', app, model],
    queryFn: async () => {
      // Mock schema for now
      return {
        fields: [
          { name: 'id', type: 'integer', required: true },
          { name: 'created_at', type: 'datetime' },
        ],
        relations: [],
      }
    },
    enabled: !!(app && model),
  })
}