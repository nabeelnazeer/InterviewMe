import './globals.css'

export const metadata = {
  title: 'InterviewMe',
  description: 'AI-powered resume analysis',
}

export default function RootLayout({ children }) {
  return (
    <html lang="en">
      <body>
        {children}
      </body>
    </html>
  )
}
