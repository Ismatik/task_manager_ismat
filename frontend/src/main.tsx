import React from 'react';
import { createRoot } from 'react-dom/client';
import { I18nextProvider } from 'react-i18next';

import './style.css';
import App from './App';
import { createI18n, fallbackLanguage } from './lib/i18n';

// Nexus — the entry point.
//
// i18next is created and initialised BEFORE the first render, and the instance
// is handed down through I18nextProvider rather than installed as a global
// singleton. Rendering first and translating afterwards is how an app ends up
// showing a flash of raw keys, and the Provider is what lets a test mount the
// same tree in either language without the two interfering.
//
// The language starts at the fallback and is corrected from `settings` once
// there is a store to read Go through: `lib/client.ts` is the only file in
// frontend/src permitted to import from wailsjs, so the call belongs there and
// not here. `lib/i18n.ts` already takes the persistence as an injected port for
// exactly that reason.
async function main() {
  const i18n = await createI18n(fallbackLanguage);

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
