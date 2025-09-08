import { useParams, useNavigate } from 'react-router-dom'
import { useState } from 'react'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { adminClient } from '@/services/client'
import { ArrowLeft, Save } from 'lucide-react'

export function ModelCreate() {
  const { app, model } = useParams<{ app: string; model: string }>()
  const navigate = useNavigate()
  const queryClient = useQueryClient()
  const [formData, setFormData] = useState<Record<string, any>>({})
  const [errors, setErrors] = useState<Record<string, string>>({})

  // Get model schema for form fields
  const { data: schemaData, isLoading } = useQuery({
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

  // Create mutation
  const createMutation = useMutation({
    mutationFn: async (data: Record<string, any>) => {
      return adminClient.createObject(app!, model!, data)
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['objects', app, model] })
      navigate(`/admin/${app}/${model}/`)
    },
    onError: (error: any) => {
      console.error('Failed to create object:', error)
      setErrors({ general: 'Failed to create object. Please try again.' })
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

    createMutation.mutate(formData)
  }

  const renderField = (field: any) => {
    const fieldName = field.name
    const fieldValue = formData[fieldName] || ''
    
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
                checked={fieldValue}
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
            Add {model}
          </h1>
          <p className="text-muted-foreground mt-2">
            Create a new {model} object in the {app} app
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
            
            <div className="flex items-center space-x-2 pt-6 border-t">
              <Button
                type="submit"
                disabled={createMutation.isPending}
              >
                <Save className="w-4 h-4 mr-2" />
                {createMutation.isPending ? 'Creating...' : 'Create'}
              </Button>
              <Button
                type="button"
                variant="outline"
                onClick={() => navigate(`/admin/${app}/${model}/`)}
              >
                Cancel
              </Button>
            </div>
          </form>
        </CardContent>
      </Card>
    </div>
  )
}