import type { Metadata } from "next";
import "@mantine/core/styles.css";
import "@mantine/dates/styles.css";
import "@mantine/schedule/styles.css";
import { MantineProvider, ColorSchemeScript, createTheme } from "@mantine/core";
import { AuthProvider } from "@/components/auth/AuthProvider";
import dayjs from "dayjs";
import "dayjs/locale/ru";

export const metadata: Metadata = {
  title: "Call Booking",
  description: "Бронирование времени для звонков",
};

const theme = createTheme({ primaryColor: "blue" });

// Настройка русской локали для dayjs
dayjs.locale("ru");

export default function RootLayout({ children }: Readonly<{ children: React.ReactNode }>) {
  return (
    <html lang="ru" suppressHydrationWarning>
      <head><ColorSchemeScript /></head>
      <body>
        <MantineProvider theme={theme} defaultColorScheme="light" dates={{ locale: "ru" }}>
          <AuthProvider>{children}</AuthProvider>
        </MantineProvider>
      </body>
    </html>
  );
}
