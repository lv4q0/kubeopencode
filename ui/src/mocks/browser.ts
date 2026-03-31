// MSW browser worker for development mode
// Intercepts API requests in the browser so the UI can run without a backend

import { setupWorker } from 'msw/browser';
import { handlers } from './handlers';

export const worker = setupWorker(...handlers);
