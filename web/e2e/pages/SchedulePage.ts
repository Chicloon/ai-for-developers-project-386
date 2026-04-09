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
    await this.page.locator('[data-testid="tab-schedule"]').click()
    await this.addButton.click()
    // Mantine Modal root остаётся в DOM с aria-hidden — ждём контент формы
    await expect(this.page.locator('[data-testid="schedule-submit-button"]')).toBeVisible({ timeout: 10000 })
    await expect(this.page.locator('[data-testid="schedule-type-select"]')).toBeVisible({ timeout: 10000 })
  }

  async fillScheduleForm(params: {
    type: 'recurring' | 'one-time'
    dayOfWeek?: string
    date?: string
    startTime: string
    endTime: string
    isBlocked?: boolean
  }) {
    const dayNames: Record<string, string> = {
      '0': 'Воскресенье',
      '1': 'Понедельник',
      '2': 'Вторник',
      '3': 'Среда',
      '4': 'Четверг',
      '5': 'Пятница',
      '6': 'Суббота'
    }

    if (params.type === 'one-time') {
      await this.page.locator('[data-testid="schedule-type-select"]').click()
      await this.page.getByRole('option', { name: 'Разовое' }).click()
      await this.page.waitForTimeout(200)
      if (params.date) {
        const dateField = this.page.locator('[data-testid="schedule-date-input"]')
        await expect(dateField).toBeVisible()
        const dateInner = dateField.locator('input')
        if ((await dateInner.count()) > 0) {
          await dateInner.fill(params.date)
        } else {
          await dateField.fill(params.date)
        }
      }
    } else if (params.dayOfWeek) {
      // Модалка по умолчанию «Повторяющееся» — не трогаем тип, только день недели
      await this.page.locator('[data-testid="schedule-day-select"]').click()
      await this.page.waitForTimeout(200)
      await this.page.getByRole('option', { name: dayNames[params.dayOfWeek] }).click()
    }

    const fillTime = async (testId: string, value: string) => {
      const root = this.page.locator(`[data-testid="${testId}"]`)
      const inner = root.locator('input')
      if ((await inner.count()) > 0) {
        await inner.fill(value)
      } else {
        await root.fill(value)
      }
    }
    await fillTime('schedule-start-time', params.startTime)
    await fillTime('schedule-end-time', params.endTime)

    if (params.isBlocked) {
      await this.page.locator('[data-testid="schedule-blocked-checkbox"]').check()
    }
  }

  async submitForm() {
    await this.page.locator('[data-testid="schedule-submit-button"]').click()
    await expect(this.page.locator('[data-testid="schedule-submit-button"]')).toBeHidden({ timeout: 15000 })
  }

  async cancelForm() {
    await this.page.locator('[data-testid="schedule-cancel-button"]').click()
    await expect(this.page.locator('[data-testid="schedule-submit-button"]')).toBeHidden({ timeout: 10000 })
  }

  async getScheduleRow(scheduleId: string) {
    return this.page.locator(`[data-testid="schedule-row-${scheduleId}"]`)
  }

  async deleteSchedule(scheduleId: string) {
    await this.page.locator(`[data-testid="schedule-delete-${scheduleId}"]`).click()
  }

  async editSchedule(scheduleId: string) {
    await this.page.locator(`[data-testid="schedule-edit-${scheduleId}"]`).click()
    // Wait for modal to open and check content is visible
    await expect(this.page.locator('[data-testid="schedule-type-select"]')).toBeVisible({ timeout: 5000 })
  }

  async expectScheduleVisible(scheduleId: string, timeRange: string) {
    const row = await this.getScheduleRow(scheduleId)
    await expect(row).toBeVisible()
    await expect(this.page.locator(`[data-testid="schedule-time-${scheduleId}"]`)).toContainText(timeRange)
  }
}
