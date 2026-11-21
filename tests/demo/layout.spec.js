const { test, expect } = require('@playwright/test')

test('flat square surfaces and active indicators are consistent across views', async ({ page }) => {
  await page.goto('/')
  for (const selector of ['.card', '.source', '.demo-badge', '.range', '.range a', '.rail nav a', '.brand-mark', '.state']) {
    await expect(page.locator(selector).first()).toHaveCSS('border-radius', '0px')
  }
  await expect(page.locator('.card').first()).toHaveCSS('box-shadow', 'none')
  await expect(page.locator('.rail nav .nav-active')).toHaveCSS('border-left-color', 'rgb(34, 211, 238)')
  await expect(page.locator('.range a[aria-current]')).toHaveCSS('border-bottom-color', 'rgb(34, 211, 238)')
  await expect(page.locator('.rail')).toHaveCSS('width', '232px')
  await expect(page.locator('.rail nav .nav-icon').first()).toHaveCSS('width', '22px')
  await expect(page.locator('.rail nav .nav-label').first()).toBeVisible()
  await page.getByRole('button', { name: 'Toggle sidebar' }).click()
  await expect(page.locator('html')).toHaveAttribute('data-rail', 'collapsed')
  await expect(page.locator('.rail')).toHaveCSS('width', '64px')
  await expect(page.locator('.rail nav .nav-icon').first()).toBeVisible()
  await expect(page.locator('.rail nav .nav-label').first()).toBeHidden()
  await page.keyboard.press('[')
  await expect(page.locator('html')).not.toHaveAttribute('data-rail', 'collapsed')
  await expect(page.locator('.rail')).toHaveCSS('width', '232px')
  await expect(page.locator('.rail nav .nav-label').first()).toBeVisible()

  await page.goto('/system?scoutTheme=light')
  for (const selector of ['.tabs a', 'input[type="search"]', 'select', 'button', '.heatmap a', '.meter-grid li']) {
    await expect(page.locator(selector).first()).toHaveCSS('border-radius', '0px')
  }
  await expect(page.locator('.tabs a[aria-current]')).toHaveCSS('border-bottom-color', 'rgb(14, 116, 144)')
  await page.goto('/topology')
  await expect(page.locator('.map-root')).toHaveCSS('border-radius', '0px')
  await expect(page.locator('.network-map a').first()).toHaveCSS('border-radius', '0px')
  await page.goto('/clients?tab=paths')
  await expect(page.locator('.visual-grid a').first()).toHaveCSS('border-radius', '0px')
})
