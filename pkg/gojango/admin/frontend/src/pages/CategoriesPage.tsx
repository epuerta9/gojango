import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Tag, Plus, Search, Filter } from 'lucide-react'

function CategoriesPage() {
  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold text-foreground">Categories</h1>
          <p className="text-muted-foreground mt-2">
            Organize content with categories and tags
          </p>
        </div>
        <Button>
          <Plus className="w-4 h-4 mr-2" />
          New Category
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
            <Tag className="w-5 h-5 mr-2" />
            Category List
          </CardTitle>
        </CardHeader>
        <CardContent>
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
            {/* Mock category data */}
            {[
              { id: 1, name: "Technology", slug: "technology", posts: 45, color: "blue" },
              { id: 2, name: "Design", slug: "design", posts: 23, color: "purple" },
              { id: 3, name: "Business", slug: "business", posts: 31, color: "green" },
              { id: 4, name: "Programming", slug: "programming", posts: 67, color: "orange" },
              { id: 5, name: "Tutorials", slug: "tutorials", posts: 89, color: "red" },
              { id: 6, name: "News", slug: "news", posts: 12, color: "indigo" },
            ].map((category) => (
              <Card key={category.id} className="hover:shadow-md transition-shadow">
                <CardHeader className="pb-3">
                  <div className="flex items-center justify-between">
                    <div className="flex items-center space-x-2">
                      <div className={`w-3 h-3 rounded-full bg-${category.color}-500`}></div>
                      <CardTitle className="text-base">{category.name}</CardTitle>
                    </div>
                    <Button variant="ghost" size="sm">Edit</Button>
                  </div>
                </CardHeader>
                <CardContent className="pt-0">
                  <div className="space-y-2">
                    <p className="text-sm text-muted-foreground">
                      Slug: /{category.slug}
                    </p>
                    <div className="flex items-center justify-between">
                      <span className="text-sm text-muted-foreground">
                        {category.posts} posts
                      </span>
                      <Button variant="outline" size="sm">
                        View Posts
                      </Button>
                    </div>
                  </div>
                </CardContent>
              </Card>
            ))}
          </div>
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle>Category Statistics</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
            <div className="text-center p-4">
              <div className="text-2xl font-bold">12</div>
              <div className="text-sm text-muted-foreground">Total Categories</div>
            </div>
            <div className="text-center p-4">
              <div className="text-2xl font-bold">267</div>
              <div className="text-sm text-muted-foreground">Total Posts</div>
            </div>
            <div className="text-center p-4">
              <div className="text-2xl font-bold">22.3</div>
              <div className="text-sm text-muted-foreground">Avg Posts/Category</div>
            </div>
          </div>
        </CardContent>
      </Card>
    </div>
  )
}

export default CategoriesPage