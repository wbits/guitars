import React from 'react';
import ReactDOM from 'react-dom/client';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import {
  createBrowserRouter,
  Navigate,
  RouterProvider,
} from 'react-router-dom';
import { App } from './App';
import { AuthGate } from './components/AuthGate';
import { GuitarList } from './pages/GuitarList';
import { GuitarNew } from './pages/GuitarNew';
import { GuitarView } from './pages/GuitarView';
import { GuitarEdit } from './pages/GuitarEdit';
import { Settings } from './pages/Settings';
import './index.css';

const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      retry: 1,
      staleTime: 30_000,
      refetchOnWindowFocus: false,
    },
  },
});

const router = createBrowserRouter([
  {
    path: '/',
    element: <App />,
    children: [
      { index: true, element: <Navigate to="/guitars" replace /> },
      {
        path: 'guitars',
        element: (
          <AuthGate>
            <GuitarList />
          </AuthGate>
        ),
      },
      {
        path: 'guitars/new',
        element: (
          <AuthGate>
            <GuitarNew />
          </AuthGate>
        ),
      },
      {
        path: 'guitars/:id',
        element: (
          <AuthGate>
            <GuitarView />
          </AuthGate>
        ),
      },
      {
        path: 'guitars/:id/edit',
        element: (
          <AuthGate>
            <GuitarEdit />
          </AuthGate>
        ),
      },
      { path: 'settings', element: <Settings /> },
      { path: '*', element: <Navigate to="/guitars" replace /> },
    ],
  },
]);

ReactDOM.createRoot(document.getElementById('root') as HTMLElement).render(
  <React.StrictMode>
    <QueryClientProvider client={queryClient}>
      <RouterProvider router={router} />
    </QueryClientProvider>
  </React.StrictMode>,
);
