"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import { Paper, Title, TextInput, PasswordInput, Button, Stack, Text, Anchor } from "@mantine/core";
import { register } from "@/lib/api";
import { useAuth } from "@/components/auth/AuthProvider";

export default function RegisterPage() {
  const router = useRouter();
  const { login: authLogin } = useAuth();
  const [name, setName] = useState("");
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [error, setError] = useState("");
  const [loading, setLoading] = useState(false);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError("");
    setLoading(true);

    try {
      const response = await register({ name, email, password });
      authLogin(response.token, response.user);
      router.push("/");
    } catch (err) {
      setError(err instanceof Error ? err.message : "Ошибка регистрации");
    } finally {
      setLoading(false);
    }
  };

  return (
    <Paper p="xl" maw={400} mx="auto" mt={100} withBorder>
      <Title order={2} mb="lg" ta="center">Регистрация</Title>
      <form onSubmit={handleSubmit}>
        <Stack>
          <TextInput label="Имя" value={name} onChange={(e) => setName(e.target.value)} required />
          <TextInput label="Email" type="email" value={email} onChange={(e) => setEmail(e.target.value)} required />
          <PasswordInput label="Пароль" value={password} onChange={(e) => setPassword(e.target.value)} required />
          {error && <Text c="red" size="sm">{error}</Text>}
          <Button type="submit" loading={loading} fullWidth>Зарегистрироваться</Button>
          <Text ta="center" size="sm">Уже есть аккаунт? <Anchor href="/login">Войти</Anchor></Text>
        </Stack>
      </form>
    </Paper>
  );
}
