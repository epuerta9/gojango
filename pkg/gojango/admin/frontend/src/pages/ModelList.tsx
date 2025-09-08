import React from 'react'
import { useParams, Link } from 'react-router-dom'
import { Plus, Search, Filter, Eye, Edit } from 'lucide-react'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Badge } from '@/components/ui/badge'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'

export function ModelList() {
  const { app, model } = useParams<{ app: string; model: string }>()
  
  // Mock data for demonstration
  const mockData = React.useMemo(() => {
    if (model === 'post') {
      return [
        { id: 1, title: 'Getting Started with Gojango', status: 'published', author_id: 1, created_at: '2025-01-15T10:30:00Z' },
        { id: 2, title: 'Django-style Admin Interface', status: 'draft', author_id: 2, created_at: '2025-01-14T15:45:00Z' },
        { id: 3, title: 'Building Modern Web Apps', status: 'published', author_id: 1, created_at: '2025-01-13T09:20:00Z' },
      ]
    } else if (model === 'user') {
      return [
        { id: 1, username: 'admin', email: 'admin@example.com', is_active: true, is_staff: true, created_at: '2025-01-10T12:00:00Z' },
        { id: 2, username: 'john_doe', email: 'john@example.com', is_active: true, is_staff: false, created_at: '2025-01-11T14:30:00Z' },
        { id: 3, username: 'jane_smith', email: 'jane@example.com', is_active: false, is_staff: false, created_at: '2025-01-12T16:15:00Z' },
      ]
    } else if (model === 'category') {
      return [
        { id: 1, name: 'Technology', slug: 'technology', created_at: '2025-01-08T10:00:00Z' },
        { id: 2, name: 'Programming', slug: 'programming', created_at: '2025-01-09T11:30:00Z' },
        { id: 3, name: 'Web Development', slug: 'web-development', created_at: '2025-01-10T13:45:00Z' },
      ]
    }
    return []
  }, [model])

  const getStatusBadge = (status: string) => {
    const variants: Record<string, any> = {
      published: 'default',
      draft: 'secondary',
      archived: 'outline',
    }
    return <Badge variant={variants[status] || 'secondary'}>{status}</Badge>
  }

  const getBooleanBadge = (value: boolean, trueLabel: string, falseLabel: string) => {
    return <Badge variant={value ? 'default' : 'secondary'}>{value ? trueLabel : falseLabel}</Badge>
  }

  const formatDate = (dateString: string) => {
    return new Date(dateString).toLocaleDateString('en-US', {
      year: 'numeric',
      month: 'short',
      day: 'numeric',
    })
  }

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold tracking-tight text-slate-900 capitalize">
            {model}s
          </h1>
          <p className="mt-2 text-slate-600">
            Manage {model} objects in the {app} app
          </p>
        </div>
        <Button asChild>
          <Link to={`/admin/${app}/${model}/add/`}>
            <Plus className="mr-2 h-4 w-4" />
            Add {model}
          </Link>
        </Button>
      </div>

      {/* Filters and Search */}
      <Card>
        <CardHeader>
          <div className="flex items-center space-x-4">
            <div className="flex-1">
              <div className="relative">
                <Search className="absolute left-3 top-3 h-4 w-4 text-slate-400" />
                <Input
                  placeholder={`Search ${model}s...`}
                  className="pl-9"
                />
              </div>
            </div>
            <Button variant="outline">
              <Filter className="mr-2 h-4 w-4" />
              Filters
            </Button>
          </div>
        </CardHeader>
      </Card>

      {/* Data Table */}
      <Card>
        <CardHeader>
          <CardTitle>{mockData.length} {model}s</CardTitle>
          <CardDescription>
            A list of all {model} objects with their current status and details.
          </CardDescription>
        </CardHeader>
        <CardContent>
          {mockData.length === 0 ? (
            <div className="text-center py-12">
              <div className="text-slate-400 text-lg mb-4">No {model}s found</div>
              <Button asChild>
                <Link to={`/admin/${app}/${model}/add/`}>
                  <Plus className="mr-2 h-4 w-4" />
                  Create your first {model}
                </Link>
              </Button>
            </div>
          ) : (
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>ID</TableHead>
                  {model === 'post' && (
                    <>
                      <TableHead>Title</TableHead>
                      <TableHead>Status</TableHead>
                      <TableHead>Author ID</TableHead>
                      <TableHead>Created</TableHead>
                    </>
                  )}
                  {model === 'user' && (
                    <>
                      <TableHead>Username</TableHead>
                      <TableHead>Email</TableHead>
                      <TableHead>Active</TableHead>
                      <TableHead>Staff</TableHead>
                      <TableHead>Created</TableHead>
                    </>
                  )}
                  {model === 'category' && (
                    <>
                      <TableHead>Name</TableHead>
                      <TableHead>Slug</TableHead>
                      <TableHead>Created</TableHead>
                    </>
                  )}
                  <TableHead className="w-[70px]">Actions</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {mockData.map((item: any) => (
                  <TableRow key={item.id}>
                    <TableCell className="font-medium">{item.id}</TableCell>
                    {model === 'post' && (
                      <>
                        <TableCell>{item.title}</TableCell>
                        <TableCell>{getStatusBadge(item.status)}</TableCell>
                        <TableCell>{item.author_id}</TableCell>
                        <TableCell>{formatDate(item.created_at)}</TableCell>
                      </>
                    )}
                    {model === 'user' && (
                      <>
                        <TableCell>{item.username}</TableCell>
                        <TableCell>{item.email}</TableCell>
                        <TableCell>{getBooleanBadge(item.is_active, 'Active', 'Inactive')}</TableCell>
                        <TableCell>{getBooleanBadge(item.is_staff, 'Staff', 'User')}</TableCell>
                        <TableCell>{formatDate(item.created_at)}</TableCell>
                      </>
                    )}
                    {model === 'category' && (
                      <>
                        <TableCell>{item.name}</TableCell>
                        <TableCell>{item.slug}</TableCell>
                        <TableCell>{formatDate(item.created_at)}</TableCell>
                      </>
                    )}
                    <TableCell>
                      <div className="flex items-center space-x-2">
                        <Button variant="ghost" size="sm" asChild>
                          <Link to={`/admin/${app}/${model}/${item.id}/`}>
                            <Eye className="h-4 w-4" />
                          </Link>
                        </Button>
                        <Button variant="ghost" size="sm" asChild>
                          <Link to={`/admin/${app}/${model}/${item.id}/change/`}>
                            <Edit className="h-4 w-4" />
                          </Link>
                        </Button>
                      </div>
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          )}
        </CardContent>
      </Card>

      {/* Pagination */}
      <div className="flex items-center justify-between">
        <div className="text-sm text-slate-600">
          Showing 1 to {mockData.length} of {mockData.length} results
        </div>
        <div className="flex space-x-2">
          <Button variant="outline" size="sm" disabled>
            Previous
          </Button>
          <Button variant="outline" size="sm" disabled>
            Next
          </Button>
        </div>
      </div>
    </div>
  )
}