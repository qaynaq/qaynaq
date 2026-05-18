import React, { Suspense, lazy, useEffect } from "react";
import {
  Routes,
  Route,
  Outlet,
  Navigate,
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

const LoginPage = lazy(() => import("./pages/login/page.tsx"));
const HomePage = lazy(() => import("./pages/HomePage.tsx"));
const FlowsPage = lazy(() => import("./pages/flows/page.tsx"));
const WorkersPage = lazy(() => import("./pages/workers/page.tsx"));
const BuffersPage = lazy(() => import("./pages/buffers/page.tsx"));
const BufferNewPage = lazy(() => import("./pages/buffers/new/page.tsx"));
const BufferEditPage = lazy(() => import("./pages/buffers/[id]/edit/page.tsx"));
const CachesPage = lazy(() => import("./pages/caches/page.tsx"));
const SecretsPage = lazy(() => import("./pages/secrets/page.tsx"));
const ConnectionsPage = lazy(() => import("./pages/connections/page.tsx"));
const CacheNewPage = lazy(() => import("./pages/caches/new/page.tsx"));
const CacheEditPage = lazy(() => import("./pages/caches/[id]/edit/page.tsx"));
const RateLimitsPage = lazy(() => import("./pages/rate-limits/page.tsx"));
const RateLimitNewPage = lazy(() => import("./pages/rate-limits/new/page.tsx"));
const RateLimitEditPage = lazy(
  () => import("./pages/rate-limits/[id]/edit/page.tsx"),
);
const FlowEditPage = lazy(() => import("./pages/flows/[id]/edit/page.tsx"));
const FlowEventsPage = lazy(() => import("./pages/flows/[id]/events/page.tsx"));
const FlowNewPage = lazy(() => import("./pages/flows/new/page.tsx"));
const FilesPage = lazy(() => import("./pages/files/page.tsx"));
const SettingsLayout = lazy(() => import("./pages/settings/layout.tsx"));
const AuthenticationSettings = lazy(
  () => import("./pages/settings/authentication.tsx"),
);
const TokensSettings = lazy(() => import("./pages/settings/tokens.tsx"));
const SessionsSettings = lazy(() => import("./pages/settings/sessions.tsx"));
const ClientsSettings = lazy(() => import("./pages/settings/clients.tsx"));
const MCPServersPage = lazy(() => import("./pages/mcp-servers/page.tsx"));
const OAuthConsentPage = lazy(() => import("./pages/oauth/consent.tsx"));
const OAuthErrorPage = lazy(() => import("./pages/oauth/error.tsx"));
const McpAccessPage = lazy(() => import("./pages/mcp-access/page.tsx"));
const NoAccessPage = lazy(() => import("./pages/no-access/page.tsx"));

const RouteFallback: React.FC = () => (
  <div className="flex h-full min-h-[200px] items-center justify-center p-8 text-sm text-muted-foreground">
    Loading...
  </div>
);

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
          <Suspense fallback={<RouteFallback />}>
            <Outlet />
          </Suspense>
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
        <Suspense fallback={<RouteFallback />}>
          <Routes>
            <Route path="/login" element={<LoginPage />} />
            <Route path="/oauth/error" element={<OAuthErrorPage />} />
            <Route
              path="/no-access"
              element={
                <ProtectedRoute requireRole="none">
                  <NoAccessPage />
                </ProtectedRoute>
              }
            />
            <Route
              path="/oauth/consent"
              element={
                <ProtectedRoute>
                  <OAuthConsentPage />
                </ProtectedRoute>
              }
            />
            <Route
              path="/mcp-access"
              element={
                <ProtectedRoute requireRole="mcp">
                  <McpAccessPage />
                </ProtectedRoute>
              }
            />
            <Route
              path="/"
              element={
                <ProtectedRoute requireRole="admin">
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
              <Route path="settings" element={<SettingsLayout />}>
                <Route
                  index
                  element={<Navigate to="/settings/authentication" replace />}
                />
                <Route
                  path="authentication"
                  element={<AuthenticationSettings />}
                />
                <Route path="tokens" element={<TokensSettings />} />
                <Route path="sessions" element={<SessionsSettings />} />
                <Route path="clients" element={<ClientsSettings />} />
              </Route>
              <Route path="mcp-servers" element={<MCPServersPage />} />
            </Route>
          </Routes>
        </Suspense>
      </AuthProvider>
    </ThemeProvider>
  );
}

export default App;
