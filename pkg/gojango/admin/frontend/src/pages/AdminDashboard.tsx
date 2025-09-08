import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { useQuery } from '@tanstack/react-query'
import { Link } from 'react-router-dom'
import { Database, Users, FileText, Tag } from 'lucide-react'

function AdminDashboard() {
  const { data: modelsData, isLoading, error } = useQuery({
    queryKey: ['models'],
    queryFn: async () => {
      const response = await fetch('/admin/api/models/')
      if (!response.ok) {
        throw new Error('Failed to fetch models')
      }
      return response.json()
    },
  })

  if (isLoading) {
    return (
      <div className="space-y-6">
        <div>
          <h1 className="text-3xl font-bold text-foreground">Dashboard</h1>
          <p className="text-muted-foreground mt-2">Loading...</p>
        </div>
      </div>
    )
  }

  if (error) {
    return (
      <div className="space-y-6">
        <div>
          <h1 className="text-3xl font-bold text-foreground">Dashboard</h1>
          <p className="text-muted-foreground mt-2">Error loading data</p>
        </div>
      </div>
    )
  }

  const modelCount = Object.keys(modelsData?.models || {}).length

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-3xl font-bold text-foreground">
          🎉 Gojango Admin Dashboard
        </h1>
        <p className="text-muted-foreground mt-2">
          {modelsData?.site?.headerTitle || 'Managing'} - {modelCount} models
        </p>
      </div>
      
      <Card>
        <CardHeader>
          <CardTitle>Quick Actions</CardTitle>
        </CardHeader>
        <CardContent>
          <p className="text-muted-foreground mb-4">
            Common admin tasks and shortcuts
          </p>
          <div className="flex gap-2">
            <Button asChild>
              <Link to="/users">
                <Users className="w-4 h-4 mr-2" />
                Manage Users
              </Link>
            </Button>
            <Button variant="outline" asChild>
              <Link to="/posts">
                <FileText className="w-4 h-4 mr-2" />
                View Posts
              </Link>
            </Button>
            <Button variant="secondary" asChild>
              <Link to="/categories">
                <Tag className="w-4 h-4 mr-2" />
                Categories
              </Link>
            </Button>
          </div>
        </CardContent>
      </Card>

      <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Users</CardTitle>
            <Users className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">42</div>
            <p className="text-muted-foreground text-sm">Total users</p>
            <Button variant="outline" size="sm" className="mt-2" asChild>
              <Link to="/users">View all</Link>
            </Button>
          </CardContent>
        </Card>
        
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Posts</CardTitle>
            <FileText className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">128</div>
            <p className="text-muted-foreground text-sm">Published posts</p>
            <Button variant="outline" size="sm" className="mt-2" asChild>
              <Link to="/posts">View all</Link>
            </Button>
          </CardContent>
        </Card>
        
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Categories</CardTitle>
            <Tag className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">12</div>
            <p className="text-muted-foreground text-sm">Active categories</p>
            <Button variant="outline" size="sm" className="mt-2" asChild>
              <Link to="/categories">View all</Link>
            </Button>
          </CardContent>
        </Card>
      </div>

      {/* Models Overview */}
      <Card>
        <CardHeader>
          <CardTitle>Registered Models</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
            {Object.entries(modelsData?.models || {}).map(([key, model]: [string, any]) => (
              <div key={key} className="flex items-center space-x-3 p-3 border rounded-lg">
                <Database className="h-5 w-5 text-muted-foreground" />
                <div className="flex-1">
                  <h3 className="font-medium">{model.verboseNamePlural}</h3>
                  <p className="text-sm text-muted-foreground">{model.app}.{model.name}</p>
                </div>
                <Button variant="ghost" size="sm" asChild>
                  <Link to={`/${model.app}/${model.name}`}>View</Link>
                </Button>
              </div>
            ))}
          </div>
        </CardContent>
      </Card>
    </div>
  )
}

export default AdminDashboard