import { GetSetupStatus } from '../wailsjs/go/main/App';
import { detectLocale, initializeLocalization, translate } from './i18n';

async function start(): Promise<void> {
  let language: string | undefined;
  try {
    language = (await GetSetupStatus()).language;
  } catch {
    // Keep the interface available so loadSetup can report the backend failure.
  }

  // The backend owns the preference. WebView storage may be unavailable or
  // cleared between launches, so it must not gate rendering or trigger reloads.
  initializeLocalization(detectLocale(language, navigator.languages));
  await import('./main');
}

void start().catch((error: unknown) => {
  console.error('Corsarr startup failed', error);
  const root = document.querySelector('#app');
  if (root) root.textContent = translate('setup.loadError');
});
