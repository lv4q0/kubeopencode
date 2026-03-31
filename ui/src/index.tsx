import React from 'react';
import { createRoot } from 'react-dom/client';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { ToastProvider } from './contexts/ToastContext';
import App from './App';
import './index.css';

const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      refetchOnWindowFocus: true,
      refetchOnMount: true,
      retry: 1,
      staleTime: 5000, // Data becomes stale after 5 seconds
    },
  },
});

async function startApp() {
  // Start MSW browser mock when MOCK_API is enabled (set via webpack DefinePlugin)
  if (typeof MOCK_API !== 'undefined' && MOCK_API) {
    const { worker } = await import('./mocks/browser');
    await worker.start({ onUnhandledRequest: 'bypass' });
    console.log('[MSW] Mock API enabled - running without backend');
  }

  const container = document.getElementById('root');
  if (!container) throw new Error('Root container not found');

  const root = createRoot(container);
  root.render(
    <React.StrictMode>
      <QueryClientProvider client={queryClient}>
        <ToastProvider>
          <App />
        </ToastProvider>
      </QueryClientProvider>
    </React.StrictMode>
  );
}

startApp();
