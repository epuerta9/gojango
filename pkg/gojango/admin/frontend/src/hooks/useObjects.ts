import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { adminClient } from '@/services/client'

export function useListObjects(
  app: string,
  model: string,
  options: {
    page?: number
    perPage?: number
    query?: string
    filters?: Record<string, string>
  } = {}
) {
  return useQuery({
    queryKey: ['objects', app, model, options],
    queryFn: async () => {
      const response = await adminClient.listObjects({
        app,
        model,
        page: options.page,
        pageSize: options.perPage,
        search: options.query,
        filters: options.filters,
      })
      return response
    },
    enabled: !!(app && model),
  })
}

export function useGetObject(app: string, model: string, id: string) {
  return useQuery({
    queryKey: ['object', app, model, id],
    queryFn: async () => {
      const response = await adminClient.getObject(app, model, id)
      return response
    },
    enabled: !!(app && model && id),
  })
}

export function useCreateObject(app: string, model: string) {
  const queryClient = useQueryClient()
  
  return useMutation({
    mutationFn: async (data: Record<string, any>) => {
      const response = await adminClient.createObject(app, model, data)
      return response
    },
    onSuccess: () => {
      // Invalidate the objects list
      queryClient.invalidateQueries({ queryKey: ['objects', app, model] })
    },
  })
}

export function useUpdateObject(app: string, model: string, id: string) {
  const queryClient = useQueryClient()
  
  return useMutation({
    mutationFn: async (data: Record<string, any>) => {
      const response = await adminClient.updateObject(app, model, id, data)
      return response
    },
    onSuccess: () => {
      // Invalidate both the objects list and the specific object
      queryClient.invalidateQueries({ queryKey: ['objects', app, model] })
      queryClient.invalidateQueries({ queryKey: ['object', app, model, id] })
    },
  })
}

export function useDeleteObject(app: string, model: string) {
  const queryClient = useQueryClient()
  
  return useMutation({
    mutationFn: async (id: string) => {
      await adminClient.deleteObject(app, model, id)
      return true
    },
    onSuccess: () => {
      // Invalidate the objects list
      queryClient.invalidateQueries({ queryKey: ['objects', app, model] })
    },
  })
}

export function useExecuteAction(app: string, model: string) {
  const queryClient = useQueryClient()
  
  return useMutation({
    mutationFn: async (params: { action: string; selectedIds: string[] }) => {
      const response = await adminClient.executeAction(app, model, params.action, params.selectedIds)
      return response
    },
    onSuccess: () => {
      // Invalidate the objects list
      queryClient.invalidateQueries({ queryKey: ['objects', app, model] })
    },
  })
}