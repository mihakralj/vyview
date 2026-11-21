module.exports = {
  testDir: './tests',
  use: { browserName: 'chromium', headless: true, baseURL: 'http://127.0.0.1:18765' },
  webServer: {
    command: 'go run ./cmd/vyview -addr 127.0.0.1:18765',
    url: 'http://127.0.0.1:18765',
    reuseExistingServer: false,
    timeout: 120000,
  },
}
