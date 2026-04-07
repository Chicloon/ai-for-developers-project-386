"use client";

import { useEffect, useState } from "react";
import {
  Paper,
  Title,
  Stack,
  Accordion,
  Button,
  Loader,
  Center,
  Text,
  Modal,
  TextInput,
  Group,
  ActionIcon,
  Badge,
} from "@mantine/core";
import { useDisclosure } from "@mantine/hooks";
import { IconTrash, IconUserPlus } from "@tabler/icons-react";
import {
  Group as GroupType,
  User,
  getMyGroups,
  createGroup,
  deleteGroup,
  getGroupMembers,
  addGroupMember,
  removeGroupMember,
} from "@/lib/api";
import { useAuth } from "@/components/auth/AuthProvider";

export default function MyGroupsPage() {
  const { user: currentUser } = useAuth();
  const [groups, setGroups] = useState<GroupType[]>([]);
  const [groupMembers, setGroupMembers] = useState<Record<string, User[]>>({});
  const [loading, setLoading] = useState(true);
  const [createOpened, { open: openCreate, close: closeCreate }] = useDisclosure(false);
  const [addMemberOpened, { open: openAddMember, close: closeAddMember }] = useDisclosure(false);
  const [submitting, setSubmitting] = useState(false);
  const [selectedGroupId, setSelectedGroupId] = useState<string | null>(null);

  // Form state
  const [groupName, setGroupName] = useState("");
  const [memberEmail, setMemberEmail] = useState("");

  useEffect(() => {
    loadGroups();
  }, []);

  const loadGroups = async () => {
    try {
      setLoading(true);
      const data = await getMyGroups();
      setGroups(data);

      // Load members for each group
      const membersMap: Record<string, User[]> = {};
      for (const group of data) {
        try {
          const members = await getGroupMembers(group.id);
          membersMap[group.id] = members;
        } catch (e) {
          console.error(e);
        }
      }
      setGroupMembers(membersMap);
    } catch (e) {
      console.error(e);
    } finally {
      setLoading(false);
    }
  };

  const handleCreateGroup = async () => {
    if (!groupName.trim()) return;
    try {
      setSubmitting(true);
      await createGroup({ name: groupName.trim() });
      closeCreate();
      setGroupName("");
      await loadGroups();
    } catch (e) {
      console.error(e);
    } finally {
      setSubmitting(false);
    }
  };

  const handleDeleteGroup = async (id: string) => {
    try {
      await deleteGroup(id);
      await loadGroups();
    } catch (e) {
      console.error(e);
    }
  };

  const handleAddMember = async () => {
    if (!selectedGroupId || !memberEmail.trim()) return;
    try {
      setSubmitting(true);
      await addGroupMember(selectedGroupId, memberEmail.trim());
      closeAddMember();
      setMemberEmail("");
      setSelectedGroupId(null);
      await loadGroups();
    } catch (e) {
      console.error(e);
    } finally {
      setSubmitting(false);
    }
  };

  const handleRemoveMember = async (groupId: string, userId: string) => {
    try {
      await removeGroupMember(groupId, userId);
      await loadGroups();
    } catch (e) {
      console.error(e);
    }
  };

  const openAddMemberModal = (groupId: string) => {
    setSelectedGroupId(groupId);
    openAddMember();
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
      <Group justify="space-between">
        <Title order={2}>Мои группы</Title>
        <Button onClick={openCreate}>Создать группу</Button>
      </Group>

      {groups.length === 0 ? (
        <Text c="dimmed">У вас пока нет групп</Text>
      ) : (
        <Accordion>
          {groups.map((group) => {
            const members = groupMembers[group.id] || [];
            const isOwner = group.ownerId === currentUser?.id;

            return (
              <Accordion.Item key={group.id} value={group.id}>
                <Accordion.Control>
                  <Group justify="space-between">
                    <Text fw={500}>{group.name}</Text>
                    {isOwner && <Badge color="blue">Владелец</Badge>}
                  </Group>
                </Accordion.Control>
                <Accordion.Panel>
                  <Stack gap="xs">
                    {members.length === 0 ? (
                      <Text c="dimmed" size="sm">
                        В группе пока нет участников
                      </Text>
                    ) : (
                      members.map((member) => (
                        <Paper key={member.id} withBorder p="xs">
                          <Group justify="space-between">
                            <div>
                              <Text size="sm" fw={500}>
                                {member.name}
                              </Text>
                              <Text size="xs" c="dimmed">
                                {member.email}
                              </Text>
                            </div>
                            {isOwner && member.id !== currentUser?.id && (
                              <ActionIcon
                                color="red"
                                size="sm"
                                onClick={() =>
                                  handleRemoveMember(group.id, member.id)
                                }
                              >
                                <IconTrash size={14} />
                              </ActionIcon>
                            )}
                          </Group>
                        </Paper>
                      ))
                    )}

                    {isOwner && (
                      <Group justify="space-between" mt="xs">
                        <Button
                          size="xs"
                          variant="light"
                          leftSection={<IconUserPlus size={14} />}
                          onClick={() => openAddMemberModal(group.id)}
                        >
                          Добавить участника
                        </Button>
                        <ActionIcon
                          color="red"
                          onClick={() => handleDeleteGroup(group.id)}
                        >
                          <IconTrash size={16} />
                        </ActionIcon>
                      </Group>
                    )}
                  </Stack>
                </Accordion.Panel>
              </Accordion.Item>
            );
          })}
        </Accordion>
      )}

      {/* Create Group Modal */}
      <Modal
        opened={createOpened}
        onClose={closeCreate}
        title="Создать группу"
      >
        <Stack gap="md">
          <TextInput
            label="Название группы"
            placeholder="Введите название"
            value={groupName}
            onChange={(e) => setGroupName(e.target.value)}
          />
          <Group justify="flex-end">
            <Button variant="default" onClick={closeCreate}>
              Отмена
            </Button>
            <Button
              onClick={handleCreateGroup}
              loading={submitting}
              disabled={!groupName.trim()}
            >
              Создать
            </Button>
          </Group>
        </Stack>
      </Modal>

      {/* Add Member Modal */}
      <Modal
        opened={addMemberOpened}
        onClose={closeAddMember}
        title="Добавить участника"
      >
        <Stack gap="md">
          <TextInput
            label="Email пользователя"
            placeholder="user@example.com"
            value={memberEmail}
            onChange={(e) => setMemberEmail(e.target.value)}
          />
          <Group justify="flex-end">
            <Button variant="default" onClick={closeAddMember}>
              Отмена
            </Button>
            <Button
              onClick={handleAddMember}
              loading={submitting}
              disabled={!memberEmail.trim()}
            >
              Добавить
            </Button>
          </Group>
        </Stack>
      </Modal>
    </Stack>
  );
}
