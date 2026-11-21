const railToggle = document.querySelector('[data-rail-toggle]')
if (railToggle) {
  const railKey = 'vyview-demo-rail'
  const syncRail = () => {
    const collapsed = document.documentElement.getAttribute('data-rail') === 'collapsed'
    railToggle.setAttribute('aria-expanded', collapsed ? 'false' : 'true')
  }
  const toggleRail = () => {
    const next = document.documentElement.getAttribute('data-rail') === 'collapsed' ? 'expanded' : 'collapsed'
    if (next === 'collapsed') document.documentElement.setAttribute('data-rail', 'collapsed')
    else document.documentElement.removeAttribute('data-rail')
    syncRail()
    try {
      localStorage.setItem(railKey, next)
    } catch {}
  }
  syncRail()
  railToggle.addEventListener('click', toggleRail)
  document.addEventListener('keydown', (event) => {
    if (event.ctrlKey || event.metaKey || event.altKey || event.key !== '[') return
    if (event.target.closest('input, textarea, select, [contenteditable]')) return
    event.preventDefault()
    toggleRail()
  })
}

document.addEventListener('keydown', (event) => {
  if ((!event.ctrlKey && !event.metaKey) || event.altKey || event.key.toLowerCase() !== 'k') return
  const search = document.querySelector('form[role="search"] input[type="search"]')
  const link = document.querySelector('[data-search-link]')
  if (!search && !link) return
  event.preventDefault()
  if (search) search.focus()
  else window.location.assign(link.href)
})

document.addEventListener('click', async (event) => {
  const button = event.target.closest('button[data-copy]')
  if (!button) return
  const status = button.closest('.inspector, section').querySelector('[data-copy-status]')
  try {
    await navigator.clipboard.writeText(button.dataset.copy)
    status.textContent = 'Copied.'
  } catch {
    status.textContent = 'Clipboard unavailable. Select the value and copy it manually.'
  }
})

const columnChooser = document.querySelector('[data-columns]')
if (columnChooser) {
  const controls = Array.from(columnChooser.querySelectorAll('[data-column-choice]'))
  const status = columnChooser.querySelector('[data-columns-status]')
  const key = 'vyview-demo-interface-columns'
  const apply = () => {
    for (const control of controls) {
      document.querySelectorAll(`[data-column="${control.dataset.columnChoice}"]`).forEach((cell) => {
        cell.hidden = !control.checked
      })
    }
  }
  try {
    const saved = localStorage.getItem(key)
    if (saved !== null) {
      const selected = JSON.parse(saved)
      if (!Array.isArray(selected) || selected.some((item) => typeof item !== 'string')) throw new Error('Invalid column preferences')
      controls.forEach((control) => { control.checked = selected.includes(control.dataset.columnChoice) })
    }
  } catch {
    status.textContent = 'Could not load saved columns; defaults are shown.'
  }
  apply()
  controls.forEach((control) => control.addEventListener('change', () => {
    apply()
    try {
      localStorage.setItem(key, JSON.stringify(controls.filter((item) => item.checked).map((item) => item.dataset.columnChoice)))
      status.textContent = 'Column preferences saved in this browser.'
    } catch {
      status.textContent = 'Could not save column preferences in this browser.'
    }
  }))
}

const replay = document.querySelector('[data-replay]')
if (replay) {
  const play = replay.querySelector('[data-replay-play]')
  const reset = replay.querySelector('[data-replay-reset]')
  const status = replay.querySelector('[data-replay-status]')
  const entries = Array.from(replay.querySelectorAll('li')).reverse()
  entries[0].parentElement.append(...entries)
  let index = 0
  let timer = null
  const stop = (message) => {
    window.clearInterval(timer)
    timer = null
    status.textContent = message
    play.textContent = 'Play fixture'
  }
  const advance = () => {
    entries[index++].hidden = false
    if (index === entries.length) {
      stop(`Replay complete. ${index} entries shown.`)
      play.disabled = true
    } else {
      status.textContent = `${index} of ${entries.length} entries shown.`
    }
  }
  play.addEventListener('click', () => {
    if (timer !== null) {
      stop(`Replay paused. ${index} of ${entries.length} entries shown.`)
      return
    }
    play.textContent = 'Pause replay'
    timer = window.setInterval(advance, 1500)
    advance()
  })
  reset.addEventListener('click', () => {
    stop('Replay is paused.')
    index = 0
    play.disabled = false
    entries.forEach((entry) => { entry.hidden = true })
  })
}
