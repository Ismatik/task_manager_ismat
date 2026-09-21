/// <reference types="vitest/config" />
import { defineConfig } from 'vite';
import react from '@vitejs/plugin-react';

// Nexus — the Vite config, and the vitest config with it.
//
// One config file rather than two on purpose: vitest reads `test` out of the
// Vite config, and a separate vitest.config.ts would have to restate the react
// plugin and every resolve decision made here. A restated setting is a setting
// that drifts.
//
// The runner exists because of the stage's ACCEPT criterion — "keyboard only,
// no mouse". That is an interaction property neither tsc nor ESLint can see,
// and a claim only a human at a keyboard can check is a claim the next reviewer
// cannot reproduce. jsdom plus @testing-library/user-event drives real key
// events through the real event pipeline, which turns it into an assertion.
// TASKS.md, "New dependencies", records why not Playwright: it downloads
// browser binaries, and this project is local-only.
export default defineConfig({
  plugins: [react()],
  test: {
    // jsdom, not happy-dom: it is what @testing-library is exercised against,
    // and focus/tabindex behaviour is exactly what S2-16 will be asserting.
    environment: 'jsdom',
    // No `globals: true`. Tests import describe/it/expect from 'vitest'
    // explicitly, so nothing depends on an ambient type that gate 4 would have
    // to be told about and ESLint would have to be told about separately.
    globals: false,
    setupFiles: ['./src/test/setup.ts'],
    // Test files live beside their subject — ARCHITECTURE.md section 6 defines
    // no parallel test tree — so they are covered by gate 3 and gate 4 like any
    // other source file.
    include: ['src/**/*.test.{ts,tsx}'],
    css: false,
  },
});
