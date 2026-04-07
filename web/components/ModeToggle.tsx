"use client";

import { Menu, ActionIcon } from '@mantine/core';
import { useState, useEffect } from 'react';
import { useRouter } from 'next/navigation';

type Mode = 'client' | 'admin';

export default function ModeToggle() {
  const router = useRouter();
  const [mode, setMode] = useState<Mode>('client');

  useEffect(() => {
    const saved = localStorage.getItem('booking-mode') as Mode;
    if (saved) setMode(saved);
  }, []);

  const handleSetMode = (newMode: Mode) => {
    setMode(newMode);
    localStorage.setItem('booking-mode', newMode);
    if (newMode === 'admin') {
      router.push('/admin');
    } else {
      router.push('/');
    }
  };

  return (
    <Menu shadow="md" width={200}>
      <Menu.Target>
        <ActionIcon variant="subtle" size="lg" title="Режим">
          ⚙️
        </ActionIcon>
      </Menu.Target>
      <Menu.Dropdown>
        <Menu.Label>Режим</Menu.Label>
        <Menu.Item 
          onClick={() => handleSetMode('client')}
          style={{ backgroundColor: mode === 'client' ? '#f0f0f0' : undefined }}
        >
          Режим записи
        </Menu.Item>
        <Menu.Item 
          onClick={() => handleSetMode('admin')}
          style={{ backgroundColor: mode === 'admin' ? '#f0f0f0' : undefined }}
        >
          Админ-режим
        </Menu.Item>
      </Menu.Dropdown>
    </Menu>
  );
}
