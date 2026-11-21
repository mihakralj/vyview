const { test, expect } = require('@playwright/test')
const AxeBuilder = require('@axe-core/playwright').default
const { pathToFileURL } = require('node:url')
const { resolve } = require('node:path')

test('local CSS and accessibility tooling work together', async ({ page }) => {
  await page.goto(pathToFileURL(resolve('testdata/tooling/smoke.html')).href)
  await expect(page.getByRole('heading', { name: 'Toolchain check' })).toBeVisible()
  await expect(page.locator('main')).toHaveCSS('padding-top', '16px')
  const results = await new AxeBuilder({ page }).analyze()
  expect(results.violations.filter(({ impact }) => impact === 'serious' || impact === 'critical')).toEqual([])
})
