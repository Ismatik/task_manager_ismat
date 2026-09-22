// Nexus — the one Node built-in the test suite reads source files with.
//
// # Why this file exists rather than `@types/node`
//
// `App.layout.test.tsx` has to read `src/style.css` AS TEXT, because that is
// where the height chain D22 specifies is written down and there is no other way
// to assert it exists. Three obvious routes are all closed:
//
//   * `import css from './style.css?raw'` resolves to the EMPTY STRING, because
//     vite.config.ts sets `test.css: false` and vitest's `vitest:css-disable`
//     plugin replaces the transform result of anything matching `\.css(?:$|\?)`
//     — `?raw` included. It fails SILENTLY, which would have made every
//     assertion over it pass while checking nothing.
//   * `new URL('./style.css', import.meta.url)` yields an `http:` URL under the
//     jsdom environment, which `readFileSync` rejects.
//   * `@types/node` would be a new dev dependency, and this project's rule is
//     that a dependency is justified in a ticket or not taken. One test reading
//     one file does not justify the whole Node typings surface, most of which
//     has no business being reachable from a browser bundle.
//
// So exactly the one function that is actually called is declared, with exactly
// the one signature it is called with. If the call drifts, tsc says so — which
// is the property `@types/node` would have given us and the only one wanted.
//
// This is TEST-ONLY. Nothing under components/, views/, lib/ or store/ may
// import a Node built-in: those files run in a WebView.

declare module 'node:fs' {
  /** Reads a file, relative to the process's working directory when not absolute. */
  export function readFileSync(path: string, encoding: 'utf8'): string;
}
