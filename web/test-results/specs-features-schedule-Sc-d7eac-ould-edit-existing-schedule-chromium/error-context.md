# Instructions

- Following Playwright test failed.
- Explain why, be concise, respect Playwright best practices.
- Provide a snippet of code with the fix, if possible.

# Test info

- Name: specs/features/schedule.spec.ts >> Schedule Management >> should edit existing schedule
- Location: e2e/specs/features/schedule.spec.ts:88:7

# Error details

```
TimeoutError: locator.click: Timeout 10000ms exceeded.
Call log:
  - waiting for locator('[data-testid="schedule-day-select"]')

```

# Page snapshot

```yaml
- generic [ref=e1]:
  - generic [ref=e6] [cursor=pointer]:
    - button "Open Next.js Dev Tools" [ref=e7]:
      - img [ref=e8]
    - generic [ref=e11]:
      - button "Open issues overlay" [ref=e12]:
        - generic [ref=e13]:
          - generic [ref=e14]: "0"
          - generic [ref=e15]: "1"
        - generic [ref=e16]: Issue
      - button "Collapse issues badge" [ref=e17]:
        - img [ref=e18]
  - alert [ref=e20]
  - generic [ref=e21]:
    - banner [ref=e22]:
      - generic [ref=e23]:
        - link "Call Booking" [ref=e25] [cursor=pointer]:
          - /url: /
        - generic [ref=e26] [cursor=pointer]:
          - generic [ref=e28]: T
          - paragraph [ref=e29]: Test User 1775631208504
    - navigation [ref=e30]:
      - generic [ref=e31]:
        - link "Каталог пользователей" [ref=e32] [cursor=pointer]:
          - /url: /
          - generic [ref=e34]: Каталог пользователей
        - link "Моё расписание" [ref=e35] [cursor=pointer]:
          - /url: /my/schedule
          - generic [ref=e37]: Моё расписание
        - link "Мои группы" [ref=e38] [cursor=pointer]:
          - /url: /my/groups
          - generic [ref=e40]: Мои группы
        - link "Мои бронирования" [ref=e41] [cursor=pointer]:
          - /url: /my/bookings
          - generic [ref=e43]: Мои бронирования
    - main [ref=e44]:
      - generic [ref=e45]:
        - generic [ref=e46]:
          - heading "Моё расписание" [level=2] [ref=e47]
          - button "Добавить расписание" [ref=e48] [cursor=pointer]:
            - generic [ref=e50]: Добавить расписание
        - paragraph [ref=e51]: У вас пока нет настроенных расписаний
  - dialog "Добавить расписание" [ref=e53]:
    - banner [ref=e54]:
      - heading "Добавить расписание" [level=2] [ref=e55]
      - button [ref=e56] [cursor=pointer]:
        - img [ref=e57]
    - generic [ref=e60]:
      - generic [ref=e61]:
        - generic [ref=e62]: Тип расписания
        - generic [ref=e63]:
          - combobox "Тип расписания" [active] [ref=e64] [cursor=pointer]
          - generic:
            - img
      - generic [ref=e65]:
        - generic [ref=e66]: Дата
        - textbox "Дата" [ref=e68]
      - generic [ref=e69]:
        - generic [ref=e70]:
          - generic [ref=e71]: Начало
          - textbox "Начало" [ref=e73]: 09:00
        - generic [ref=e74]:
          - generic [ref=e75]: Конец
          - textbox "Конец" [ref=e77]: 17:00
      - generic [ref=e79]:
        - generic [ref=e80]:
          - checkbox "Заблокировать (недоступно для бронирования)" [ref=e81]
          - img
        - generic [ref=e83]: Заблокировать (недоступно для бронирования)
      - generic [ref=e84]:
        - button "Отмена" [ref=e85] [cursor=pointer]:
          - generic [ref=e87]: Отмена
        - button "Создать" [ref=e88] [cursor=pointer]:
          - generic [ref=e90]: Создать
```

# Test source

```ts
  1   | import { Page, Locator, expect } from '@playwright/test'
  2   | 
  3   | export class SchedulePage {
  4   |   readonly page: Page
  5   |   readonly title: Locator
  6   |   readonly addButton: Locator
  7   |   readonly loadingIndicator: Locator
  8   |   readonly emptyMessage: Locator
  9   |   readonly modal: Locator
  10  | 
  11  |   constructor(page: Page) {
  12  |     this.page = page
  13  |     this.title = page.locator('[data-testid="schedule-title"]')
  14  |     this.addButton = page.locator('[data-testid="schedule-add-button"]')
  15  |     this.loadingIndicator = page.locator('[data-testid="schedule-loading"]')
  16  |     this.emptyMessage = page.locator('[data-testid="schedule-empty"]')
  17  |     this.modal = page.locator('[data-testid="schedule-modal"]')
  18  |   }
  19  | 
  20  |   async goto() {
  21  |     await this.page.goto('/my/schedule')
  22  |     await expect(this.page.locator('[data-testid="schedule-page"]')).toBeVisible()
  23  |   }
  24  | 
  25  |   async waitForLoad() {
  26  |     await this.loadingIndicator.waitFor({ state: 'hidden' })
  27  |   }
  28  | 
  29  |   async clickAdd() {
  30  |     await this.addButton.click()
  31  |     // Wait for modal to open and check content is visible
  32  |     await expect(this.page.locator('[data-testid="schedule-type-select"]')).toBeVisible({ timeout: 5000 })
  33  |   }
  34  | 
  35  |   async fillScheduleForm(params: {
  36  |     type: 'recurring' | 'one-time'
  37  |     dayOfWeek?: string
  38  |     date?: string
  39  |     startTime: string
  40  |     endTime: string
  41  |     isBlocked?: boolean
  42  |   }) {
  43  |     // Select type - click on the select to open dropdown
  44  |     await this.page.locator('[data-testid="schedule-type-select"]').click()
  45  |     await this.page.waitForTimeout(300)
  46  |     // Click on the option by text in the dropdown (Mantine renders dropdown in portal)
  47  |     const typeLabel = params.type === 'recurring' ? 'Повторяющееся' : 'Разовое'
  48  |     await this.page.locator('.mantine-Select-option', { hasText: typeLabel }).click()
  49  |     // Wait for form to re-render based on type
  50  |     await this.page.waitForTimeout(400)
  51  | 
  52  |     if (params.type === 'recurring' && params.dayOfWeek) {
> 53  |       await this.page.locator('[data-testid="schedule-day-select"]').click()
      |                                                                      ^ TimeoutError: locator.click: Timeout 10000ms exceeded.
  54  |       await this.page.waitForTimeout(300)
  55  |       // Map day number to Russian name
  56  |       const dayNames: Record<string, string> = {
  57  |         '0': 'Воскресенье',
  58  |         '1': 'Понедельник',
  59  |         '2': 'Вторник',
  60  |         '3': 'Среда',
  61  |         '4': 'Четверг',
  62  |         '5': 'Пятница',
  63  |         '6': 'Суббота'
  64  |       }
  65  |       await this.page.locator('.mantine-Select-option', { hasText: dayNames[params.dayOfWeek] }).click()
  66  |     } else if (params.type === 'one-time' && params.date) {
  67  |       await this.page.locator('[data-testid="schedule-date-input"]').fill(params.date)
  68  |     }
  69  | 
  70  |     await this.page.locator('[data-testid="schedule-start-time"]').fill(params.startTime)
  71  |     await this.page.locator('[data-testid="schedule-end-time"]').fill(params.endTime)
  72  | 
  73  |     if (params.isBlocked) {
  74  |       await this.page.locator('[data-testid="schedule-blocked-checkbox"]').check()
  75  |     }
  76  |   }
  77  | 
  78  |   async submitForm() {
  79  |     await this.page.locator('[data-testid="schedule-submit-button"]').click()
  80  |     await this.modal.waitFor({ state: 'hidden' })
  81  |   }
  82  | 
  83  |   async cancelForm() {
  84  |     await this.page.locator('[data-testid="schedule-cancel-button"]').click()
  85  |     await this.modal.waitFor({ state: 'hidden' })
  86  |   }
  87  | 
  88  |   async getScheduleRow(scheduleId: string) {
  89  |     return this.page.locator(`[data-testid="schedule-row-${scheduleId}"]`)
  90  |   }
  91  | 
  92  |   async deleteSchedule(scheduleId: string) {
  93  |     await this.page.locator(`[data-testid="schedule-delete-${scheduleId}"]`).click()
  94  |   }
  95  | 
  96  |   async editSchedule(scheduleId: string) {
  97  |     await this.page.locator(`[data-testid="schedule-edit-${scheduleId}"]`).click()
  98  |     // Wait for modal to open and check content is visible
  99  |     await expect(this.page.locator('[data-testid="schedule-type-select"]')).toBeVisible({ timeout: 5000 })
  100 |   }
  101 | 
  102 |   async expectScheduleVisible(scheduleId: string, timeRange: string) {
  103 |     const row = await this.getScheduleRow(scheduleId)
  104 |     await expect(row).toBeVisible()
  105 |     await expect(this.page.locator(`[data-testid="schedule-time-${scheduleId}"]`)).toContainText(timeRange)
  106 |   }
  107 | }
  108 | 
```