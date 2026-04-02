import React, {
  createContext,
  useContext,
  useState,
  useEffect,
  ReactNode,
} from "react";

interface AuthContextType {
  isAuthenticated: boolean;
  isLoading: boolean;
  authType: string;
  token: string | null;
  login: (token: string) => void;
  logout: () => void;
}

const AuthContext = createContext<AuthContextType | undefined>(undefined);

export function AuthProvider({ children }: { children: ReactNode }) {
  const [isAuthenticated, setIsAuthenticated] = useState(false);
  const [isLoading, setIsLoading] = useState(true);
  const [authType, setAuthType] = useState("none");
  const [token, setToken] = useState<string | null>(null);

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
        } else {
          const savedToken = localStorage.getItem("qaynaq_token");
          if (savedToken) {
            setToken(savedToken);
            setIsAuthenticated(true);
          } else {
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

  const login = (newToken: string) => {
    localStorage.setItem("qaynaq_token", newToken);
    setToken(newToken);
    setIsAuthenticated(true);
  };

  const logout = async () => {
    const savedToken = localStorage.getItem("qaynaq_token");
    if (savedToken) {
      await fetch("/auth/logout", {
        method: "POST",
        headers: {
          Authorization: `Bearer ${savedToken}`,
        },
      });
    }
    localStorage.removeItem("qaynaq_token");
    setToken(null);
    setIsAuthenticated(false);
    window.location.reload();
  };

  return (
    <AuthContext.Provider
      value={{ isAuthenticated, isLoading, authType, token, login, logout }}
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
