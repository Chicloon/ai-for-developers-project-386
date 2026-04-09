import { Page, Locator, expect } from '@playwright/test'

/** «Мои бронирования»: календарь Mantine Schedule, без таблицы строк с data-testid. */
export class BookingsPage {
  readonly page: Page
  readonly title: Locator
  readonly loadingIndicator: Locator
  readonly emptyMessage: Locator
  readonly scheduleWrap: Locator

  constructor(page: Page) {
    this.page = page
    this.title = page.locator('[data-testid="bookings-title"]')
    this.loadingIndicator = page.locator('[data-testid="bookings-loading"]')
    this.emptyMessage = page.locator('[data-testid="bookings-empty"]')
    this.scheduleWrap = page.locator('.bookings-schedule-wrap')
  }

  async goto() {
    await this.page.goto('/my/bookings')
    await expect(this.page.locator('[data-testid="bookings-page"]')).toBeVisible()
  }

  async waitForLoad() {
    await this.loadingIndicator.waitFor({ state: 'hidden' })
  }
}
