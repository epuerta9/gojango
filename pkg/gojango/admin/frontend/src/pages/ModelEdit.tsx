import { useParams, useNavigate } from 'react-router-dom'
import { useState, useEffect } from 'react'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { adminClient } from '@/services/client'
import { ArrowLeft, Save, Trash2 } from 'lucide-react'

export function ModelEdit() {
  const { app, model, id } = useParams<{ app: string; model: string; id: string }>()
  const navigate = useNavigate()
  const queryClient = useQueryClient()
  const [formData, setFormData] = useState<Record<string, any>>({})
  const [errors, setErrors] = useState<Record<string, string>>({})

  // Get object data
  const { data: objectData, isLoading: isLoadingObject } = useQuery({
    queryKey: ['object', app, model, id],
    queryFn: async () => {
      return adminClient.getObject(app!, model!, id!)
    },
    enabled: !!(app && model && id),
  })

  // Get model schema for form fields
  const { data: schemaData, isLoading: isLoadingSchema } = useQuery({
    queryKey: ['modelSchema', app, model],
    queryFn: async () => {
      // TODO: Use the actual getModelSchema method when available
      // For now, return mock schema based on model type
      if (model === 'user') {
        return {
          fields: [
            { name: 'username', fieldType: 'string', required: true, verboseName: 'Username' },
            { name: 'email', fieldType: 'string', required: true, verboseName: 'Email' },
            { name: 'firstName', fieldType: 'string', required: false, verboseName: 'First Name' },
            { name: 'lastName', fieldType: 'string', required: false, verboseName: 'Last Name' },
            { name: 'isActive', fieldType: 'boolean', required: false, verboseName: 'Active' },
            { name: 'isStaff', fieldType: 'boolean', required: false, verboseName: 'Staff' },
          ]
        }
      }
      return { fields: [] }
    },
    enabled: !!(app && model),
  })

  // Initialize form data when object data loads
  useEffect(() => {
    if (objectData && 'fields' in objectData) {
      const initialData: Record<string, any> = {}
      Object.entries((objectData as any).fields).forEach(([key, value]: [string, any]) => {
        // Convert protobuf Value to native JS value
        if (value?.kind === 'boolValue') {
          initialData[key] = value.boolValue
        } else if (value?.kind === 'numberValue') {
          initialData[key] = value.numberValue
        } else if (value?.kind === 'stringValue') {
          initialData[key] = value.stringValue
        } else {
          initialData[key] = value
        }
      })
      setFormData(initialData)
    }
  }, [objectData])

  // Update mutation
  const updateMutation = useMutation({
    mutationFn: async (data: Record<string, any>) => {
      return adminClient.updateObject(app!, model!, id!, data)
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['objects', app, model] })
      queryClient.invalidateQueries({ queryKey: ['object', app, model, id] })
      navigate(`/admin/${app}/${model}/`)
    },
    onError: (error: any) => {
      console.error('Failed to update object:', error)
      setErrors({ general: 'Failed to update object. Please try again.' })
    },
  })

  // Delete mutation
  const deleteMutation = useMutation({
    mutationFn: async () => {
      return adminClient.deleteObject(app!, model!, id!)
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['objects', app, model] })
      navigate(`/admin/${app}/${model}/`)
    },
    onError: (error: any) => {
      console.error('Failed to delete object:', error)
      setErrors({ general: 'Failed to delete object. Please try again.' })
    },
  })

  const handleInputChange = (fieldName: string, value: any) => {
    setFormData(prev => ({
      ...prev,
      [fieldName]: value
    }))
    // Clear error when user starts typing
    if (errors[fieldName]) {
      setErrors(prev => {
        const newErrors = { ...prev }
        delete newErrors[fieldName]
        return newErrors
      })
    }
  }

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault()
    
    // Validate required fields
    const newErrors: Record<string, string> = {}
    schemaData?.fields?.forEach(field => {
      if (field.required && (!formData[field.name] || formData[field.name] === '')) {
        newErrors[field.name] = `${field.verboseName || field.name} is required`
      }
    })

    if (Object.keys(newErrors).length > 0) {
      setErrors(newErrors)
      return
    }

    updateMutation.mutate(formData)
  }

  const handleDelete = () => {
    if (window.confirm(`Are you sure you want to delete this ${model}? This action cannot be undone.`)) {
      deleteMutation.mutate()
    }
  }

  const renderField = (field: any) => {
    const fieldName = field.name
    const fieldValue = formData[fieldName] ?? ''
    
    switch (field.fieldType) {
      case 'boolean':
        return (
          <div key={fieldName} className="space-y-2">
            <label htmlFor={fieldName} className="text-sm font-medium text-foreground">
              {field.verboseName || field.name}
            </label>
            <div className="flex items-center space-x-2">
              <input
                id={fieldName}
                type="checkbox"
                checked={Boolean(fieldValue)}
                onChange={(e) => handleInputChange(fieldName, e.target.checked)}
                className="h-4 w-4 text-primary focus:ring-primary border-gray-300 rounded"
              />
              <span className="text-sm text-muted-foreground">
                {field.helpText || `Enable ${field.verboseName || field.name}`}
              </span>
            </div>
            {errors[fieldName] && (
              <p className="text-sm text-red-600">{errors[fieldName]}</p>
            )}
          </div>
        )
      
      default: // string, integer, etc.
        return (
          <div key={fieldName} className="space-y-2">
            <label htmlFor={fieldName} className="text-sm font-medium text-foreground">
              {field.verboseName || field.name}
              {field.required && <span className="text-red-500 ml-1">*</span>}
            </label>
            <Input
              id={fieldName}
              type={field.fieldType === 'integer' ? 'number' : 'text'}
              value={fieldValue}
              onChange={(e) => handleInputChange(fieldName, e.target.value)}
              className={errors[fieldName] ? 'border-red-500' : ''}
              placeholder={field.helpText || `Enter ${field.verboseName || field.name}`}
            />
            {errors[fieldName] && (
              <p className="text-sm text-red-600">{errors[fieldName]}</p>
            )}
          </div>
        )
    }
  }

  const isLoading = isLoadingObject || isLoadingSchema

  if (isLoading) {
    return (
      <div className="space-y-6">
        <div className="animate-pulse">
          <div className="h-8 bg-gray-200 rounded w-1/4 mb-2"></div>
          <div className="h-4 bg-gray-200 rounded w-1/2"></div>
        </div>
        <Card>
          <CardContent className="p-6">
            <div className="animate-pulse space-y-4">
              <div className="h-4 bg-gray-200 rounded w-1/4"></div>
              <div className="h-10 bg-gray-200 rounded"></div>
              <div className="h-4 bg-gray-200 rounded w-1/4"></div>
              <div className="h-10 bg-gray-200 rounded"></div>
            </div>
          </CardContent>
        </Card>
      </div>
    )
  }

  if (!objectData) {
    return (
      <div className="space-y-6">
        <div>
          <h1 className="text-3xl font-bold text-foreground capitalize">
            {model} not found
          </h1>
          <p className="text-muted-foreground mt-2">
            The requested {model} object could not be found.
          </p>
        </div>
        <Card>
          <CardContent className="p-6">
            <div className="text-center">
              <p className="text-muted-foreground mb-4">Object #{id} not found</p>
              <Button onClick={() => navigate(`/admin/${app}/${model}/`)}>
                <ArrowLeft className="w-4 h-4 mr-2" />
                Back to List
              </Button>
            </div>
          </CardContent>
        </Card>
      </div>
    )
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <div className="flex items-center space-x-2">
            <Button
              variant="ghost"
              size="sm"
              onClick={() => navigate(`/admin/${app}/${model}/`)}
            >
              <ArrowLeft className="w-4 h-4 mr-1" />
              Back
            </Button>
          </div>
          <h1 className="text-3xl font-bold text-foreground capitalize mt-2">
            Edit {model} #{id}
          </h1>
          <p className="text-muted-foreground mt-2">
            Edit {model} object in the {app} app
          </p>
        </div>
      </div>
      
      {errors.general && (
        <div className="bg-red-50 border border-red-200 rounded-md p-4">
          <p className="text-red-800">{errors.general}</p>
        </div>
      )}

      <Card>
        <CardHeader>
          <CardTitle>Object Details</CardTitle>
        </CardHeader>
        <CardContent>
          <form onSubmit={handleSubmit} className="space-y-6">
            <div className="grid grid-cols-1 gap-6">
              {schemaData?.fields?.map(renderField)}
            </div>
            
            <div className="flex items-center justify-between pt-6 border-t">
              <div className="flex items-center space-x-2">
                <Button
                  type="submit"
                  disabled={updateMutation.isPending}
                >
                  <Save className="w-4 h-4 mr-2" />
                  {updateMutation.isPending ? 'Saving...' : 'Save Changes'}
                </Button>
                <Button
                  type="button"
                  variant="outline"
                  onClick={() => navigate(`/admin/${app}/${model}/`)}
                >
                  Cancel
                </Button>
              </div>
              
              <Button
                type="button"
                variant="destructive"
                onClick={handleDelete}
                disabled={deleteMutation.isPending}
              >
                <Trash2 className="w-4 h-4 mr-2" />
                {deleteMutation.isPending ? 'Deleting...' : 'Delete'}
              </Button>
            </div>
          </form>
        </CardContent>
      </Card>
    </div>
  )
}