# Nexus design tokens — handoff

Files:
- `tokens.css` — all color/radius/shadow/font tokens. Two palettes × two themes.
- `tailwind.config.js` — maps Tailwind color names to the CSS variables.

## Wiring (React + Tailwind + Wails)

```html
<html data-palette="aurora" class="dark">
```

```js
// palette toggle (aurora | studio)
document.documentElement.dataset.palette = palette;
// theme toggle (light | dark)
document.documentElement.classList.toggle('dark', isDark);
// user accent picker
document.documentElement.style.setProperty('--accent', color);
```

Persist all three in app settings; defaults: `aurora` + `dark` (studio reads best as `light`).

## Fonts
- Aurora: Space Grotesk (UI) + JetBrains Mono (numbers, dates, shortcuts, timers)
- Studio: Figtree (UI) + JetBrains Mono
Load both UI fonts; `--font-ui` picks the right one per palette. Base UI size 13px, 8px grid.

## Semantics
- `accent` — active view, running timer, focus rings, primary buttons (`on-accent` for text on it)
- `danger` — overdue badges, P0/P1 chips, bug icon
- `warning` — P2, streak flames · `success` — done, habit checks, task icon
- Aurora surfaces are translucent — pair `bg-surface` with `backdrop-blur-glass`; Studio surfaces are solid + `shadow-sm`.

## Motion
Interactions ≤ 200ms (`duration-fast/base/slow`). `tokens.css` kills all motion under `prefers-reduced-motion`. Aurora's background drift (~60s) must also pause there.
