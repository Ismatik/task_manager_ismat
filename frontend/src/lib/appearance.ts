// Nexus — the appearance runtime: palette, theme, accent and the document
// language, applied to the document exactly as design/README.md prescribes.
//
//     document.documentElement.dataset.palette = palette;
//     document.documentElement.classList.toggle('dark', isDark);
//     document.documentElement.style.setProperty('--accent', color);
//     document.documentElement.lang = language;
//
// # This module decides nothing and stores nothing
//
// The four values live in the `settings` table (D6) and are read and written
// through SettingsService. What is here is the DOM half: given a SettingsView,
// put the document into the state it describes. It holds no state of its own,
// which is what lets the store apply the service's answer — including after a
// refusal — by calling the same function again.
//
// # Why <html lang> is written HERE and nowhere else (K11, D23)
//
// index.html hard-codes lang="en" and, until S3-06, nothing ever updated it:
// the Russian UI was served as an English document, which is wrong for a screen
// reader, for hyphenation and for the font matching that K11 is about. The
// language is a persisted setting like the other three, this module already
// receives the whole SettingsView, already runs once at boot and once per
// settings write, and is already the only module in frontend/src that writes to
// documentElement. So it gets one more line and NOT a second module.
//
// The static lang="en" in index.html stays as the pre-boot default, exactly as
// data-palette="aurora" does. And the value is NOT derived from anything: both
// i18next and this line take the same SettingsView field, so neither can be a
// second opinion about what the language is.
//
// # No hex, ever
//
// Not one colour is named in this file. `palette` and `theme` select between
// declarations that live in design/tokens.css, and `accent` is whatever string
// the user chose, carried through to setProperty untouched. design/ is
// read-only and is the only place a colour is written down.

/**
 * The subset of SettingsView this module needs. It is structural on purpose:
 * the store passes the real service.SettingsView, and a test passes a literal,
 * and neither has to know about the other.
 */
export interface Appearance {
  palette: string;
  theme: string;
  accent: string;
  language: string;
}

// The three presentation mappings, each written down ONCE.
//
// The Tailwind config is `darkMode: 'class'` and tokens.css keys the dark
// variables off `.dark`, so the class name is fixed by design/ and not by us.
// That it is spelt the same as the `theme` value is a coincidence worth not
// relying on, which is why the two constants are separate.
const DARK_CLASS = 'dark';
const DARK_THEME = 'dark';
const AURORA_PALETTE = 'aurora';

/** The custom property design/tokens.css declares per palette and per theme. */
const ACCENT_PROPERTY = '--accent';

/**
 * `data-appearance="pending"` is set statically in index.html and removed here.
 * While it is present the inline style block in index.html keeps the body
 * hidden and the page background transparent, so the WebView shows the window
 * colour GTK already painted from `settings` (D12) instead of flashing whatever
 * the static palette/theme in the markup happens to be. It is the frontend half
 * of the flash D12 fixes on the GTK side.
 */
const BOOT_ATTRIBUTE = 'appearance';

/** Reports whether the user has asked the system for reduced motion. */
export function prefersReducedMotion(view: Window = window): boolean {
  // jsdom does not implement matchMedia, and neither did older WebKit builds
  // for every media feature. "Cannot tell" means "do not reduce": the default
  // is the full experience, and a test that wants the other answer stubs it.
  if (typeof view.matchMedia !== 'function') {
    return false;
  }
  return view.matchMedia('(prefers-reduced-motion: reduce)').matches;
}

/**
 * Reports whether Aurora's slow background drift may run.
 *
 * design/README.md: "Aurora's background drift (~60s) must also pause" under
 * `prefers-reduced-motion`. tokens.css already kills every CSS animation and
 * transition under that media query, but that is a blunt instrument and it
 * cannot see an animation driven from JavaScript. So the decision is made here,
 * once, and published on the document as `data-drift`; anything that draws the
 * drift — CSS keyframes or a requestAnimationFrame loop — is gated on that
 * attribute and therefore cannot disagree with this function.
 */
export function auroraDriftEnabled(palette: string, view: Window = window): boolean {
  return palette === AURORA_PALETTE && !prefersReducedMotion(view);
}

/**
 * Puts the document into the state `appearance` describes, and reveals it.
 *
 * Idempotent, and safe to call again with the service's answer after a refusal.
 */
export function applyAppearance(
  appearance: Appearance,
  view: Window = window,
  doc: Document = view.document,
): void {
  const root = doc.documentElement;

  root.dataset.palette = appearance.palette;
  root.classList.toggle(DARK_CLASS, appearance.theme === DARK_THEME);

  // The one writer of <html lang> in frontend/src (K11, D23). The value is
  // carried through untouched, like `accent`: which languages exist is
  // domain.Languages()' answer, and a mapping here would be a second one.
  root.lang = appearance.language;

  if (appearance.accent === '') {
    // Clearing must be CLEARING: D6 says "" means "use the palette's own
    // --accent", and the only way the palette gets its value back is for the
    // inline declaration to be GONE.
    //
    // Two things worth being precise about, because the obvious statement of
    // this rule is half wrong. CSSOM defines setProperty(p, "") as a call to
    // removeProperty(p), so those two are the same thing and swapping one for
    // the other is not the bug to worry about. The bug to worry about is an
    // `if (accent) setProperty(...)` with NO else, which silently leaves the
    // previous accent in place and makes clearing a no-op — and that a custom
    // property CAN hold an empty token stream, so anything that writes the
    // declaration through cssText or setAttribute ("--accent: ;") really does
    // shadow the palette with nothing. Hence: an explicit branch, and
    // removeProperty, which says what it means.
    root.style.removeProperty(ACCENT_PROPERTY);
  } else {
    root.style.setProperty(ACCENT_PROPERTY, appearance.accent);
  }

  root.dataset.drift = auroraDriftEnabled(appearance.palette, view) ? 'on' : 'off';

  revealAfterBoot(doc);
}

/**
 * Removes the boot gate, so the document paints.
 *
 * Called at the end of applyAppearance, and separately by the bootstrap when
 * the settings read itself failed — a Go error must not leave the user looking
 * at a permanently blank window.
 */
export function revealAfterBoot(doc: Document = document): void {
  delete doc.documentElement.dataset[BOOT_ATTRIBUTE];
}
