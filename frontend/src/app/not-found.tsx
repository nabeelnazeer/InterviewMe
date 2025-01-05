import Link from 'next/link'
 
export default function NotFound() {
  return (
    <div className="min-h-screen bg-gray-900 text-white flex items-center justify-center">
      <div className="text-center">
        <h2 className="text-3xl font-bold mb-4">Page Not Found</h2>
        <Link 
          href="/"
          className="px-6 py-3 bg-gradient-to-r from-blue-500 to-purple-500 rounded-full 
                     text-white font-semibold hover:scale-105 transform transition-all"
        >
          Return Home
        </Link>
      </div>
    </div>
  )
}
