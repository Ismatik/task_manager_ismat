// Nexus — the application shell.
//
// This is deliberately empty. S2-09 emptied the room: the Wails scaffold's
// `Greet` demo box is gone (S2-07 had already removed the Go method it called)
// and App.css went with it, along with the raw 3px corner radius it declared —
// a radius from nowhere, in a project whose radii are `rounded-sm/md/lg` and
// nothing else. The scaffold's orphaned webfont and its logo image went too:
// the fonts that stay are the vendored @fontsource packages imported by
// style.css, loaded from node_modules and never from a CDN.
//
// What is left is the page background, and it is drawn entirely through the
// design tokens: `bg-bg` and `text-ink` resolve to the CSS custom properties
// design/tokens.css declares per palette and per theme, so this shell is
// already correct under aurora/studio × light/dark without knowing any of the
// four values. There is no hex literal here and there never will be.
//
// Nothing here renders a user-visible string, so there is nothing to translate
// yet; the i18n runtime arrives in S2-11, the appearance runtime in S2-12, the
// Go client and the store in S2-13, and the first card in S2-14.
function App() {
  return <div className="min-h-screen bg-bg text-ink" />;
}

export default App;
