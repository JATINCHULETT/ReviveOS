import './globals.css';
import AppLayout from '@/components/layout/AppLayout';

export const metadata = {
  title: 'ReviveOS — Autonomous AI Payment Recovery & Recurring Autopay Engine',
  description: 'Diagnose, orchestrate, and recover failed recurring payments and checkout declines with DeepSeek-R1 intelligence and zero-touch retry orchestration.',
};

export default function RootLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <html lang="en" suppressHydrationWarning className="dark">
      <head>
        <link rel="preconnect" href="https://fonts.googleapis.com" />
        <link rel="preconnect" href="https://fonts.gstatic.com" crossOrigin="anonymous" />
      </head>
      <body suppressHydrationWarning>
        <AppLayout>{children}</AppLayout>
      </body>
    </html>
  );
}
