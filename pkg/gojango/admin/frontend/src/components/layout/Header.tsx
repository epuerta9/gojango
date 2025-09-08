import React from 'react'
import { Bars3Icon, BellIcon, UserCircleIcon } from '@heroicons/react/24/outline'
import { Menu, Transition } from '@headlessui/react'
import { cn } from '@/utils/cn'

interface HeaderProps {
  onMenuClick: () => void
}

export function Header({ onMenuClick }: HeaderProps) {
  const userNavigation = [
    { name: 'Your Profile', href: '#' },
    { name: 'Settings', href: '#' },
    { name: 'Sign out', href: '#' },
  ]

  return (
    <div className="sticky top-0 z-10 flex h-16 flex-shrink-0 bg-white shadow-sm border-b border-admin-border">
      {/* Mobile menu button */}
      <button
        type="button"
        className="border-r border-admin-border px-4 text-gray-500 focus:outline-none focus:ring-2 focus:ring-inset focus:ring-primary-500 lg:hidden"
        onClick={onMenuClick}
      >
        <Bars3Icon className="h-6 w-6" />
      </button>

      <div className="flex flex-1 justify-between px-4 lg:px-6">
        {/* Left side - breadcrumb or title would go here */}
        <div className="flex flex-1 items-center">
          {/* Desktop spacing to account for sidebar */}
          <div className="hidden lg:block lg:w-64" />
          
          {/* Breadcrumb or page title */}
          <div className="ml-4">
            {/* This could be a breadcrumb component */}
          </div>
        </div>

        {/* Right side - notifications and user menu */}
        <div className="ml-4 flex items-center space-x-4">
          {/* Notifications */}
          <button
            type="button"
            className="text-gray-400 hover:text-gray-500 focus:outline-none focus:ring-2 focus:ring-primary-500 focus:ring-offset-2"
          >
            <BellIcon className="h-6 w-6" />
          </button>

          {/* Profile dropdown */}
          <Menu as="div" className="relative">
            <div>
              <Menu.Button className="flex max-w-xs items-center rounded-full text-gray-400 hover:text-gray-500 focus:outline-none focus:ring-2 focus:ring-primary-500 focus:ring-offset-2">
                <UserCircleIcon className="h-8 w-8" />
              </Menu.Button>
            </div>
            <Transition
              as={React.Fragment}
              enter="transition ease-out duration-100"
              enterFrom="transform opacity-0 scale-95"
              enterTo="transform opacity-100 scale-100"
              leave="transition ease-in duration-75"
              leaveFrom="transform opacity-100 scale-100"
              leaveTo="transform opacity-0 scale-95"
            >
              <Menu.Items className="absolute right-0 z-10 mt-2 w-48 origin-top-right rounded-md bg-white py-1 shadow-lg ring-1 ring-black ring-opacity-5 focus:outline-none">
                {userNavigation.map((item) => (
                  <Menu.Item key={item.name}>
                    {({ active }) => (
                      <a
                        href={item.href}
                        className={cn(
                          active ? 'bg-gray-100' : '',
                          'block px-4 py-2 text-sm text-gray-700'
                        )}
                      >
                        {item.name}
                      </a>
                    )}
                  </Menu.Item>
                ))}
              </Menu.Items>
            </Transition>
          </Menu>
        </div>
      </div>
    </div>
  )
}