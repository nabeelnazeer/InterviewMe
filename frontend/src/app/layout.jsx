'use client';

import ClientEffects from '@/components/ClientEffects';

export default function RootLayout({ children }) {
  return (
    <html lang="en" suppressHydrationWarning>
      <head>
        <meta charSet="utf-8" />
        <meta name="viewport" content="width=device-width, initial-scale=1" />
        <title>Resume Evaluator</title>
        <link rel="icon" href="data:image/x-icon;," />
      </head>
      <body suppressHydrationWarning>
        <ClientEffects />
        <main>{children}</main>
      </body>
    </html>
  );
}
