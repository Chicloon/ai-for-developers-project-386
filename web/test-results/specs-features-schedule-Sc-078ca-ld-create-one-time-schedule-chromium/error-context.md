# Instructions

- Following Playwright test failed.
- Explain why, be concise, respect Playwright best practices.
- Provide a snippet of code with the fix, if possible.

# Test info

- Name: specs/features/schedule.spec.ts >> Schedule Management >> should create one-time schedule
- Location: e2e/specs/features/schedule.spec.ts:42:7

# Error details

```
TimeoutError: page.waitForSelector: Timeout 10000ms exceeded.
Call log:
  - waiting for locator('[data-testid^="schedule-row-"]') to be visible

```

# Page snapshot

```yaml
- generic [active] [ref=e1]:
  - generic [ref=e6] [cursor=pointer]:
    - button "Open Next.js Dev Tools" [ref=e7]:
      - img [ref=e8]
    - generic [ref=e11]:
      - button "Open issues overlay" [ref=e12]:
        - generic [ref=e13]:
          - generic [ref=e14]: "1"
          - generic [ref=e15]: "2"
        - generic [ref=e16]:
          - text: Issue
          - generic [ref=e17]: s
      - button "Collapse issues badge" [ref=e18]:
        - img [ref=e19]
  - alert [ref=e21]
  - generic [ref=e22]:
    - banner [ref=e23]:
      - generic [ref=e24]:
        - link "Call Booking" [ref=e26] [cursor=pointer]:
          - /url: /
        - generic [ref=e27] [cursor=pointer]:
          - generic [ref=e29]: T
          - paragraph [ref=e30]: Test User 1775631180881
    - navigation [ref=e31]:
      - generic [ref=e32]:
        - link "Каталог пользователей" [ref=e33] [cursor=pointer]:
          - /url: /
          - generic [ref=e35]: Каталог пользователей
        - link "Моё расписание" [ref=e36] [cursor=pointer]:
          - /url: /my/schedule
          - generic [ref=e38]: Моё расписание
        - link "Мои группы" [ref=e39] [cursor=pointer]:
          - /url: /my/groups
          - generic [ref=e41]: Мои группы
        - link "Мои бронирования" [ref=e42] [cursor=pointer]:
          - /url: /my/bookings
          - generic [ref=e44]: Мои бронирования
    - main [ref=e45]:
      - generic [ref=e46]:
        - generic [ref=e47]:
          - heading "Моё расписание" [level=2] [ref=e48]
          - button "Добавить расписание" [ref=e49] [cursor=pointer]:
            - generic [ref=e51]: Добавить расписание
        - paragraph [ref=e52]: У вас пока нет настроенных расписаний
  - dialog "Добавить расписание" [ref=e54]:
    - banner [ref=e55]:
      - heading "Добавить расписание" [level=2] [ref=e56]
      - button [ref=e57] [cursor=pointer]:
        - img [ref=e58]
    - generic [ref=e61]:
      - generic [ref=e62]:
        - generic [ref=e63]: Тип расписания
        - generic [ref=e64]:
          - combobox "Тип расписания" [ref=e65] [cursor=pointer]: Разовое
          - generic:
            - img
      - generic [ref=e66]:
        - generic [ref=e67]: Дата
        - textbox "Дата" [ref=e69]: 2026-04-09
      - generic [ref=e70]:
        - generic [ref=e71]:
          - generic [ref=e72]: Начало
          - textbox "Начало" [ref=e74]: 10:00
        - generic [ref=e75]:
          - generic [ref=e76]: Конец
          - textbox "Конец" [ref=e78]: 14:00
      - generic [ref=e80]:
        - generic [ref=e81]:
          - checkbox "Заблокировать (недоступно для бронирования)" [ref=e82]
          - img
        - generic [ref=e84]: Заблокировать (недоступно для бронирования)
      - generic [ref=e85]:
        - button "Отмена" [ref=e86] [cursor=pointer]:
          - generic [ref=e88]: Отмена
        - button "Создать" [ref=e89] [cursor=pointer]:
          - generic [ref=e91]: Создать
```

# Test source

```ts
  1   | import { test, expect, registerAndLogin } from '../../fixtures/auth'
  2   | import { getTomorrow } from '../../fixtures/data'
  3   | 
  4   | test.describe('Schedule Management', () => {
  5   |   test.beforeEach(async ({ page, testUser }) => {
  6   |     await registerAndLogin(page, testUser)
  7   |   })
  8   | 
  9   |   test('should display empty state when no schedules', async ({ schedulePage }) => {
  10  |     await schedulePage.goto()
  11  |     await schedulePage.waitForLoad()
  12  |     
  13  |     await expect(schedulePage.emptyMessage).toBeVisible()
  14  |     await expect(schedulePage.emptyMessage).toContainText('У вас пока нет настроенных расписаний')
  15  |   })
  16  | 
  17  |   test('should create recurring schedule', async ({ page, schedulePage }) => {
  18  |     await schedulePage.goto()
  19  |     await schedulePage.waitForLoad()
  20  |     
  21  |     await schedulePage.clickAdd()
  22  |     await schedulePage.fillScheduleForm({
  23  |       type: 'recurring',
  24  |       dayOfWeek: '1', // Monday
  25  |       startTime: '09:00',
  26  |       endTime: '17:00',
  27  |       isBlocked: false
  28  |     })
  29  |     await schedulePage.submitForm()
  30  |     
  31  |     // Wait for the schedule to appear in the table
  32  |     await page.waitForSelector('[data-testid^="schedule-row-"]')
  33  |     
  34  |     // Verify schedule is visible
  35  |     const scheduleRow = page.locator('[data-testid^="schedule-row-"]').first()
  36  |     await expect(scheduleRow).toBeVisible()
  37  |     await expect(page.locator('[data-testid^="schedule-type-"]').first()).toContainText('Повторяющееся')
  38  |     await expect(page.locator('[data-testid^="schedule-day-"]').first()).toContainText('Понедельник')
  39  |     await expect(page.locator('[data-testid^="schedule-time-"]').first()).toContainText('09:00 - 17:00')
  40  |   })
  41  | 
  42  |   test('should create one-time schedule', async ({ page, schedulePage }) => {
  43  |     const tomorrow = getTomorrow()
  44  |     
  45  |     await schedulePage.goto()
  46  |     await schedulePage.waitForLoad()
  47  |     
  48  |     await schedulePage.clickAdd()
  49  |     await schedulePage.fillScheduleForm({
  50  |       type: 'one-time',
  51  |       date: tomorrow,
  52  |       startTime: '10:00',
  53  |       endTime: '14:00',
  54  |       isBlocked: false
  55  |     })
  56  |     await schedulePage.submitForm()
  57  |     
  58  |     // Wait for the schedule to appear
> 59  |     await page.waitForSelector('[data-testid^="schedule-row-"]')
      |                ^ TimeoutError: page.waitForSelector: Timeout 10000ms exceeded.
  60  |     
  61  |     const scheduleRow = page.locator('[data-testid^="schedule-row-"]').first()
  62  |     await expect(scheduleRow).toBeVisible()
  63  |     await expect(page.locator('[data-testid^="schedule-type-"]').first()).toContainText('Разовое')
  64  |     await expect(page.locator('[data-testid^="schedule-time-"]').first()).toContainText('10:00 - 14:00')
  65  |   })
  66  | 
  67  |   test('should create blocked schedule', async ({ page, schedulePage }) => {
  68  |     await schedulePage.goto()
  69  |     await schedulePage.waitForLoad()
  70  |     
  71  |     await schedulePage.clickAdd()
  72  |     await schedulePage.fillScheduleForm({
  73  |       type: 'recurring',
  74  |       dayOfWeek: '6', // Saturday
  75  |       startTime: '00:00',
  76  |       endTime: '23:59',
  77  |       isBlocked: true
  78  |     })
  79  |     await schedulePage.submitForm()
  80  |     
  81  |     // Wait for the schedule to appear
  82  |     await page.waitForSelector('[data-testid^="schedule-row-"]')
  83  |     
  84  |     // Verify blocked status
  85  |     await expect(page.locator('[data-testid^="schedule-status-"]').first()).toContainText('Заблокировано')
  86  |   })
  87  | 
  88  |   test('should edit existing schedule', async ({ page, schedulePage }) => {
  89  |     // First create a schedule
  90  |     await schedulePage.goto()
  91  |     await schedulePage.waitForLoad()
  92  |     
  93  |     await schedulePage.clickAdd()
  94  |     await schedulePage.fillScheduleForm({
  95  |       type: 'recurring',
  96  |       dayOfWeek: '1',
  97  |       startTime: '09:00',
  98  |       endTime: '17:00'
  99  |     })
  100 |     await schedulePage.submitForm()
  101 |     
  102 |     // Wait for schedule to appear and get its ID
  103 |     await page.waitForSelector('[data-testid^="schedule-row-"]')
  104 |     const scheduleId = await page.locator('[data-testid^="schedule-row-"]').first().getAttribute('data-testid')
  105 |     const id = scheduleId?.replace('schedule-row-', '')
  106 |     
  107 |     // Edit the schedule
  108 |     await schedulePage.editSchedule(id!)
  109 |     await schedulePage.page.locator('[data-testid="schedule-start-time"]').fill('08:00')
  110 |     await schedulePage.page.locator('[data-testid="schedule-end-time"]').fill('16:00')
  111 |     await schedulePage.submitForm()
  112 |     
  113 |     // Verify updated time
  114 |     await expect(page.locator(`[data-testid="schedule-time-${id}"]`)).toContainText('08:00 - 16:00')
  115 |   })
  116 | 
  117 |   test('should delete schedule', async ({ page, schedulePage }) => {
  118 |     // Create a schedule first
  119 |     await schedulePage.goto()
  120 |     await schedulePage.waitForLoad()
  121 |     
  122 |     await schedulePage.clickAdd()
  123 |     await schedulePage.fillScheduleForm({
  124 |       type: 'recurring',
  125 |       dayOfWeek: '2',
  126 |       startTime: '09:00',
  127 |       endTime: '17:00'
  128 |     })
  129 |     await schedulePage.submitForm()
  130 |     
  131 |     // Wait for schedule to appear
  132 |     await page.waitForSelector('[data-testid^="schedule-row-"]')
  133 |     const scheduleRow = page.locator('[data-testid^="schedule-row-"]').first()
  134 |     await expect(scheduleRow).toBeVisible()
  135 |     
  136 |     // Get schedule ID and delete
  137 |     const scheduleId = await scheduleRow.getAttribute('data-testid')
  138 |     const id = scheduleId?.replace('schedule-row-', '')
  139 |     await schedulePage.deleteSchedule(id!)
  140 |     
  141 |     // Verify schedule is removed
  142 |     await expect(scheduleRow).not.toBeVisible()
  143 |   })
  144 | 
  145 |   test('should cancel form without saving', async ({ page, schedulePage }) => {
  146 |     await schedulePage.goto()
  147 |     await schedulePage.waitForLoad()
  148 |     
  149 |     await schedulePage.clickAdd()
  150 |     await schedulePage.fillScheduleForm({
  151 |       type: 'recurring',
  152 |       dayOfWeek: '1',
  153 |       startTime: '09:00',
  154 |       endTime: '17:00'
  155 |     })
  156 |     await schedulePage.cancelForm()
  157 |     
  158 |     // Verify no schedule was created
  159 |     await expect(schedulePage.emptyMessage).toBeVisible()
```