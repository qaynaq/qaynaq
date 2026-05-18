import { Navigate, useLocation } from 'react-router-dom';
import { useAuth, UserRole } from '@/lib/auth';

interface ProtectedRouteProps {
  children: React.ReactNode;
  requireRole?: 'admin' | 'mcp' | 'none';
}

const homeForRole = (role: UserRole): string => {
  if (role === 'admin') return '/';
  if (role === 'mcp') return '/mcp-access';
  return '/no-access';
};

export default function ProtectedRoute({ children, requireRole }: ProtectedRouteProps) {
  const { isAuthenticated, isLoading, role } = useAuth();
  const location = useLocation();

  if (isLoading) {
    return (
      <div className="min-h-screen flex items-center justify-center">
        <div className="text-center">
          <div className="inline-block h-8 w-8 animate-spin rounded-full border-4 border-solid border-current border-r-transparent align-[-0.125em] motion-reduce:animate-[spin_1.5s_linear_infinite]" />
          <p className="mt-4 text-slate-600 dark:text-slate-400">Loading...</p>
        </div>
      </div>
    );
  }

  if (!isAuthenticated) {
    const target = `${location.pathname}${location.search}${location.hash}`;
    const dest =
      target && target !== '/'
        ? `/login?return_to=${encodeURIComponent(target)}`
        : '/login';
    return <Navigate to={dest} replace />;
  }

  if (requireRole) {
    const userRole: 'admin' | 'mcp' | 'none' = role === '' ? 'none' : role;
    if (userRole !== requireRole) {
      return <Navigate to={homeForRole(role)} replace />;
    }
  }

  return <>{children}</>;
}
