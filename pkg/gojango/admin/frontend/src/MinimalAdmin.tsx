import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Link } from 'react-router-dom'
import { useQuery } from '@tanstack/react-query'

function MinimalAdmin() {
  // Test API integration
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
  return (
    <div className="min-h-screen bg-background flex">
      {/* Sidebar */}
      <div className="w-64 bg-slate-800 text-white p-6">
        <h1 className="text-xl font-bold mb-8">Gojango Admin</h1>
        <nav className="space-y-2">
          <Link to="/" className="block px-3 py-2 rounded-md bg-slate-700 text-white">
            Dashboard
          </Link>
          <Link to="/users" className="block px-3 py-2 rounded-md text-slate-300 hover:bg-slate-700 hover:text-white transition-colors">
            Users
          </Link>
          <Link to="/posts" className="block px-3 py-2 rounded-md text-slate-300 hover:bg-slate-700 hover:text-white transition-colors">
            Posts
          </Link>
          <Link to="/categories" className="block px-3 py-2 rounded-md text-slate-300 hover:bg-slate-700 hover:text-white transition-colors">
            Categories
          </Link>
        </nav>
      </div>
      
      {/* Main Content */}
      <div className="flex-1 p-8">
        <div className="max-w-4xl mx-auto space-y-6">
        <div>
          <h1 className="text-3xl font-bold text-foreground">
            🎉 Gojango Admin
          </h1>
          <p className="text-muted-foreground mt-2">
            {isLoading ? 'Loading models...' : error ? 'Error loading models' : `Admin interface with ${Object.keys(modelsData?.models || {}).length} models`}
          </p>
        </div>
        
        <Card>
          <CardHeader>
            <CardTitle>Dashboard</CardTitle>
          </CardHeader>
          <CardContent>
            <p className="text-muted-foreground mb-4">
              This is a test of Shadcn UI components working properly.
            </p>
            <div className="flex gap-2">
              <Button>Primary Button</Button>
              <Button variant="outline">Outline Button</Button>
              <Button variant="secondary">Secondary Button</Button>
            </div>
          </CardContent>
        </Card>

        <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
          <Card>
            <CardHeader>
              <CardTitle>Users</CardTitle>
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold">42</div>
              <p className="text-muted-foreground text-sm">Total users</p>
            </CardContent>
          </Card>
          
          <Card>
            <CardHeader>
              <CardTitle>Posts</CardTitle>
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold">128</div>
              <p className="text-muted-foreground text-sm">Published posts</p>
            </CardContent>
          </Card>
          
          <Card>
            <CardHeader>
              <CardTitle>Categories</CardTitle>
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold">12</div>
              <p className="text-muted-foreground text-sm">Active categories</p>
            </CardContent>
          </Card>
        </div>
      </div>
      </div>
    </div>
  )
}

export default MinimalAdmin