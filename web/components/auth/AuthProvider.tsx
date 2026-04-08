"use client";

import React, { createContext, useContext, useEffect, useState } from "react";
import { User, getMe, getAuthToken, setAuthToken, logout as apiLogout } from "@/lib/api";

interface AuthContextType {
  user: User | null;
  isLoading: boolean;
  isAuthenticated: boolean;
  login: (token: string, user: User) => void;
  logout: () => void;
}

const AuthContext = createContext<AuthContextType | undefined>(undefined);

export function AuthProvider({ children }: { children: React.ReactNode }) {
  const [user, setUser] = useState<User | null>(null);
  const [isLoading, setIsLoading] = useState(true);

  useEffect(() => {
    const token = getAuthToken();
    if (token) {
      getMe()
        .then((userData) => {
          setUser(userData);
        })
        .catch(() => {
          apiLogout();
        })
        .finally(() => {
          setIsLoading(false);
        });
    } else {
      setIsLoading(false);
    }
  }, []);

  const login = (token: string, userData: User) => {
    // #region agent log H4
    fetch('http://127.0.0.1:7924/ingest/df065418-75a6-4c94-b505-bfe4e2e4e84a',{method:'POST',headers:{'Content-Type':'application/json','X-Debug-Session-Id':'eb49d8'},body:JSON.stringify({sessionId:'eb49d8',runId:'debug1',hypothesisId:'H4',location:'AuthProvider.tsx:login',message:'AuthProvider login called',data:{hasToken:!!token,tokenLength:token?.length,hasUser:!!userData,userId:userData?.id},timestamp:Date.now()})}).catch(()=>{});
    // #endregion
    setAuthToken(token);
    setUser(userData);
  };

  const logout = () => {
    apiLogout();
    setUser(null);
  };

  return (
    <AuthContext.Provider
      value={{
        user,
        isLoading,
        isAuthenticated: !!user,
        login,
        logout,
      }}
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
