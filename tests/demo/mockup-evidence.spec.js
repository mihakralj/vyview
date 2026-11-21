const { test, expect } = require('@playwright/test')
const AxeBuilder = require('@axe-core/playwright').default

test('example evidence keeps preferences and opens the corresponding inspector', async ({ page }) => {
  await page.goto('/clients?scenario=normal&range=7d&scoutTheme=light')
  const identity = page.getByRole('region', { name: 'Client identity evidence' })
  await identity.getByRole('link', { name: 'nas-01' }).click()
  await expect(page.getByRole('heading', { name: 'nas-01 · details' })).toBeVisible()
  await expect(page).toHaveURL(/scenario=normal/)
  await expect(page).toHaveURL(/range=7d/)
  await expect(page).toHaveURL(/scoutTheme=light/)
  const { violations } = await new AxeBuilder({ page }).analyze()
  expect(violations.filter(({ impact }) => impact === 'serious' || impact === 'critical')).toEqual([])
})

test('unavailable snapshots omit summaries and example evidence on mobile', async ({ page }) => {
  await page.setViewportSize({ width: 390, height: 844 })
  for (const path of ['/clients', '/firewall', '/conntrack', '/routing', '/vpn/wireguard', '/qos', '/dataplane', '/system']) {
    await page.goto(`${path}?scenario=unavailable`)
    await expect(page.getByRole('region', { name: 'Snapshot summary' })).toHaveCount(0)
    await expect(page.getByText('This fixture source is unavailable.')).toBeVisible()
    expect(await page.evaluate(() => document.documentElement.scrollWidth <= window.innerWidth)).toBe(true)
  }
})
