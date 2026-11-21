#!/usr/bin/env node
const { execFileSync } = require('node:child_process')
const fs = require('node:fs')
const path = require('node:path')

const root = path.resolve(__dirname, '..')
const mode = process.argv[2] || 'free'

// Symbol ID -> { pro: FA Pro icon name, free: FA Free Solid icon name }
const manifest = [
  { id: 'fa-router', pro: 'router', free: 'hard-drive' },
  { id: 'fa-gauge-max', pro: 'gauge-max', free: 'gauge-high' },
  { id: 'fa-wave-pulse', pro: 'wave-pulse', free: 'wave-square' },
  { id: 'fa-ethernet', pro: 'ethernet', free: 'ethernet' },
  { id: 'fa-chart-network', pro: 'chart-network', free: 'network-wired' },
  { id: 'fa-route', pro: 'route', free: 'route' },
  { id: 'fa-block-brick-fire', pro: 'block-brick-fire', free: 'shield-halved' },
  { id: 'fa-arrows-left-right', pro: 'arrows-left-right', free: 'arrows-left-right' },
  { id: 'fa-lock-keyhole', pro: 'lock-keyhole', free: 'lock' },
  { id: 'fa-sliders', pro: 'sliders', free: 'sliders' },
  { id: 'fa-diagram-project', pro: 'diagram-project', free: 'diagram-project' },
  { id: 'fa-server', pro: 'server', free: 'server' },
  { id: 'fa-microchip', pro: 'microchip', free: 'microchip' },
  { id: 'fa-code-compare', pro: 'code-compare', free: 'code-compare' },
  { id: 'fa-list-timeline', pro: 'list-timeline', free: 'list' },
  { id: 'fa-radar', pro: 'radar', free: 'stethoscope' },
  { id: 'fa-gear', pro: 'gear', free: 'gear' },
  { id: 'fa-link', pro: 'link', free: 'link' },
  { id: 'fa-angles-left', pro: 'angles-left', free: 'angles-left' },
  { id: 'fa-angles-right', pro: 'angles-right', free: 'angles-right' },
]

function buildFree() {
  const lines = [
    '{{define "icons-sprite"}}',
    '<!-- Font Awesome Free 7.3.1 by @fontawesome - https://fontawesome.com License - https://fontawesome.com/license/free (Icons: CC BY 4.0) -->',
    '<svg class="icon-sprite" aria-hidden="true">',
  ]
  for (const item of manifest) {
    const svgPath = path.join(root, 'node_modules/@fortawesome/fontawesome-free/svgs/solid', `${item.free}.svg`)
    const raw = fs.readFileSync(svgPath, 'utf8')
    const vb = raw.match(/viewBox="([^"]+)"/)[1]
    const d = raw.match(/<path[^>]*d="([^"]+)"/)[1]
    lines.push(`  <symbol id="${item.id}" viewBox="${vb}"><path fill="currentColor" d="${d}"/></symbol>`)
  }
  lines.push('</svg>', '{{end}}', '')
  const out = path.join(root, 'internal/demo/templates/icons_free.html')
  fs.writeFileSync(out, lines.join('\n'))
  console.log(`Wrote ${out} (${manifest.length} Free fallback symbols)`)
}

function buildPro() {
  const lines = [
    '{{define "icons-sprite"}}',
    '<svg class="icon-sprite" aria-hidden="true">',
  ]
  const env = { ...process.env, PATH: `${process.env.HOME}/.local/bin:${process.env.HOME}/bin:${process.env.PATH}` }
  for (const item of manifest) {
    const stdout = execFileSync('fa', ['icons', '--version', '7.3.1', '--name', item.pro, '--svg-format', 'html'], {
      encoding: 'utf8',
      env,
    })
    const parsed = JSON.parse(stdout)
    const svgs = parsed?.data?.release?.icon?.svgs || []
    const sharpLight = svgs.find((s) => s.familyStyle.family === 'sharp' && s.familyStyle.style === 'light')
    if (!sharpLight) {
      throw new Error(`Missing sharp light variant for ${item.pro}`)
    }
    const match = sharpLight.html.match(/<svg[^>]*viewBox="([^"]+)"[^>]*>([\s\S]*?)<\/svg>/)
    lines.push(`  <symbol id="${item.id}" viewBox="${match[1]}">${match[2]}</symbol>`)
  }
  lines.push('</svg>', '{{end}}', '')
  const outDir = path.join(root, 'internal/demo/pro_templates')
  fs.mkdirSync(outDir, { recursive: true })
  const out = path.join(outDir, 'icons_pro.html')
  fs.writeFileSync(out, lines.join('\n'))
  console.log(`Wrote ${out} (${manifest.length} Pro Sharp Light symbols; gitignored)`)
}

if (mode === 'pro') {
  buildPro()
} else if (mode === 'both') {
  buildFree()
  buildPro()
} else {
  buildFree()
}
