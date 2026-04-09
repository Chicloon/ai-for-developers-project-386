import { test as base, expect, Page } from '@playwright/test'
import { LoginPage } from '../pages/LoginPage'
import { RegisterPage } from '../pages/RegisterPage'
import { UsersPage } from '../pages/UsersPage'
import { SchedulePage } from '../pages/SchedulePage'
import { BookingsPage } from '../pages/BookingsPage'
import { GroupsPage } from '../pages/GroupsPage'
import { UserProfilePage } from '../pages/UserProfilePage'
import { generateTestUser, TestUser } from './data'

// Extend base test with fixtures
type Fixtures = {
  loginPage: LoginPage
  registerPage: RegisterPage
  usersPage: UsersPage
  schedulePage: SchedulePage
  bookingsPage: BookingsPage
  groupsPage: GroupsPage
  userProfilePage: UserProfilePage
  testUser: TestUser
}

export const test = base.extend<Fixtures>({
  loginPage: async ({ page }, use) => {
    await use(new LoginPage(page))
  },
  registerPage: async ({ page }, use) => {
    await use(new RegisterPage(page))
  },
  usersPage: async ({ page }, use) => {
    await use(new UsersPage(page))
  },
  schedulePage: async ({ page }, use) => {
    await use(new SchedulePage(page))
  },
  bookingsPage: async ({ page }, use) => {
    await use(new BookingsPage(page))
  },
  groupsPage: async ({ page }, use) => {
    await use(new GroupsPage(page))
  },
  userProfilePage: async ({ page }, use) => {
    await use(new UserProfilePage(page))
  },
  testUser: async ({}, use) => {
    await use(generateTestUser())
  },
})

export { expect }

// Helper function to register and login in one step
export async function registerAndLogin(page: Page, user: TestUser): Promise<void> {
  const registerPage = new RegisterPage(page)

  await registerPage.goto()
  await registerPage.register(user.name, user.email, user.password)
  await expect(page).toHaveURL(/\/($|my\/bookings$)/)
  await expect(page.locator('[data-testid="user-menu"]')).toBeVisible()
}

// Helper function to login existing user
export async function loginUser(page: Page, user: TestUser): Promise<void> {
  const loginPage = new LoginPage(page)

  await loginPage.goto()
  await loginPage.login(user.email, user.password)
  await expect(page).toHaveURL(/\/($|my\/bookings$)/)
  await expect(page.locator('[data-testid="user-menu"]')).toBeVisible()
}
