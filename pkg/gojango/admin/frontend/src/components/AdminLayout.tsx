import { Link, useLocation } from 'react-router-dom'
import { Home, Users, FileText, Tag } from 'lucide-react'
import { cn } from '@/utils/cn'

interface AdminLayoutProps {
  children: React.ReactNode
}

const navigation = [
  { name: 'Dashboard', href: '/', icon: Home },
  { name: 'Users', href: '/users', icon: Users },
  { name: 'Posts', href: '/posts', icon: FileText },
  { name: 'Categories', href: '/categories', icon: Tag },
]

export function AdminLayout({ children }: AdminLayoutProps) {
  const location = useLocation()

  const isActive = (href: string) => {
    if (href === '/') {
      return location.pathname === '/'
    }
    return location.pathname.startsWith(href)
  }

  return (
    <div className="min-h-screen bg-background flex">
      {/* Sidebar */}
      <div className="w-64 bg-slate-800 text-white">
        <div className="p-6">
          <h1 className="text-xl font-bold mb-8">Gojango Admin</h1>
          <nav className="space-y-2">
            {navigation.map((item) => {
              const Icon = item.icon
              const active = isActive(item.href)
              
              return (
                <Link
                  key={item.name}
                  to={item.href}
                  className={cn(
                    'flex items-center px-3 py-2 rounded-md text-sm font-medium transition-colors',
                    active 
                      ? 'bg-slate-700 text-white' 
                      : 'text-slate-300 hover:bg-slate-700 hover:text-white'
                  )}
                >
                  <Icon className="w-5 h-5 mr-3" />
                  {item.name}
                </Link>
              )
            })}
          </nav>
        </div>
        
        {/* Footer */}
        <div className="absolute bottom-0 left-0 w-64 p-4 border-t border-slate-700">
          <div className="text-sm text-slate-400">
            Gojango v0.2.0
          </div>
        </div>
      </div>
      
      {/* Main Content */}
      <div className="flex-1">
        <main className="p-8">
          <div className="max-w-7xl mx-auto">
            {children}
          </div>
        </main>
      </div>
    </div>
  )
}