import { Link } from 'react-router-dom'
import { HomeIcon } from '@heroicons/react/24/outline'

export function NotFound() {
  return (
    <div className="min-h-full flex flex-col justify-center py-12">
      <div className="mx-auto max-w-md">
        <div className="text-center">
          <div className="text-6xl font-bold text-primary-600">404</div>
          <h1 className="mt-4 text-xl font-semibold text-gray-900">
            Page not found
          </h1>
          <p className="mt-2 text-sm text-gray-600">
            Sorry, we couldn't find the page you're looking for.
          </p>
          <div className="mt-6">
            <Link
              to="/admin/"
              className="inline-flex items-center px-4 py-2 border border-transparent text-sm font-medium rounded-md text-white bg-primary-600 hover:bg-primary-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-primary-500"
            >
              <HomeIcon className="mr-2 h-4 w-4" />
              Go back home
            </Link>
          </div>
        </div>
      </div>
    </div>
  )
}