import { Page, Locator, expect } from '@playwright/test'

/** Страница каталога «Запись на встречу» (/users): Select + календарь слотов. */
export class UsersPage {
  readonly page: Page
  readonly title: Locator
  readonly userSelect: Locator
  readonly visibilityHint: Locator

  constructor(page: Page) {
    this.page = page
    this.title = page.locator('[data-testid="users-title"]')
    this.userSelect = page.locator('[data-testid="users-select"]')
    this.visibilityHint = page.locator('[data-testid="users-visibility-hint"]')
  }

  async goto() {
    await this.page.goto('/users')
    await expect(this.page.locator('[data-testid="users-page"]')).toBeVisible()
  }

  /** Выбор пользователя в searchable Select (как в Mantine: ввод и клик по опции). */
  async selectUserByEmail(email: string) {
    await this.userSelect.click()
    await this.userSelect.fill(email)
    await this.page.locator(`text=${email}`).first().click()
  }

  async getUserCard(userId: string) {
    return this.page.locator(`[data-testid="user-card-${userId}"]`)
  }

  getUserName(userId: string): Locator {
    return this.page.locator(`[data-testid="user-name-${userId}"]`)
  }

  getUserEmail(userId: string): Locator {
    return this.page.locator(`[data-testid="user-email-${userId}"]`)
  }

  async expectUserVisible(userId: string, name: string, email: string) {
    const card = await this.getUserCard(userId)
    await expect(card).toBeVisible()
    await expect(this.getUserName(userId)).toContainText(name)
    await expect(this.getUserEmail(userId)).toContainText(email)
  }
}
