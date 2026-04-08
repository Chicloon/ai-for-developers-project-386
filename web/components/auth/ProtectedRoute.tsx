"use client";

import { useEffect } from "react";
import { useRouter } from "next/navigation";
import { Loader, Center } from "@mantine/core";
import { useAuth } from "./AuthProvider";

export default function ProtectedRoute({ children }: { children: React.ReactNode }) {
  const { isAuthenticated, isLoading } = useAuth();
  const router = useRouter();

  useEffect(() => {
    // #region agent log H3,H4
    fetch('http://127.0.0.1:7924/ingest/df065418-75a6-4c94-b505-bfe4e2e4e84a',{method:'POST',headers:{'Content-Type':'application/json','X-Debug-Session-Id':'eb49d8'},body:JSON.stringify({sessionId:'eb49d8',runId:'debug1',hypothesisId:'H4',location:'ProtectedRoute.tsx:useEffect',message:'ProtectedRoute check',data:{isLoading,isAuthenticated},timestamp:Date.now()})}).catch(()=>{});
    // #endregion
    if (!isLoading && !isAuthenticated) {
      // #region agent log H3
      fetch('http://127.0.0.1:7924/ingest/df065418-75a6-4c94-b505-bfe4e2e4e84a',{method:'POST',headers:{'Content-Type':'application/json','X-Debug-Session-Id':'eb49d8'},body:JSON.stringify({sessionId:'eb49d8',runId:'debug1',hypothesisId:'H3',location:'ProtectedRoute.tsx:redirectToLogin',message:'Redirecting to login',data:{},timestamp:Date.now()})}).catch(()=>{});
      // #endregion
      router.push("/login");
    }
  }, [isLoading, isAuthenticated, router]);

  if (isLoading) {
    return (
      <Center h="100vh">
        <Loader />
      </Center>
    );
  }

  if (!isAuthenticated) {
    return null;
  }

  return <>{children}</>;
}
