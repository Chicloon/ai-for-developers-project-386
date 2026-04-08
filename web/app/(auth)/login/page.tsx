"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import { Paper, Title, TextInput, PasswordInput, Button, Stack, Text, Anchor } from "@mantine/core";
import { login } from "@/lib/api";
import { useAuth } from "@/components/auth/AuthProvider";

export default function LoginPage() {
  const router = useRouter();
  const { login: authLogin } = useAuth();
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [error, setError] = useState("");
  const [loading, setLoading] = useState(false);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError("");
    setLoading(true);

    try {
      // #region agent log H1,H2
      fetch('http://127.0.0.1:7924/ingest/df065418-75a6-4c94-b505-bfe4e2e4e84a',{method:'POST',headers:{'Content-Type':'application/json','X-Debug-Session-Id':'eb49d8'},body:JSON.stringify({sessionId:'eb49d8',runId:'debug1',hypothesisId:'H1',location:'login/page.tsx:handleSubmit:entry',message:'Login form submitted',data:{email:email},timestamp:Date.now()})}).catch(()=>{});
      // #endregion
      const response = await login({ email, password });
      // #region agent log H2
      fetch('http://127.0.0.1:7924/ingest/df065418-75a6-4c94-b505-bfe4e2e4e84a',{method:'POST',headers:{'Content-Type':'application/json','X-Debug-Session-Id':'eb49d8'},body:JSON.stringify({sessionId:'eb49d8',runId:'debug1',hypothesisId:'H2',location:'login/page.tsx:handleSubmit:loginSuccess',message:'Login API succeeded',data:{hasToken:!!response.token,hasUser:!!response.user,userId:response.user?.id},timestamp:Date.now()})}).catch(()=>{});
      // #endregion
      authLogin(response.token, response.user);
      // #region agent log H4
      fetch('http://127.0.0.1:7924/ingest/df065418-75a6-4c94-b505-bfe4e2e4e84a',{method:'POST',headers:{'Content-Type':'application/json','X-Debug-Session-Id':'eb49d8'},body:JSON.stringify({sessionId:'eb49d8',runId:'debug1',hypothesisId:'H4',location:'login/page.tsx:handleSubmit:beforeRouterPush',message:'About to call router.push',data:{},timestamp:Date.now()})}).catch(()=>{});
      // #endregion
      router.push("/");
    } catch (err) {
      // #region agent log H1,H2
      fetch('http://127.0.0.1:7924/ingest/df065418-75a6-4c94-b505-bfe4e2e4e84a',{method:'POST',headers:{'Content-Type':'application/json','X-Debug-Session-Id':'eb49d8'},body:JSON.stringify({sessionId:'eb49d8',runId:'debug1',hypothesisId:'H1',location:'login/page.tsx:handleSubmit:error',message:'Login failed with error',data:{error:err instanceof Error?err.message:String(err)},timestamp:Date.now()})}).catch(()=>{});
      // #endregion
      setError(err instanceof Error ? err.message : "Ошибка входа");
    } finally {
      setLoading(false);
    }
  };

  return (
    <Paper p="xl" maw={400} mx="auto" mt={100} withBorder data-testid="login-page">
      <Title order={2} mb="lg" ta="center" data-testid="login-title">Вход</Title>
      <form onSubmit={handleSubmit} data-testid="login-form">
        <Stack>
          <TextInput 
            label="Email" 
            type="email" 
            value={email} 
            onChange={(e) => setEmail(e.target.value)} 
            required 
            data-testid="login-email-input"
          />
          <PasswordInput 
            label="Пароль" 
            value={password} 
            onChange={(e) => setPassword(e.target.value)} 
            required 
            data-testid="login-password-input"
          />
          {error && <Text c="red" size="sm" data-testid="login-error">{error}</Text>}
          <Button type="submit" loading={loading} fullWidth data-testid="login-submit-button">
            Войти
          </Button>
          <Text ta="center" size="sm">
            Нет аккаунта? <Anchor href="/register" data-testid="login-register-link">Зарегистрироваться</Anchor>
          </Text>
        </Stack>
      </form>
    </Paper>
  );
}
