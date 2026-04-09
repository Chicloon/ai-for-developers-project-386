import { test, expect, registerAndLogin } from '../../fixtures/auth'

/**
 * Группы видимости теперь на странице «Моё расписание» → вкладка «Настройки видимости».
 * Отдельного маршрута /my/groups в приложении нет.
 */
test.describe('Visibility groups (schedule tab)', () => {
  test.beforeEach(async ({ page, testUser }) => {
    await registerAndLogin(page, testUser)
  })

  test('should show fixed groups on visibility tab', async ({ page }) => {
    await page.goto('/my/schedule')
    await expect(page.locator('[data-testid="schedule-page"]')).toBeVisible()
    await page.locator('[data-testid="tab-visibility"]').click()
    await expect(page.getByText('Мои группы')).toBeVisible()
    await expect(page.getByText('Семья').first()).toBeVisible()
    await expect(page.getByText('Работа').first()).toBeVisible()
    await expect(page.getByText('Друзья').first()).toBeVisible()
  })

  test('should show public profile toggle', async ({ page }) => {
    await page.goto('/my/schedule')
    await page.locator('[data-testid="tab-visibility"]').click()
    await expect(page.getByText('Публичный профиль')).toBeVisible()
  })
})
