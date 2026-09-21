import React from 'react';
import { createRoot } from 'react-dom/client';
import { I18nextProvider } from 'react-i18next';

import './style.css';
import App from './App';
import { revealAfterBoot } from './lib/appearance';
import { wailsClient } from './lib/client';
import { createI18n, fallbackLanguage } from './lib/i18n';
import { createAppStore } from './store';

// Nexus — the entry point, and the bootstrap order that matters.
//
//   1. read `settings` from Go and apply palette, theme and accent;
//   2. build i18next in the language those settings name;
//   3. only then render.
//
// Nothing paints before step 3. index.html holds the document behind
// `data-appearance="pending"` until step 1 removes it, so the WebView never
// shows a theme the user did not choose — the frontend half of the startup
// flash D12 fixes on the GTK side. And i18next is initialised before the first
// render, so no frame shows raw keys.
//
// If Go cannot be reached the boot gate still comes off and the app still
// mounts, on the markup's static default and the fallback language, with a
// toast. A failed read is a bad session, not a blank window.
async function main() {
  const store = createAppStore(wailsClient);

  const settings = await store.getState().loadSettings();

  const i18n = await createI18n(settings?.language ?? fallbackLanguage);

  revealAfterBoot();

  const container = document.getElementById('root');
  const root = createRoot(container!);

  root.render(
    <React.StrictMode>
      <I18nextProvider i18n={i18n}>
        <App />
      </I18nextProvider>
    </React.StrictMode>,
  );
}

void main();
