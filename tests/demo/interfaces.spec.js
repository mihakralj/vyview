const { test, expect } = require('@playwright/test')
const AxeBuilder = require('@axe-core/playwright').default

test('interface hierarchy exposes physical strip, VLAN/PPPoE nesting, multi-IPv4/IPv6, VPN underlay, and saved columns', async ({ page, context }) => {
  await context.grantPermissions(['clipboard-read', 'clipboard-write'])
  await page.goto('/interfaces')
  await expect(page.getByText('11 interfaces · 9 up · 2 down')).toBeVisible()
  await expect(page.getByRole('link', { name: 'Ethernet', exact: true })).toContainText('5')

  // Physical front-panel port strip
  const ports = page.getByRole('list', { name: 'Physical router ports' })
  await expect(ports.getByRole('listitem')).toHaveCount(5)
  await expect(ports.getByRole('link', { name: /eth0/ })).toContainText('1G · u/u')
  await expect(ports.getByRole('link', { name: /eth1/ })).toContainText('No carrier · u/D')
  await expect(ports.getByRole('link', { name: /eth2/ })).toContainText('No carrier · u/D')
  await expect(ports.getByRole('link', { name: /eth3/ })).toContainText('10G · u/u')
  await expect(ports.getByRole('link', { name: /eth4/ })).toContainText('10G · u/u')

  // Unified hierarchy tree-table
  const table = page.getByRole('region', { name: 'Record list' })
  await expect(table.locator('tbody tr.iface-row')).toHaveCount(11)
  await expect(table.locator('tbody tr.iface-group-row')).toHaveCount(3)

  // eth0 has MAC + IPv4 (DHCP) + IPv6 GUA + IPv6 ULA + IPv6 LLA in canonical order
  const eth0Row = table.locator('tr.iface-row', { has: page.getByRole('link', { name: 'eth0', exact: true }) })
  await expect(eth0Row).toContainText('192.168.1.185/16')
  await expect(eth0Row).toContainText('DHCP · 22h')
  await expect(eth0Row).toContainText('2001:db8:1::185/64')
  await expect(eth0Row).toContainText('fd00:1::185/64')
  await expect(eth0Row).toContainText('fe80::eaf6:d7ff:fe00:1600/64')
  await expect(eth0Row.locator('.scope-tag')).toHaveText(['MAC', 'IPv4', 'IPv6 GUA', 'IPv6 ULA', 'IPv6 LLA'])

  // eth1 & eth2 down with no carrier
  await expect(table.locator('tr.iface-row', { has: page.getByRole('link', { name: 'eth1', exact: true }) })).toContainText('No carrier')
  await expect(table.locator('tr.iface-row', { has: page.getByRole('link', { name: 'eth2', exact: true }) })).toContainText('No carrier')

  // eth3 -> eth3.10 -> pppoe10 and eth4 -> eth4.20 nesting + directional centerline warnings + P2P peer
  await expect(table.locator('tr.iface-row', { has: page.getByRole('link', { name: 'eth3', exact: true }) }).getByRole('link', { name: '2 rx err' })).toHaveClass(/interface-warning/)
  await expect(table.locator('tr.iface-row', { has: page.getByRole('link', { name: 'eth4', exact: true }) }).getByRole('link', { name: '12 rx drop' })).toHaveClass(/interface-warning/)
  await expect(table.locator('tr.iface-row', { has: page.getByRole('link', { name: 'pppoe10', exact: true }) })).toContainText('10.99.50.16')
  await expect(table.locator('tr.iface-row', { has: page.getByRole('link', { name: 'pppoe10', exact: true }) })).toContainText('10.99.50.1/32')
  await expect(table.locator('tr.iface-row', { has: page.getByRole('link', { name: 'wg0', exact: true }) })).toContainText('via pppoe10')

  await expect(table.getByRole('img', { name: 'eth0 example 1-hour trend' })).toBeVisible()
  await table.getByRole('button', { name: 'Copy eth0 address' }).click()
  await expect(page.locator('[data-copy-status]').last()).toHaveText('Copied.')
  expect(await page.evaluate(() => navigator.clipboard.readText())).toBe('192.168.1.185/16')

  // Flat view toggle
  await page.getByRole('navigation', { name: 'Interface layout' }).getByRole('link', { name: 'Flat table' }).click()
  await expect(table.locator('tbody tr.iface-group-row')).toHaveCount(0)
  await expect(table.locator('tbody tr.iface-row')).toHaveCount(11)

  await expect(eth0Row).toContainText('MTU 1500')
  await expect(table.locator('tr.iface-row', { has: page.getByRole('link', { name: 'pppoe10', exact: true }) })).toContainText('MTU 1492')

  // Column chooser persists across reload
  await page.getByText('Choose columns').click()
  await page.getByRole('checkbox', { name: 'VRF' }).uncheck()
  await expect(table.getByRole('columnheader', { name: 'VRF' })).toHaveCount(0)
  await page.reload()
  await expect(table.getByRole('columnheader', { name: 'VRF' })).toHaveCount(0)
  await page.getByText('Choose columns').click()
  await page.getByRole('checkbox', { name: 'VRF' }).check()
  await expect(table.getByRole('columnheader', { name: 'VRF' })).toBeVisible()

  const { violations } = await new AxeBuilder({ page }).analyze()
  expect(violations.filter(({ impact }) => impact === 'serious' || impact === 'critical')).toEqual([])
})

test('eth3 10G SFP+ detail provides mirrored series, optical ranges and complete tabs while eth0 RJ45 omits optics', async ({ page }) => {
  await page.goto('/interfaces/eth3?range=1h')
  await expect(page.getByRole('heading', { name: 'eth3', exact: true })).toBeVisible()
  await expect(page.getByText(/driver fsl_dpaa_eth/)).toBeVisible()
  await expect(page.locator('.interface-kpis .kpi')).toHaveCount(6)
  await expect(page.locator('.chart-rx').first()).toHaveAttribute('d', /M /)
  await expect(page.locator('.chart-tx').first()).toHaveAttribute('d', /M /)
  await expect(page.locator('.event-marker')).toHaveAttribute('d', 'M 400 20 V 200')
  await page.getByRole('navigation', { name: 'Chart range' }).getByRole('link', { name: '6h' }).click()
  await expect(page).toHaveURL(/range=6h/)
  await expect(page.getByRole('heading', { name: 'Optics · SFP+ DOM' })).toBeVisible()
  await expect(page.getByText('Warning limits: -9.9 to 0 dBm · alarm limits: -14 to 2 dBm')).toBeVisible()
  await expect(page.getByRole('region', { name: 'Driver counters' }).getByText('bp_count[0]')).toBeVisible()
  const { violations } = await new AxeBuilder({ page }).analyze()
  expect(violations.filter(({ impact }) => impact === 'serious' || impact === 'critical')).toEqual([])
  for (const tab of ['Counters', 'Optics', 'Addresses', 'Queues', 'Config']) {
    await page.getByRole('navigation', { name: 'Section views' }).getByRole('link', { name: tab }).click()
    await expect(page.getByRole('navigation', { name: 'Section views' }).getByRole('link', { name: tab })).toHaveAttribute('aria-current', 'page')
    await expect(page.getByRole('region', { name: 'Record list' })).toBeVisible()
  }

  // eth0 is 1G RJ45: shows IPv4 + GUA + ULA addresses and hides SFP+ optics
  await page.goto('/interfaces/eth0')
  await expect(page.getByRole('heading', { name: 'Optics · SFP+ DOM' })).toHaveCount(0)
  await expect(page.getByText('192.168.1.185/16')).toBeVisible()
  await expect(page.getByText('2001:db8:1::185/64')).toBeVisible()
  await expect(page.getByText('fd00:1::185/64')).toBeVisible()
})

test('all fixture interfaces have detail routes and unavailable state omits samples', async ({ page }) => {
  await page.goto('/interfaces')
  const links = await page.getByRole('region', { name: 'Record list' }).locator('th[scope="row"] a.iface-link').evaluateAll((nodes) => nodes.map((node) => node.href))
  expect(links.length).toBe(11)
  for (const href of links) expect((await page.goto(href)).status(), href).toBe(200)
  await page.goto('/interfaces/eth3?tab=optics')
  await expect(page.getByRole('link', { name: 'Receive power' })).toBeVisible()
  await page.setViewportSize({ width: 390, height: 844 })
  await page.goto('/interfaces/eth3')
  expect(await page.evaluate(() => document.documentElement.scrollWidth <= window.innerWidth)).toBe(true)
  await page.goto('/interfaces/eth3?scenario=unavailable')
  await expect(page.getByText('This fixture source is unavailable.')).toBeVisible()
  await expect(page.locator('.interface-kpis')).toHaveCount(0)
  await expect(page.getByRole('heading', { name: 'Optics · SFP+ DOM' })).toHaveCount(0)
})
