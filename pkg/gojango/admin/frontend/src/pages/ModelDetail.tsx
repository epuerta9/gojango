import { useParams } from 'react-router-dom'

export function ModelDetail() {
  const { app, model, id } = useParams<{ app: string; model: string; id: string }>()
  
  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-semibold text-gray-900 capitalize">
          {model} Detail
        </h1>
        <p className="mt-1 text-sm text-gray-500">
          View {model} object #{id} in the {app} app
        </p>
      </div>
      
      <div className="admin-card p-6">
        <div className="text-center text-gray-500">
          Model detail implementation coming soon...
        </div>
      </div>
    </div>
  )
}