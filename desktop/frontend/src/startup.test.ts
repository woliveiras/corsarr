import assert from 'node:assert/strict';
import { readFileSync } from 'node:fs';
import test from 'node:test';
import { fileURLToPath } from 'node:url';
import { Window } from 'happy-dom';
import { build } from 'vite';

// Execute the complete frontend, including module initialization and Wails bindings.
const html = readFileSync(new URL('../index.html', import.meta.url), 'utf8');
const entry = html.match(/<script type="module" src="\/([^"]+)"/);
assert.ok(entry);
const bundle = await build({
  configFile: false,
  root: fileURLToPath(new URL('..', import.meta.url)),
  logLevel: 'silent',
  build: {
    write: false,
    minify: false,
    lib: {
      entry: fileURLToPath(new URL(`../${entry[1]}`, import.meta.url)),
      name: 'CorsarrStartupTest',
      formats: ['iife'],
    },
    rollupOptions: { output: { inlineDynamicImports: true } },
  },
});
const output = Array.isArray(bundle) ? bundle[0] : bundle;
assert.ok('output' in output);
const script = output.output.find((item) => item.type === 'chunk');
assert.ok(script?.type === 'chunk');

test('Linux onboarding saves Engine selection, renews consent and preserves permission guidance', async () => {
  const window = new Window({ url: 'http://localhost/' });
  const schedule = window.setTimeout.bind(window);
  // Recovery polling is tested separately; keep this test finite while the
  // simulated daemon stays unavailable.
  window.setTimeout = (handler, delay, ...args) =>
    delay === 5_000 ? 0 : schedule(handler, delay, ...args);
  const saved: { kind: string; context?: string }[] = [];
  let configuration = { kind: 'docker-desktop', context: '' };
  const status = {
    language: 'pt-BR',
    applications: [],
    onboardingCompleted: false,
    onboardingStep: 'permissions',
  };
  try {
    window.go = {
      main: {
        App: new Proxy(
          {
            GetSetupStatus: async () => status,
            GetExecutionConfiguration: async () => configuration,
            SaveExecutionConfiguration: async (next: typeof configuration) => {
              saved.push(next);
              configuration = next;
              return configuration;
            },
            GetEnvironmentStatus: async () => ({
              platform: 'linux',
              host: { ready: true, issues: [] },
              runtime: { state: 'unavailable' },
            }),
            PrepareRuntime: async () => ({
              installed: true,
              ready: false,
              message: 'Ask your administrator to grant Docker socket access.',
            }),
          },
          { get: (target, key) => target[key] ?? (async () => []) },
        ),
      },
    };
    window.runtime = { EventsOnMultiple: () => () => {} };
    window.confirm = () => true;
    window.document.body.innerHTML = '<div id="app"></div>';
    window.eval(script.code);
    await window.happyDOM.waitUntilComplete();
    const select = window.document.querySelector('#execution-kind');
    const terms = window.document.querySelector('#onboarding-terms');
    assert.ok(select && terms);
    select.value = 'docker-engine';
    select.dispatchEvent(new window.Event('change'));
    assert.equal(terms.disabled, true, 'unsaved selection must not be authorized');
    window.document.querySelector('#execution-selection button').click();
    await window.happyDOM.waitUntilComplete();
    assert.equal(saved.at(-1)?.kind, 'docker-engine');
    assert.equal(terms.disabled, false);
    assert.equal(terms.checked, false);
    assert.ok(
      !window.document
        .querySelector('#onboarding-terms + span')
        .textContent.includes('Docker Desktop'),
    );
    window.document.querySelector('#onboarding-prepare-runtime').click();
    await window.happyDOM.waitUntilComplete();
    assert.ok(
      window.document
        .querySelector('#onboarding-environment-message')
        .textContent.includes('socket access'),
    );
  } finally {
    await window.happyDOM.close();
  }
});

for (const scenario of [
  { name: 'new onboarding', language: '', completed: false, storage: 'denied' },
  { name: 'an upgraded installation', language: 'it', completed: true, storage: 'denied' },
  {
    name: 'an installation with volatile storage',
    language: 'it',
    completed: true,
    storage: 'volatile',
  },
  {
    name: 'an installation with stale cached language',
    language: 'it',
    completed: true,
    storage: 'stale',
  },
  { name: 'a failed setup read', language: '', completed: false, storage: 'denied', fails: true },
]) {
  test(`renders ${scenario.name} without depending on WebView storage`, async () => {
    const window = new Window({ url: 'http://wails.localhost' });
    try {
      if (scenario.storage === 'denied') {
        window.eval(`Object.defineProperty(window, 'localStorage', {
          get() { throw new DOMException('Access denied', 'SecurityError'); }
        })`);
      } else if (scenario.storage === 'volatile') {
        window.eval(`Object.defineProperty(window, 'localStorage', {
          value: { getItem: () => null, setItem: () => {} }
        })`);
      } else {
        window.localStorage.setItem('corsarr.desktop.language', 'es');
      }
      Object.defineProperty(window.navigator, 'languages', { value: ['pt-BR'] });
      let reloads = 0;
      window.location.reload = () => {
        reloads++;
      };
      const savedLanguages: string[] = [];
      const status = {
        language: scenario.language,
        applications: [],
        onboardingCompleted: scenario.completed,
        onboardingStep: 'splash',
      };
      window.go = {
        main: {
          App: new Proxy(
            {
              GetSetupStatus: async () => {
                if (scenario.fails) throw new Error('Setup unavailable');
                return status;
              },
              GetEnvironmentStatus: async () => ({
                host: { supported: true, issues: [], platform: 'darwin', architecture: 'arm64' },
                runtime: { state: 'ready', provider: 'docker' },
              }),
              SetLanguagePreference: async (language: string) => {
                savedLanguages.push(language);
                status.language = language;
                return status;
              },
            },
            { get: (target, key) => target[key] ?? (async () => []) },
          ),
        },
      };
      window.runtime = { EventsOnMultiple: () => () => {} };
      window.document.body.innerHTML = '<div id="app"></div>';
      window.eval(script.code);
      await window.happyDOM.waitUntilComplete();

      assert.equal(window.document.documentElement.lang, scenario.language || 'pt-BR');
      assert.equal(reloads, 0, 'startup must not enter a locale synchronization reload loop');
      const visibleShell = window.document.querySelector(
        scenario.completed || scenario.fails ? '#dashboard-shell' : '#onboarding',
      );
      assert.ok(visibleShell);
      assert.equal(visibleShell.hidden, false);
      assert.ok(visibleShell.textContent.includes('Corsarr'));
      if (scenario.fails) {
        assert.ok(window.document.querySelector('#message.error')?.textContent);
      } else {
        assert.deepEqual(savedLanguages, scenario.language ? [] : ['pt-BR']);
        const languageSelect = window.document.querySelector('.language-select');
        assert.ok(languageSelect);
        languageSelect.value = 'es';
        languageSelect.dispatchEvent(new window.Event('change'));
        await window.happyDOM.waitUntilComplete();
        assert.equal(savedLanguages.at(-1), 'es');
        assert.equal(reloads, 1, 'a user language change reloads only after saving to the backend');
      }
    } finally {
      await window.happyDOM.close();
    }
  });
}
