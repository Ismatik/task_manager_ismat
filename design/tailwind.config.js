// Nexus — Tailwind config. All colors resolve through CSS variables from tokens.css,
// so palette (data-palette) and theme (.dark) switch without touching this file.
module.exports = {
  darkMode: 'class',
  theme: {
    extend: {
      colors: {
        bg: 'var(--bg)',
        surface: 'var(--surface)',
        elevated: 'var(--surface-elevated)',
        line: 'var(--border)',
        ink: 'var(--text)',
        muted: 'var(--text-muted)',
        accent: 'var(--accent)',
        'accent-2': 'var(--accent-2)',
        'on-accent': 'var(--on-accent)',
        danger: 'var(--danger)',
        warning: 'var(--warning)',
        success: 'var(--success)',
      },
      borderRadius: {
        sm: 'var(--radius-sm)',
        md: 'var(--radius-md)',
        lg: 'var(--radius-lg)',
      },
      boxShadow: {
        sm: 'var(--shadow-sm)',
        lg: 'var(--shadow-lg)',
      },
      fontFamily: {
        ui: 'var(--font-ui)',
        mono: 'var(--font-mono)',
      },
      backdropBlur: {
        glass: 'var(--blur)',
      },
      transitionDuration: {
        fast: 'var(--dur-fast, 120ms)',
        base: 'var(--dur-base, 150ms)',
        slow: 'var(--dur-slow, 200ms)',
      },
    },
  },
};
