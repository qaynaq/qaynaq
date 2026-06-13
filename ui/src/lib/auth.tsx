import React, {
  createContext,
  useContext,
  useState,
  useEffect,
  ReactNode,
} from "react";

export type UserRole = "admin" | "mcp" | "";

interface AuthContextType {
  isAuthenticated: boolean;
  isLoading: boolean;
  authType: string;
  token: string | null;
  role: UserRole;
  email: string;
  login: (token: string) => Promise<void>;
  logout: () => void;
}

const AuthContext = createContext<AuthContextType | undefined>(undefined);

async function exchangeSessionCookie(token: string): Promise<void> {
  try {
    await fetch("/auth/exchange", {
      method: "POST",
      headers: { Authorization: `Bearer ${token}` },
      credentials: "include",
    });
  } catch (error) {
    console.error("Session cookie exchange failed:", error);
  }
}

type SessionResult =
  | { status: "authenticated"; role: UserRole; email: string }
  | { status: "unauthenticated" }
  | { status: "unknown" };

async function fetchSession(): Promise<SessionResult> {
  try {
    const resp = await fetch("/auth/session", { credentials: "include" });
    if (!resp.ok) return { status: "unknown" };
    const data = await resp.json();
    if (!data.authenticated) return { status: "unauthenticated" };
    return {
      status: "authenticated",
      role: (data.role || "") as UserRole,
      email: data.email || "",
    };
  } catch {
    return { status: "unknown" };
  }
}

export function AuthProvider({ children }: { children: ReactNode }) {
  const [isAuthenticated, setIsAuthenticated] = useState(false);
  const [isLoading, setIsLoading] = useState(true);
  const [authType, setAuthType] = useState("none");
  const [token, setToken] = useState<string | null>(null);
  const [role, setRole] = useState<UserRole>("");
  const [email, setEmail] = useState<string>("");

  useEffect(() => {
    const savedToken = localStorage.getItem("qaynaq_token");
    if (savedToken) {
      setToken(savedToken);
    }
    checkAuthStatus();
  }, []);

  const checkAuthStatus = async () => {
    try {
      const infoResponse = await fetch("/auth/info", {
        credentials: "include",
      });

      if (infoResponse.ok) {
        const infoData = await infoResponse.json();
        const authTypeValue = infoData.auth_type || "none";
        setAuthType(authTypeValue);

        if (authTypeValue === "none") {
          setIsAuthenticated(true);
          setRole("admin");
        } else {
          const session = await fetchSession();
          if (session.status === "authenticated") {
            setIsAuthenticated(true);
            setRole(session.role);
            setEmail(session.email);
            const savedToken = localStorage.getItem("qaynaq_token");
            if (savedToken) {
              setToken(savedToken);
            }
          } else if (session.status === "unauthenticated") {
            setIsAuthenticated(false);
            localStorage.removeItem("qaynaq_token");
          } else {
            // Request never completed (aborted by a reload, network blip, 5xx).
            // Not a real logout - keep the token and stay put.
            setIsAuthenticated(false);
          }
        }
      } else {
        setIsAuthenticated(false);
        setAuthType("none");
      }
    } catch (error) {
      console.error("Auth check failed:", error);
      setIsAuthenticated(false);
    } finally {
      setIsLoading(false);
    }
  };

  const login = async (newToken: string) => {
    localStorage.setItem("qaynaq_token", newToken);
    setToken(newToken);
    setIsAuthenticated(true);
    await exchangeSessionCookie(newToken);
    const session = await fetchSession();
    if (session.status === "authenticated") {
      setRole(session.role);
      setEmail(session.email);
    }
  };

  const logout = async () => {
    const savedToken = localStorage.getItem("qaynaq_token");
    const headers: Record<string, string> = {};
    if (savedToken) {
      headers.Authorization = `Bearer ${savedToken}`;
    }
    try {
      // credentials:"include" is required so the browser sends our session
      // cookie and accepts the Set-Cookie that clears it server-side.
      await fetch("/auth/logout", {
        method: "POST",
        headers,
        credentials: "include",
      });
    } catch (error) {
      console.error("Logout request failed:", error);
    }
    localStorage.removeItem("qaynaq_token");
    setToken(null);
    setIsAuthenticated(false);
    window.location.reload();
  };

  return (
    <AuthContext.Provider
      value={{ isAuthenticated, isLoading, authType, token, role, email, login, logout }}
    >
      {children}
    </AuthContext.Provider>
  );
}

export function useAuth() {
  const context = useContext(AuthContext);
  if (context === undefined) {
    throw new Error("useAuth must be used within an AuthProvider");
  }
  return context;
}
