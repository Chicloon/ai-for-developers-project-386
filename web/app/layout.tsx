import type { Metadata } from "next";
import "@mantine/core/styles.css";
import "@mantine/dates/styles.css";
import "@mantine/notifications/styles.css";
import "@mantine/schedule/styles.css";
import {
  MantineProvider,
  ColorSchemeScript,
  createTheme,
  MantineColorsTuple,
} from "@mantine/core";
import { Notifications } from "@mantine/notifications";
import { DatesProvider } from "@mantine/dates";
import { AuthProvider } from "@/components/auth/AuthProvider";
import dayjs from "dayjs";
import "dayjs/locale/ru";

export const metadata: Metadata = {
  title: "Call Booking",
  description: "Бронирование времени для звонков",
};

// Light theme primary colors based on #E67E22 (pumpkin warm)
const lightPrimaryColors: MantineColorsTuple = [
  "#FFF5EB",
  "#FFE6D1",
  "#FFCCAA",
  "#F5B07A",
  "#EB9955",
  "#E67E22",
  "#D4701F",
  "#BF621B",
  "#A85417",
  "#8F4613",
];

// Dark theme primary colors based on #FF8C00 (amber accent)
const darkPrimaryColors: MantineColorsTuple = [
  "#FFF4E0",
  "#FFE4B8",
  "#FFD08A",
  "#FFBC5C",
  "#FFA82E",
  "#FF8C00",
  "#E67D00",
  "#CC6F00",
  "#B36000",
  "#995200",
];

const theme = createTheme({
  primaryColor: "orange",
  colors: {
    orange: lightPrimaryColors,
    darkOrange: darkPrimaryColors,
  },
  primaryShade: { light: 5, dark: 5 },
  other: {
    backgroundColors: {
      light: "#FDF6E3",
      dark: "#1A1A2E",
    },
  },
});

// Настройка русской локали для dayjs
dayjs.locale("ru");

export default function RootLayout({ children }: Readonly<{ children: React.ReactNode }>) {
  return (
    <html lang="ru" suppressHydrationWarning>
      <head><ColorSchemeScript /></head>
      <body>
        <MantineProvider theme={theme} defaultColorScheme="dark">
          <DatesProvider settings={{ locale: "ru" }}>
            <Notifications position="top-right" zIndex={10000} />
            <AuthProvider>{children}</AuthProvider>
          </DatesProvider>
        </MantineProvider>
      </body>
    </html>
  );
}
