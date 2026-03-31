// Shared test utilities and wrappers

import React from 'react';
import { render, type RenderOptions } from '@testing-library/react';
import { MemoryRouter, type MemoryRouterProps } from 'react-router-dom';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { ToastProvider } from '../contexts/ToastContext';
import { NamespaceProvider } from '../contexts/NamespaceContext';

interface WrapperOptions {
  initialEntries?: MemoryRouterProps['initialEntries'];
}

function createTestQueryClient() {
  return new QueryClient({
    defaultOptions: {
      queries: {
        retry: false,
        gcTime: 0,
      },
      mutations: {
        retry: false,
      },
    },
  });
}

function createWrapper(options: WrapperOptions = {}) {
  const queryClient = createTestQueryClient();

  return function Wrapper({ children }: { children: React.ReactNode }) {
    return (
      <QueryClientProvider client={queryClient}>
        <ToastProvider>
          <NamespaceProvider>
            <MemoryRouter initialEntries={options.initialEntries || ['/']}>
              {children}
            </MemoryRouter>
          </NamespaceProvider>
        </ToastProvider>
      </QueryClientProvider>
    );
  };
}

/**
 * Custom render function that wraps components with all required providers.
 */
export function renderWithProviders(
  ui: React.ReactElement,
  options?: WrapperOptions & Omit<RenderOptions, 'wrapper'>
) {
  const { initialEntries, ...renderOptions } = options || {};
  return render(ui, {
    wrapper: createWrapper({ initialEntries }),
    ...renderOptions,
  });
}

export { createTestQueryClient };
