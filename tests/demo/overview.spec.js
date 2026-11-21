const { test, expect } = require('@playwright/test')
const AxeBuilder = require('@axe-core/playwright').default

for (const viewport of [{ width: 1440, height: 900 }, { width: 390, height: 844 }]) {
  test(`overview is usable and accessible at ${viewport.width}px`, async ({ page }) => {
    await page.setViewportSize(viewport)
    await page.goto('/')
    await expect(page.getByRole('heading', { name: 'Overview', exact: true })).toBeVisible()
    await expect(page.locator('.demo-badge')).toBeVisible()
    await expect(page.locator('.demo-badge')).toContainText('Demo data')
    await expect(page.getByRole('img', { name: 'Throughput · WAN eth0' })).toBeVisible()
    expect(await page.evaluate(() => document.documentElement.scrollWidth <= window.innerWidth)).toBe(true)
    const results = await new AxeBuilder({ page }).analyze()
    expect(results.violations.filter(({ impact }) => impact === 'serious' || impact === 'critical')).toEqual([])

    const originalPath = await page.locator('.chart-rx').getAttribute('d')
    await page.getByRole('link', { name: 'Last hour' }).click()
    await expect(page).toHaveURL(/range=1h/)
    await expect(page.getByRole('link', { name: 'Last hour' })).toHaveAttribute('aria-current', 'true')
    await expect(page.locator('#chart-description')).toHaveText(/Peak receive 630/)
    const selectedPath = await page.locator('.chart-rx').getAttribute('d')
    expect(selectedPath).not.toBe(originalPath)
    await page.reload()
    await expect(page.getByRole('link', { name: 'Last hour' })).toHaveAttribute('aria-current', 'true')
    await expect(page.locator('#chart-description')).toHaveText(/Peak receive 630/)
    await expect(page.locator('.chart-rx')).toHaveAttribute('d', selectedPath)
    if (viewport.width < 640) {
      await page.getByRole('navigation', { name: 'Page navigation' }).getByRole('link', { name: 'Links' }).click()
      await expect(page).toHaveURL(/#links$/)
    }
  })
}

test('fixture scenarios and chart samples are explicit', async ({ page }) => {
  await page.goto('/?range=7d&scenario=normal')
  await expect(page.getByRole('status')).toHaveCount(0)
  await expect(page.locator('#chart-description')).toHaveText(/Peak receive 720/)
  await page.getByText('View chart data').click()
  await expect(page.getByRole('cell', { name: '720 Mb/s' })).toBeVisible()
  await page.goto('/?scenario=unavailable')
  await expect(page.getByText('Throughput data unavailable')).toBeVisible()
  await expect(page.getByText('Link data unavailable')).toBeVisible()
  await expect(page.getByRole('img', { name: 'Throughput · WAN eth0' })).toHaveCount(0)
  const results = await new AxeBuilder({ page }).analyze()
  expect(results.violations.filter(({ impact }) => impact === 'serious' || impact === 'critical')).toEqual([])
})

test('keyboard skip link and in-page navigation work', async ({ page }) => {
  await page.goto('/')
  await page.keyboard.press('Tab')
  await expect(page.getByRole('link', { name: 'Skip to content' })).toBeFocused()
  await page.keyboard.press('Enter')
  await expect(page).toHaveURL(/#main$/)
  await page.getByRole('navigation', { name: 'Main navigation' }).getByRole('link', { name: 'Links' }).click()
  await expect(page).toHaveURL(/#links$/)
})
