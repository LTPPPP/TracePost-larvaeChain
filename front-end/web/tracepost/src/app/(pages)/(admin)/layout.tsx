import { Outfit } from 'next/font/google';
import './globals.css';

import { SidebarProvider } from '@/context/SidebarContext';
import { ThemeProvider } from '@/context/ThemeContext';

const outfit = Outfit({
  subsets: ['latin']
});

export default function RootLayout({
  children
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <ThemeProvider>
      <SidebarProvider>
        <div className={`${outfit.className} dark:bg-gray-900`}>{children}</div>
      </SidebarProvider>
    </ThemeProvider>
  );
}
