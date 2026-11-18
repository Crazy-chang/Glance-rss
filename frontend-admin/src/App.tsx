import type { ReactElement } from 'react';
import { HashRouter, Navigate, Route, Routes } from 'react-router-dom';
import AdminLayout from './components/AdminLayout';
import Categories from './pages/Categories';
import Login from './pages/Login';
import Sources from './pages/Sources';
import System from './pages/System';
import Users from './pages/Users';
import { getToken } from './api';

function RequireAuth({ children }: { children: ReactElement }) {
  return getToken() ? children : <Navigate to="/login" replace />;
}

export default function App() {
  return (
    <HashRouter>
      <Routes>
        <Route path="/login" element={<Login />} />
        <Route
          path="/"
          element={
            <RequireAuth>
              <AdminLayout />
            </RequireAuth>
          }
        >
          <Route index element={<Navigate to="/recommend/categories" replace />} />
          <Route path="recommend">
            <Route index element={<Navigate to="/recommend/categories" replace />} />
            <Route path="categories" element={<Categories />} />
            <Route path="sources" element={<Sources />} />
          </Route>
          <Route path="users" element={<Users />} />
          <Route path="system" element={<System />} />
        </Route>
        <Route path="*" element={<Navigate to="/" replace />} />
      </Routes>
    </HashRouter>
  );
}
