const { test, expect } = require('@playwright/test')

test('IBM Plex fonts load locally for text, labels, and monospaced values', async ({ page }) => {
  const fontResponses = []
  page.on('response', (response) => {
    if (response.url().includes('/assets/fonts/')) fontResponses.push(response)
  })
  await page.goto('/')
  await page.evaluate(() => document.fonts.ready)
  await expect(page.locator('body')).toHaveCSS('font-family', /IBM Plex Sans/)
  await expect(page.locator('.wan-address')).toHaveCSS('font-family', /IBM Plex Mono/)
  await expect(page.locator('.axis-labels')).toHaveCSS('font-family', /IBM Plex Sans/)
  const fonts = await page.evaluate(async () => {
    const descriptors = [
      '400 14px "IBM Plex Sans"',
      '600 14px "IBM Plex Sans"',
      '700 14px "IBM Plex Sans"',
      '400 14px "IBM Plex Mono"',
      '700 14px "IBM Plex Mono"',
    ]
    return Promise.all(descriptors.map(async (descriptor) => {
      const loaded = await document.fonts.load(descriptor, 'Test 123')
      return loaded.length > 0 && loaded.every((font) => font.status === 'loaded')
    }))
  })
  expect(fonts).toEqual([true, true, true, true, true])
  expect(fontResponses.length).toBeGreaterThanOrEqual(5)
  expect(fontResponses.every((response) => response.ok() && new URL(response.url()).host === '127.0.0.1:18765')).toBe(true)
})
