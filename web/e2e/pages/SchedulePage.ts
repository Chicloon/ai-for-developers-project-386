import { Page, Locator, expect } from '@playwright/test'

export class SchedulePage {
  readonly page: Page
  readonly title: Locator
  readonly addButton: Locator
  readonly loadingIndicator: Locator
  readonly emptyMessage: Locator
  readonly modal: Locator

  constructor(page: Page) {
    this.page = page
    this.title = page.locator('[data-testid="schedule-title"]')
    this.addButton = page.locator('[data-testid="schedule-add-button"]')
    this.loadingIndicator = page.locator('[data-testid="schedule-loading"]')
    this.emptyMessage = page.locator('[data-testid="schedule-empty"]')
    this.modal = page.locator('[data-testid="schedule-modal"]')
  }

  async goto() {
    await this.page.goto('/my/schedule')
    await expect(this.page.locator('[data-testid="schedule-page"]')).toBeVisible()
  }

  async waitForLoad() {
    await this.loadingIndicator.waitFor({ state: 'hidden' })
  }

  async clickAdd() {
    await this.addButton.click()
    await expect(this.modal).toBeVisible()
  }

  async fillScheduleForm(params: {
    type: 'recurring' | 'one-time'
    dayOfWeek?: string
    date?: string
    startTime: string
    endTime: string
    isBlocked?: boolean
  }) {
    // Select type
    await this.page.locator('[data-testid="schedule-type-select"]').click()
    await this.page.locator(`[data-value="${params.type}"]`).click()

    if (params.type === 'recurring' && params.dayOfWeek) {
      await this.page.locator('[data-testid="schedule-day-select"]').click()
      await this.page.locator(`[data-value="${params.dayOfWeek}"]`).click()
    } else if (params.type === 'one-time' && params.date) {
      await this.page.locator('[data-testid="schedule-date-input"]').fill(params.date)
    }

    await this.page.locator('[data-testid="schedule-start-time"]').fill(params.startTime)
    await this.page.locator('[data-testid="schedule-end-time"]').fill(params.endTime)

    if (params.isBlocked) {
      await this.page.locator('[data-testid="schedule-blocked-checkbox"]').check()
    }
  }

  async submitForm() {
    await this.page.locator('[data-testid="schedule-submit-button"]').click()
    await this.modal.waitFor({ state: 'hidden' })
  }

  async cancelForm() {
    await this.page.locator('[data-testid="schedule-cancel-button"]').click()
    await this.modal.waitFor({ state: 'hidden' })
  }

  async getScheduleRow(scheduleId: string) {
    return this.page.locator(`[data-testid="schedule-row-${scheduleId}"]`)
  }

  async deleteSchedule(scheduleId: string) {
    await this.page.locator(`[data-testid="schedule-delete-${scheduleId}"]`).click()
  }

  async editSchedule(scheduleId: string) {
    await this.page.locator(`[data-testid="schedule-edit-${scheduleId}"]`).click()
    await expect(this.modal).toBeVisible()
  }

  async expectScheduleVisible(scheduleId: string, timeRange: string) {
    const row = await this.getScheduleRow(scheduleId)
    await expect(row).toBeVisible()
    await expect(this.page.locator(`[data-testid="schedule-time-${scheduleId}"]`)).toContainText(timeRange)
  }
}
