import { test, expect } from '../../fixtures/auth'
import { generateTestUser } from '../../fixtures/data'

test.describe('Users Booking Page UX', () => {
  test('should show updated title, visibility hint and no nearest button', async ({ page, request }) => {
    const viewer = generateTestUser()
    const viewerRegister = await request.post('/api/auth/register', {
      data: viewer,
    })
    expect(viewerRegister.ok()).toBeTruthy()
    const viewerAuth = await viewerRegister.json()

    await page.addInitScript((token: string) => {
      window.localStorage.setItem('auth_token', token)
    }, viewerAuth.token)
    await page.goto('/users')

    await expect(page.locator('[data-testid="users-title"]')).toHaveText('Запись на встречу')
    await expect(page.locator('text=Поиск осуществляется по пользователям с публичным профилем')).toBeVisible()
    await expect(page.locator('button:has-text("Записаться на ближайшее")')).toHaveCount(0)
    await expect(page.locator('text=Шаг 1')).toBeVisible()
    await expect(page.locator('text=Шаг 2')).toBeVisible()
    await expect(page.locator('text=Шаг 3')).toBeVisible()
  })

  test('should hide visibility hint after selecting user', async ({ page, request }) => {
    const owner = generateTestUser()
    const ownerRegister = await request.post('/api/auth/register', {
      data: owner,
    })
    expect(ownerRegister.ok()).toBeTruthy()
    const ownerAuth = await ownerRegister.json()
    const makeOwnerPublic = await request.put('/api/users/me', {
      headers: { Authorization: `Bearer ${ownerAuth.token}` },
      data: { isPublic: true },
    })
    expect(makeOwnerPublic.ok()).toBeTruthy()

    const viewer = generateTestUser()
    const viewerRegister = await request.post('/api/auth/register', {
      data: viewer,
    })
    expect(viewerRegister.ok()).toBeTruthy()
    const viewerAuth = await viewerRegister.json()

    await page.addInitScript((token: string) => {
      window.localStorage.setItem('auth_token', token)
    }, viewerAuth.token)

    await page.goto('/users')

    const selectInput = page.locator('[data-testid="users-select"]')
    await selectInput.click()
    await selectInput.fill(owner.email)
    await page.locator(`text=${owner.email}`).first().click()

    await expect(page.locator('text=Поиск осуществляется по пользователям с публичным профилем')).toHaveCount(0)
  })
})
