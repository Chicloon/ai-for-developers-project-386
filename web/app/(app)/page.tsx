"use client";

import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import { Paper, Title, Stack, Card, Group, Text, Button, Loader, Center } from "@mantine/core";
import { User, getUsers } from "@/lib/api";

export default function UsersPage() {
  const router = useRouter();
  const [users, setUsers] = useState<User[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    loadUsers();
  }, []);

  const loadUsers = async () => {
    try {
      setLoading(true);
      const data = await getUsers();
      setUsers(data);
    } catch (e) {
      console.error(e);
    } finally {
      setLoading(false);
    }
  };

  if (loading) {
    return (
      <Center h="50vh">
        <Loader />
      </Center>
    );
  }

  return (
    <Stack gap="md">
      <Title order={2}>Каталог пользователей</Title>
      {users.length === 0 ? (
        <Text c="dimmed">Пока нет доступных пользователей</Text>
      ) : (
        users.map((user) => (
          <Card key={user.id} withBorder>
            <Group justify="space-between">
              <div>
                <Text fw={500}>{user.name}</Text>
                <Text size="sm" c="dimmed">{user.email}</Text>
              </div>
              <Button onClick={() => router.push(`/users/${user.id}`)} variant="light">Записаться</Button>
            </Group>
          </Card>
        ))
      )}
    </Stack>
  );
}
