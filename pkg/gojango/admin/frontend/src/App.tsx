import { Routes, Route } from 'react-router-dom'
import { AdminLayout } from '@/components/layout/AdminLayout'
import AdminDashboard from '@/pages/AdminDashboard'
import { ModelList } from '@/pages/ModelList'
import { ModelDetail } from '@/pages/ModelDetail'
import { ModelCreate } from '@/pages/ModelCreate'
import { ModelEdit } from '@/pages/ModelEdit'
import { NotFound } from '@/pages/NotFound'

function App() {
  return (
    <AdminLayout>
      <Routes>
        <Route path="/" element={<AdminDashboard />} />
        <Route path="/:app/:model/" element={<ModelList />} />
        <Route path="/:app/:model/add/" element={<ModelCreate />} />
        <Route path="/:app/:model/:id/" element={<ModelDetail />} />
        <Route path="/:app/:model/:id/change/" element={<ModelEdit />} />
        <Route path="*" element={<NotFound />} />
      </Routes>
    </AdminLayout>
  )
}

export default App