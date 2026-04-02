import React, { useEffect } from "react";
import {
  Routes,
  Route,
  Outlet,
  useNavigate,
  useSearchParams,
} from "react-router-dom";
import { ThemeProvider } from "@/components/theme-provider";
import { TooltipProvider } from "@/components/ui/tooltip";
import { ToastProvider } from "@/components/toast";
import { Helmet, HelmetProvider } from "react-helmet-async";
import { AuthProvider } from "@/lib/auth";
import ProtectedRoute from "@/components/ProtectedRoute";

import Layout from "./components/Layout";

import LoginPage from "./pages/login/page.tsx";
import HomePage from "./pages/HomePage.tsx";
import FlowsPage from "./pages/flows/page.tsx";
import WorkersPage from "./pages/workers/page.tsx";
import BuffersPage from "./pages/buffers/page.tsx";
import BufferNewPage from "./pages/buffers/new/page.tsx";
import BufferEditPage from "./pages/buffers/[id]/edit/page.tsx";
import CachesPage from "./pages/caches/page.tsx";
import SecretsPage from "./pages/secrets/page.tsx";
import ConnectionsPage from "./pages/connections/page.tsx";
import CacheNewPage from "./pages/caches/new/page.tsx";
import CacheEditPage from "./pages/caches/[id]/edit/page.tsx";
import RateLimitsPage from "./pages/rate-limits/page.tsx";
import RateLimitNewPage from "./pages/rate-limits/new/page.tsx";
import RateLimitEditPage from "./pages/rate-limits/[id]/edit/page.tsx";
import FlowEditPage from "./pages/flows/[id]/edit/page.tsx";
import FlowEventsPage from "./pages/flows/[id]/events/page.tsx";
import FlowNewPage from "./pages/flows/new/page.tsx";
import FilesPage from "./pages/files/page.tsx";
import SettingsPage from "./pages/settings/page.tsx";

const AppLayout: React.FC = () => {
  return (
    <TooltipProvider>
      <ToastProvider>
        <Helmet>
          <title>Qaynaq Dashboard</title>
          <meta
            name="description"
            content="Dashboard for managing flows and components"
          />
        </Helmet>
        <Layout>
          <Outlet />
        </Layout>
      </ToastProvider>
    </TooltipProvider>
  );
};

function TokenHandler() {
  const navigate = useNavigate();
  const [searchParams] = useSearchParams();

  useEffect(() => {
    const token = searchParams.get("token");
    if (token) {
      localStorage.setItem("qaynaq_token", token);
      navigate("/", { replace: true });
      window.location.reload();
    }
  }, [searchParams, navigate]);

  return null;
}

function App() {
  return (
    <ThemeProvider
      attribute="class"
      defaultTheme="system"
      enableSystem
      disableTransitionOnChange
    >
      <AuthProvider>
        <TokenHandler />
        <Routes>
          <Route path="/login" element={<LoginPage />} />
          <Route
            path="/"
            element={
              <ProtectedRoute>
                <AppLayout />
              </ProtectedRoute>
            }
          >
            <Route index element={<HomePage />} />
            <Route path="flows" element={<FlowsPage />} />
            <Route path="flows/new" element={<FlowNewPage />} />
            <Route path="flows/:id/edit" element={<FlowEditPage />} />
            <Route path="flows/:id/events" element={<FlowEventsPage />} />
            <Route path="workers" element={<WorkersPage />} />
            <Route path="secrets" element={<SecretsPage />} />
            <Route path="connections" element={<ConnectionsPage />} />
            <Route path="buffers" element={<BuffersPage />} />
            <Route path="buffers/new" element={<BufferNewPage />} />
            <Route path="buffers/:id/edit" element={<BufferEditPage />} />
            <Route path="caches" element={<CachesPage />} />
            <Route path="caches/new" element={<CacheNewPage />} />
            <Route path="caches/:id/edit" element={<CacheEditPage />} />
            <Route path="rate-limits" element={<RateLimitsPage />} />
            <Route path="rate-limits/new" element={<RateLimitNewPage />} />
            <Route
              path="rate-limits/:id/edit"
              element={<RateLimitEditPage />}
            />
            <Route path="files" element={<FilesPage />} />
            <Route path="settings" element={<SettingsPage />} />
          </Route>
        </Routes>
      </AuthProvider>
    </ThemeProvider>
  );
}

export default App;
