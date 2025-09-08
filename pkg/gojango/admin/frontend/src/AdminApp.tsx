import { Routes, Route } from 'react-router-dom'
import { AdminLayout } from '@/components/AdminLayout'
import AdminDashboard from '@/pages/AdminDashboard'
import UsersPage from '@/pages/UsersPage'
import PostsPage from '@/pages/PostsPage'
import CategoriesPage from '@/pages/CategoriesPage'

function AdminApp() {
  return (
    <AdminLayout>
      <Routes>
        <Route path="/" element={<AdminDashboard />} />
        <Route path="/users" element={<UsersPage />} />
        <Route path="/posts" element={<PostsPage />} />
        <Route path="/categories" element={<CategoriesPage />} />
        <Route path="*" element={
          <div className="text-center py-12">
            <h2 className="text-2xl font-bold text-foreground mb-2">Page Not Found</h2>
            <p className="text-muted-foreground">The page you're looking for doesn't exist.</p>
          </div>
        } />
      </Routes>
    </AdminLayout>
  )
}

export default AdminApp