"use client";

import { ActionIcon, Menu, useMantineColorScheme } from "@mantine/core";
import { IconSun, IconMoon } from "@tabler/icons-react";
import { useEffect } from "react";

const THEME_STORAGE_KEY = "call-booking-theme";

type Theme = "light" | "dark";

export default function ThemeToggle() {
  const { colorScheme, setColorScheme } = useMantineColorScheme();

  // Load theme preference from localStorage on mount
  useEffect(() => {
    const savedTheme = localStorage.getItem(THEME_STORAGE_KEY) as Theme | null;
    if (savedTheme && (savedTheme === "light" || savedTheme === "dark")) {
      setColorScheme(savedTheme);
    }
  }, [setColorScheme]);

  const handleSetTheme = (theme: Theme) => {
    setColorScheme(theme);
    localStorage.setItem(THEME_STORAGE_KEY, theme);
  };

  const toggleTheme = () => {
    const newTheme = colorScheme === "dark" ? "light" : "dark";
    handleSetTheme(newTheme);
  };

  // Show sun icon when in dark mode (clicking switches to light)
  // Show moon icon when in light mode (clicking switches to dark)
  const isDark = colorScheme === "dark";

  return (
    <Menu shadow="md" width={200} position="bottom-end">
      <Menu.Target>
        <ActionIcon
          variant="subtle"
          size="lg"
          title={isDark ? "Переключить на светлую тему" : "Переключить на тёмную тему"}
          onClick={toggleTheme}
          aria-label={isDark ? "Переключить на светлую тему" : "Переключить на тёмную тему"}
        >
          {isDark ? (
            <IconSun size={20} stroke={1.5} />
          ) : (
            <IconMoon size={20} stroke={1.5} />
          )}
        </ActionIcon>
      </Menu.Target>
      <Menu.Dropdown>
        <Menu.Label>Тема оформления</Menu.Label>
        <Menu.Item
          onClick={() => handleSetTheme("light")}
          leftSection={<IconSun size={16} stroke={1.5} />}
          style={{
            backgroundColor: !isDark ? "var(--mantine-color-blue-1)" : undefined,
          }}
        >
          Светлая тема
        </Menu.Item>
        <Menu.Item
          onClick={() => handleSetTheme("dark")}
          leftSection={<IconMoon size={16} stroke={1.5} />}
          style={{
            backgroundColor: isDark ? "var(--mantine-color-blue-1)" : undefined,
          }}
        >
          Тёмная тема
        </Menu.Item>
      </Menu.Dropdown>
    </Menu>
  );
}
