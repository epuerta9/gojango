import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Users, Plus, Search, Filter, ChevronLeft, ChevronRight } from 'lucide-react'
import { useQuery } from '@tanstack/react-query'
import { adminClient } from '@/services/client'
import { useState } from 'react'

function UsersPage() {
  const [page, setPage] = useState(1)
  const [search] = useState('')
  const pageSize = 25

  const { data, isLoading, error } = useQuery({
    queryKey: ['objects', 'main', 'user', page, search],
    queryFn: () => adminClient.listObjects({
      app: 'main',
      model: 'user',
      page,
      pageSize,
      search,
    }),
  })

  if (error) {
    return (
      <div className="space-y-6">
        <div>
          <h1 className="text-3xl font-bold text-foreground">Users</h1>
          <p className="text-muted-foreground mt-2">Error loading users</p>
        </div>
      </div>
    )
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold text-foreground">Users</h1>
          <p className="text-muted-foreground mt-2">
            Manage user accounts and permissions
          </p>
        </div>
        <Button>
          <Plus className="w-4 h-4 mr-2" />
          Add User
        </Button>
      </div>

      <div className="flex gap-2">
        <Button variant="outline" size="sm">
          <Search className="w-4 h-4 mr-2" />
          Search
        </Button>
        <Button variant="outline" size="sm">
          <Filter className="w-4 h-4 mr-2" />
          Filter
        </Button>
      </div>

      <Card>
        <CardHeader>
          <CardTitle className="flex items-center">
            <Users className="w-5 h-5 mr-2" />
            User List ({data?.totalCount || 0} total)
          </CardTitle>
        </CardHeader>
        <CardContent>
          {isLoading ? (
            <div className="flex items-center justify-center h-32">
              <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary"></div>
            </div>
          ) : (
            <div className="space-y-4">
              {data?.objects.map((user) => (
                <div key={user.id} className="flex items-center justify-between p-4 border rounded-lg">
                  <div className="flex items-center space-x-4">
                    <div className="w-10 h-10 bg-primary/10 rounded-full flex items-center justify-center">
                      <Users className="w-5 h-5 text-primary" />
                    </div>
                    <div>
                      <h3 className="font-medium">{String(user.fields.username || '')}</h3>
                      <p className="text-sm text-muted-foreground">{String(user.fields.email || '')}</p>
                      <p className="text-sm text-muted-foreground">
                        {String(user.fields.firstName || '')} {String(user.fields.lastName || '')}
                      </p>
                    </div>
                  </div>
                  <div className="flex items-center space-x-2">
                    <span className="text-sm text-muted-foreground">
                      {user.fields.isStaff ? 'Staff' : 'User'}
                    </span>
                    <span className={`px-2 py-1 text-xs rounded-full ${
                      user.fields.isActive 
                        ? 'bg-green-100 text-green-800' 
                        : 'bg-gray-100 text-gray-800'
                    }`}>
                      {user.fields.isActive ? 'Active' : 'Inactive'}
                    </span>
                    <Button variant="ghost" size="sm">Edit</Button>
                  </div>
                </div>
              ))}
              
              {/* Pagination */}
              {data && data.totalPages > 1 && (
                <div className="flex items-center justify-between mt-6">
                  <div className="text-sm text-muted-foreground">
                    Page {data.page} of {data.totalPages} ({data.totalCount} total items)
                  </div>
                  <div className="flex items-center space-x-2">
                    <Button
                      variant="outline"
                      size="sm"
                      onClick={() => setPage(p => p - 1)}
                      disabled={!data.hasPrevious}
                    >
                      <ChevronLeft className="w-4 h-4" />
                      Previous
                    </Button>
                    <Button
                      variant="outline"
                      size="sm"
                      onClick={() => setPage(p => p + 1)}
                      disabled={!data.hasNext}
                    >
                      Next
                      <ChevronRight className="w-4 h-4" />
                    </Button>
                  </div>
                </div>
              )}
            </div>
          )}
        </CardContent>
      </Card>
    </div>
  )
}

export default UsersPage