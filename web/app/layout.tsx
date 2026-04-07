import '@mantine/core/styles.css';
import '@mantine/dates/styles.css';
import { MantineProvider, createTheme, Group, Container, Title } from '@mantine/core';
import ModeToggle from '@/components/ModeToggle';

const theme = createTheme({
  primaryColor: 'blue',
  fontFamily: 'system-ui, -apple-system, sans-serif',
});

export const metadata = {
  title: 'Запись на звонок',
  description: 'Выберите удобное время для звонка',
};

export default function RootLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <html lang="ru">
      <body>
        <MantineProvider theme={theme}>
          <header style={{ borderBottom: '1px solid #eee', padding: '12px 0', marginBottom: 24 }}>
            <Container size="sm">
              <Group justify="space-between">
                <Title order={3}>Запись на звонок !</Title>
                <ModeToggle />
              </Group>
            </Container>
          </header>
          <main style={{ maxWidth: 960, margin: '0 auto', padding: '0 16px' }}>
            {children}
          </main>
        </MantineProvider>
      </body>
    </html>
  );
}
