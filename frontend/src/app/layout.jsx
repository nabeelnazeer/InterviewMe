'use client';

import ClientEffects from '@/components/ClientEffects';

export default function RootLayout({ children }) {
  return (
    <html lang="en" suppressHydrationWarning>
      <body suppressHydrationWarning>
        <ClientEffects />
        <main>{children}</main>
      </body>
    </html>
  );
}
