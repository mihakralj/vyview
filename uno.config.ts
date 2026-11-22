import { defineConfig, presetWind3 } from 'unocss'

export default defineConfig({
  content: {
    filesystem: ['testdata/tooling/**/*.html', 'templates/**/*.html', 'templates/**/*.templ', 'assets/js/**/*.js', 'internal/demo/templates/**/*.html'],
  },
  theme: {
    colors: Object.fromEntries([
      'bg', 'rail', 'surface', 'surface-2', 'border', 'text', 'text-muted',
      'accent', 'accent-2', 'ok', 'warn', 'crit', 'info', 'unknown',
      'chart-1', 'chart-2', 'chart-3', 'chart-4', 'chart-5', 'chart-6', 'chart-7', 'chart-8',
      'chart-9', 'chart-10', 'chart-11', 'chart-12', 'chart-13', 'chart-14', 'chart-15', 'chart-16',
      'chart-other', 'chart-hw', 'chart-sw', 'chart-grid', 'chart-axis', 'chart-crosshair',
      'heat-0', 'heat-1', 'heat-2', 'heat-3', 'heat-4', 'heat-5',
    ].map((token) => [token, `var(--${token})`])),
    fontFamily: {
      sans: '"IBM Plex Sans", system-ui, sans-serif',
      mono: '"IBM Plex Mono", ui-monospace, monospace',
    },
  },
  presets: [presetWind3()],
})
