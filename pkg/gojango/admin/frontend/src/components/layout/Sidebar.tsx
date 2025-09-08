import React from 'react'
import { Link, useLocation } from 'react-router-dom'
import { 
  HomeIcon, 
  CubeIcon,
  XMarkIcon,
  ChevronRightIcon 
} from '@heroicons/react/24/outline'
import { useModels } from '@/hooks/useModels'
import { cn } from '@/utils/cn'

interface SidebarProps {
  isOpen: boolean
  onClose: () => void
}

export function Sidebar({ isOpen, onClose }: SidebarProps) {
  const location = useLocation()
  const { data: modelsResponse, isLoading, error } = useModels()
  
  const navigation = [
    { name: 'Dashboard', href: '/admin/', icon: HomeIcon },
  ]
  
  // Group models by app
  const modelsByApp = React.useMemo(() => {
    if (!modelsResponse?.models) return {}
    
    const grouped: Record<string, Array<{ name: string; app: string; verboseName: string }>> = {}
    
    Object.entries(modelsResponse.models).forEach(([, model]) => {
      if (!grouped[model.app]) {
        grouped[model.app] = []
      }
      grouped[model.app].push({
        name: model.name,
        app: model.app,
        verboseName: model.verboseName || model.name,
      })
    })
    
    return grouped
  }, [modelsResponse])

  const isCurrentPage = (href: string) => {
    return location.pathname === href || location.pathname.startsWith(href.replace(/\/$/, ''))
  }

  return (
    <>
      {/* Mobile sidebar */}
      <div className={cn(
        'fixed inset-y-0 left-0 z-50 w-64 bg-slate-800 transform lg:hidden',
        isOpen ? 'translate-x-0' : '-translate-x-full',
        'transition-transform duration-300 ease-in-out'
      )}>
        <SidebarContent 
          navigation={navigation}
          modelsByApp={modelsByApp}
          isCurrentPage={isCurrentPage}
          showCloseButton={true}
          onClose={onClose}
          isLoading={isLoading}
          error={error}
        />
      </div>

      {/* Desktop sidebar */}
      <div className="hidden lg:fixed lg:inset-y-0 lg:flex lg:w-64 lg:flex-col">
        <div className="flex min-h-0 flex-1 flex-col bg-slate-800">
          <SidebarContent 
            navigation={navigation}
            modelsByApp={modelsByApp}
            isCurrentPage={isCurrentPage}
            showCloseButton={false}
            onClose={onClose}
            isLoading={isLoading}
            error={error}
          />
        </div>
      </div>
    </>
  )
}

interface SidebarContentProps {
  navigation: Array<{ name: string; href: string; icon: React.ComponentType<any> }>
  modelsByApp: Record<string, Array<{ name: string; app: string; verboseName: string }>>
  isCurrentPage: (href: string) => boolean
  showCloseButton: boolean
  onClose: () => void
  isLoading: boolean
  error: any
}

function SidebarContent({ 
  navigation, 
  modelsByApp, 
  isCurrentPage, 
  showCloseButton, 
  onClose, 
  isLoading, 
  error 
}: SidebarContentProps) {
  const [expandedApps, setExpandedApps] = React.useState<Record<string, boolean>>({})
  
  const toggleApp = (appName: string) => {
    setExpandedApps(prev => ({
      ...prev,
      [appName]: !prev[appName],
    }))
  }

  return (
    <>
      {/* Header */}
      <div className="flex h-16 flex-shrink-0 items-center justify-between px-4 bg-gray-900">
        <h1 className="text-lg font-semibold text-white">Gojango Admin</h1>
        {showCloseButton && (
          <button
            type="button"
            className="text-gray-400 hover:text-white"
            onClick={onClose}
          >
            <XMarkIcon className="h-6 w-6" />
          </button>
        )}
      </div>

      {/* Navigation */}
      <div className="flex flex-1 flex-col overflow-y-auto">
        <nav className="space-y-1 px-2 py-3">
          {/* Main navigation */}
          {navigation.map((item) => {
            const Icon = item.icon
            const current = isCurrentPage(item.href)
            
            return (
              <Link
                key={item.name}
                to={item.href}
                className={cn(
                  'flex items-center rounded-md px-2 py-2 text-sm font-medium transition-colors',
                  current ? 'bg-accent text-accent-foreground' : 'text-muted-foreground hover:bg-accent hover:text-accent-foreground'
                )}
                onClick={() => showCloseButton && onClose()}
              >
                <Icon className="mr-3 h-5 w-5" />
                {item.name}
              </Link>
            )
          })}

          {/* Models section */}
          <div className="mt-8">
            <h3 className="px-2 text-xs font-semibold text-gray-400 uppercase tracking-wider">
              Models
            </h3>
            
            {isLoading && (
              <div className="mt-2 px-2">
                <div className="text-sm text-gray-500">Loading models...</div>
              </div>
            )}
            
            {error && (
              <div className="mt-2 px-2">
                <div className="text-sm text-red-400">Failed to load models</div>
              </div>
            )}
            
            {Object.entries(modelsByApp).map(([appName, models]) => (
              <div key={appName} className="mt-3">
                <button
                  type="button"
                  className="flex w-full items-center px-2 py-1 text-sm text-gray-400 hover:text-white"
                  onClick={() => toggleApp(appName)}
                >
                  <ChevronRightIcon 
                    className={cn(
                      'mr-2 h-4 w-4 transition-transform',
                      expandedApps[appName] ? 'rotate-90' : ''
                    )} 
                  />
                  <span className="capitalize">{appName}</span>
                </button>
                
                {expandedApps[appName] && (
                  <div className="ml-6 mt-1 space-y-1">
                    {models.map((model) => {
                      const href = `/admin/${model.app}/${model.name}/`
                      const current = isCurrentPage(href)
                      
                      return (
                        <Link
                          key={`${model.app}.${model.name}`}
                          to={href}
                          className={cn(
                            'flex items-center rounded-md px-2 py-2 text-sm font-medium transition-colors',
                            current ? 'bg-accent text-accent-foreground' : 'text-muted-foreground hover:bg-accent hover:text-accent-foreground'
                          )}
                          onClick={() => showCloseButton && onClose()}
                        >
                          <CubeIcon className="mr-3 h-4 w-4" />
                          {model.verboseName}
                        </Link>
                      )
                    })}
                  </div>
                )}
              </div>
            ))}
          </div>
        </nav>
      </div>

      {/* Footer */}
      <div className="flex flex-shrink-0 bg-gray-800 p-4">
        <div className="text-sm text-gray-400">
          Gojango v0.2.0
        </div>
      </div>
    </>
  )
}