import { test, expect, registerAndLogin } from '../../fixtures/auth'
import { getTomorrow } from '../../fixtures/data'

test.describe('Booking Flows', () => {
  test.beforeEach(async ({ page, testUser }) => {
    await registerAndLogin(page, testUser)
  })

  test('should display empty state when no bookings', async ({ bookingsPage }) => {
    await bookingsPage.goto()
    await bookingsPage.waitForLoad()

    await expect(bookingsPage.emptyMessage).toBeVisible()
    await expect(bookingsPage.emptyMessage).toContainText('У вас пока нет бронирований')
  })

  test('should open users booking page from nav', async ({ page }) => {
    await page.locator('[data-testid="nav-users"]').click()
    await expect(page).toHaveURL('/users')
    await expect(page.locator('[data-testid="users-title"]')).toContainText('Запись на встречу')
  })

  test('should create schedule and see it in list', async ({ page, schedulePage }) => {
    await schedulePage.goto()
    await schedulePage.waitForLoad()

    await schedulePage.clickAdd()
    await schedulePage.fillScheduleForm({
      type: 'recurring',
      dayOfWeek: '1',
      startTime: '09:00',
      endTime: '17:00',
      isBlocked: false
    })
    await schedulePage.submitForm()

    await expect(page.locator('[data-testid^="schedule-row-"]')).toBeVisible()
  })

  test('should book flow reach bookings list', async ({ bookingsPage }) => {
    await bookingsPage.goto()
    await bookingsPage.waitForLoad()
    await expect(bookingsPage.emptyMessage).toBeVisible()
  })

  test('should create one-time schedule', async ({ page, schedulePage }) => {
    const tomorrow = getTomorrow()

    await schedulePage.goto()
    await schedulePage.waitForLoad()

    await schedulePage.clickAdd()
    await schedulePage.fillScheduleForm({
      type: 'one-time',
      date: tomorrow,
      startTime: '10:00',
      endTime: '14:00',
      isBlocked: false
    })
    await schedulePage.submitForm()

    await page.waitForSelector('[data-testid^="schedule-row-"]')
    await expect(page.locator('[data-testid^="schedule-type-"]').first()).toContainText('Разовое')
    const timeCell = page.locator('[data-testid^="schedule-time-"]').first()
    await expect(timeCell).toContainText('10:00')
    await expect(timeCell).toContainText('14:00')
  })
})
