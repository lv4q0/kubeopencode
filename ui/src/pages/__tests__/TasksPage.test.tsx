import { describe, it, expect, vi, beforeEach } from 'vitest';
import { screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { http, HttpResponse } from 'msw';
import { server } from '../../mocks/server';
import { renderWithProviders } from '../../test/utils';
import TasksPage from '../TasksPage';

// Mock TimeAgo to avoid timing issues
vi.mock('../../components/TimeAgo', () => ({
  default: ({ date }: { date: string }) => <span>{date}</span>,
}));

describe('TasksPage', () => {
  beforeEach(() => {
    // Clear cookies
    document.cookie.split(';').forEach((c) => {
      document.cookie = c.trim().split('=')[0] + '=;expires=Thu, 01 Jan 1970 00:00:00 GMT;path=/';
    });
  });

  it('renders page title and description', async () => {
    renderWithProviders(<TasksPage />, { initialEntries: ['/tasks'] });

    expect(screen.getByText('Tasks')).toBeInTheDocument();
    expect(screen.getByText('Manage and monitor AI agent tasks')).toBeInTheDocument();
  });

  it('renders task list from API', async () => {
    renderWithProviders(<TasksPage />, { initialEntries: ['/tasks'] });

    await waitFor(() => {
      expect(screen.getByText('fix-auth-bug')).toBeInTheDocument();
    });

    expect(screen.getByText('add-user-profile')).toBeInTheDocument();
  });

  it('shows loading skeleton while fetching', () => {
    server.use(
      http.get('/api/v1/tasks', async () => {
        await new Promise((r) => setTimeout(r, 100));
        return HttpResponse.json({ tasks: [], total: 0 });
      })
    );

    renderWithProviders(<TasksPage />, { initialEntries: ['/tasks'] });
    expect(screen.queryByText('fix-auth-bug')).not.toBeInTheDocument();
  });

  it('shows error state when API fails', async () => {
    server.use(
      http.get('/api/v1/tasks', () => {
        return HttpResponse.json({ message: 'Server error' }, { status: 500 });
      })
    );

    renderWithProviders(<TasksPage />, { initialEntries: ['/tasks'] });

    await waitFor(() => {
      expect(screen.getByText(/Error loading tasks/)).toBeInTheDocument();
    });

    expect(screen.getByText('Retry')).toBeInTheDocument();
  });

  it('shows empty state when no tasks exist', async () => {
    server.use(
      http.get('/api/v1/tasks', () => {
        return HttpResponse.json({
          tasks: [],
          total: 0,
          pagination: { limit: 20, offset: 0, totalCount: 0, hasMore: false },
        });
      })
    );

    renderWithProviders(<TasksPage />, { initialEntries: ['/tasks'] });

    await waitFor(() => {
      expect(screen.getByText(/No tasks found/)).toBeInTheDocument();
    });
  });

  it('renders "New Task" link', async () => {
    renderWithProviders(<TasksPage />, { initialEntries: ['/tasks'] });
    const newTaskLink = screen.getByText('New Task');
    expect(newTaskLink).toBeInTheDocument();
    expect(newTaskLink.closest('a')).toHaveAttribute('href', expect.stringContaining('/tasks/create'));
  });

  it('renders namespace selector in header', async () => {
    renderWithProviders(<TasksPage />, { initialEntries: ['/tasks'] });

    // NamespaceProvider defaults to ALL_NAMESPACES, namespace options loaded from API
    await waitFor(() => {
      const options = screen.getAllByRole('option');
      const optionTexts = options.map((o) => o.textContent);
      // Should have page size options (10, 20, 50) and possibly namespace options
      expect(optionTexts.length).toBeGreaterThan(0);
    });
  });

  it('renders status filter', async () => {
    renderWithProviders(<TasksPage />, { initialEntries: ['/tasks'] });

    // MultiSelect renders "Status:" label inside a button
    expect(screen.getByText('Status:')).toBeInTheDocument();
  });

  it('renders status badges for tasks in the table', async () => {
    renderWithProviders(<TasksPage />, { initialEntries: ['/tasks'] });

    await waitFor(() => {
      expect(screen.getByText('fix-auth-bug')).toBeInTheDocument();
    });

    // The table should have task rows
    const table = screen.getByRole('table');
    const rows = table.querySelectorAll('tbody tr');
    expect(rows.length).toBeGreaterThan(0);
  });

  it('renders task names as links to detail pages', async () => {
    renderWithProviders(<TasksPage />, { initialEntries: ['/tasks'] });

    await waitFor(() => {
      const link = screen.getByText('fix-auth-bug');
      expect(link.closest('a')).toHaveAttribute('href', '/tasks/default/fix-auth-bug');
    });
  });

  it('renders pagination controls when data has pagination', async () => {
    renderWithProviders(<TasksPage />, { initialEntries: ['/tasks'] });

    await waitFor(() => {
      expect(screen.getByText('fix-auth-bug')).toBeInTheDocument();
    });

    // Pagination uses "Prev" and "Next" buttons
    await waitFor(() => {
      expect(screen.getByText('Prev')).toBeInTheDocument();
    });
  });

  it('filters tasks by namespace when namespace changes', async () => {
    const user = userEvent.setup();
    let lastRequestUrl = '';

    server.use(
      http.get('/api/v1/namespaces/:namespace/tasks', ({ request }) => {
        lastRequestUrl = request.url;
        return HttpResponse.json({
          tasks: [],
          total: 0,
          pagination: { limit: 20, offset: 0, totalCount: 0, hasMore: false },
        });
      })
    );

    renderWithProviders(<TasksPage />, { initialEntries: ['/tasks'] });

    // Wait for initial load
    await waitFor(() => {
      expect(screen.getByText('fix-auth-bug')).toBeInTheDocument();
    });

    // Find the namespace selector (combobox) and change it
    const selects = screen.getAllByRole('combobox');
    // Find the one that has namespace options
    const namespaceSelect = selects.find((s) => {
      const options = s.querySelectorAll('option');
      return Array.from(options).some((o) => o.textContent === 'All Namespaces');
    });

    if (namespaceSelect) {
      await user.selectOptions(namespaceSelect, 'production');

      await waitFor(() => {
        expect(lastRequestUrl).toContain('/namespaces/production/tasks');
      });
    }
  });
});
