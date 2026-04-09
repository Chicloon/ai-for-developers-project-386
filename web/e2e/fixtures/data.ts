export interface TestUser {
  name: string
  email: string
  password: string
}

export function generateTestUser(): TestUser {
  const timestamp = Date.now()
  return {
    name: `Test User ${timestamp}`,
    email: `test${timestamp}@example.com`,
    password: 'TestPassword123!'
  }
}

export function generateTestUsers(count: number): TestUser[] {
  return Array.from({ length: count }, (_, i) => {
    const timestamp = Date.now() + i
    return {
      name: `Test User ${timestamp}`,
      email: `test${timestamp}@example.com`,
      password: 'TestPassword123!'
    }
  })
}

/** Локальная дата YYYY-MM-DD (без сдвига UTC как у toISOString). */
export function formatDate(date: Date): string {
  const y = date.getFullYear()
  const m = String(date.getMonth() + 1).padStart(2, '0')
  const d = String(date.getDate()).padStart(2, '0')
  return `${y}-${m}-${d}`
}

export function formatTime(date: Date): string {
  return date.toTimeString().slice(0, 5)
}

export function getTomorrow(): string {
  const tomorrow = new Date()
  tomorrow.setDate(tomorrow.getDate() + 1)
  return formatDate(tomorrow)
}

export function getNextWeek(): string {
  const nextWeek = new Date()
  nextWeek.setDate(nextWeek.getDate() + 7)
  return formatDate(nextWeek)
}
