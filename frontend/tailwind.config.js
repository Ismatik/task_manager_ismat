// Nexus — frontend Tailwind config.
//
// The token map (colours, radii, shadows, fonts, blur, durations) is owned by the
// design handoff and is CONSUMED here, never restated: `design/` is read-only and
// `design/tailwind.config.js` is its single source of truth. This file adds only the
// `content` globs, which are a build concern the handoff cannot know about.
//
// Tailwind is pinned to v3 on purpose: the design config is a CommonJS v3 config
// using `darkMode: 'class'` + `theme.extend`, a shape Tailwind v4 no longer reads.
import designConfig from '../design/tailwind.config.js';

/** @type {import('tailwindcss').Config} */
export default {
  ...designConfig,
  content: ['./index.html', './src/**/*.{js,jsx,ts,tsx}'],
};
