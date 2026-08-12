import assert from 'node:assert/strict';
import { readFileSync } from 'node:fs';
import test from 'node:test';

import {
  runningServicesSummary,
  sortApplicationsByInstallation,
} from './dashboard-applications.ts';

test('summarizes running services against all installed services', () => {
  assert.deepEqual(
    runningServicesSummary([
      { applicationId: 'qbittorrent', state: 'running' },
      { applicationId: 'prowlarr', state: 'running' },
      { applicationId: 'radarr', state: 'stopped' },
      { applicationId: 'sonarr', state: 'attention' },
      { applicationId: 'jellyfin', state: 'not_installed' },
    ]),
    { installed: 4, running: 2 },
  );
});

test('sorts installed applications before unavailable applications while preserving catalog order', () => {
  const applications = [
    { id: 'lazylibrarian' },
    { id: 'qbittorrent' },
    { id: 'radarr' },
    { id: 'jellyfin' },
    { id: 'bazarr' },
  ];
  const statuses = [
    { applicationId: 'qbittorrent', state: 'running' },
    { applicationId: 'radarr', state: 'stopped' },
    { applicationId: 'bazarr', state: 'attention' },
  ];

  assert.deepEqual(
    sortApplicationsByInstallation(applications, statuses).map(({ id }) => id),
    ['qbittorrent', 'radarr', 'bazarr', 'lazylibrarian', 'jellyfin'],
  );
});

test('dashboard no longer renders the obsolete preparation review', () => {
  const source = readFileSync(new URL('./main.ts', import.meta.url), 'utf8');

  assert.equal(source.includes('Revise sua preparação'), false);
  assert.equal(source.includes('class="installation-review"'), false);
});

test('dashboard does not repeat onboarding guidance in the sidebar', () => {
  const source = readFileSync(new URL('./main.ts', import.meta.url), 'utf8');

  assert.equal(source.includes('class="sidebar-note"'), false);
  assert.equal(source.includes('Tudo em um só lugar'), false);
});

test('dashboard does not repeat application selection guidance from onboarding', () => {
  const source = readFileSync(new URL('./main.ts', import.meta.url), 'utf8');

  assert.equal(source.includes('Não sabe por onde começar?'), false);
  assert.equal(source.includes('id="select-recommended"'), false);
});

test('dashboard installs unavailable applications instead of presenting onboarding selection', () => {
  const source = readFileSync(new URL('./main.ts', import.meta.url), 'utf8');
  const cardSource = source.slice(
    source.indexOf('function createApplicationCard'),
    source.indexOf('function updateApplicationButton'),
  );

  assert.match(cardSource, /'Instalar'/);
  assert.doesNotMatch(cardSource, /'Selecionar'/);
  assert.match(cardSource, /await InstallSelectedApplications\(\)/);
});

test('dashboard exposes managed Arr credentials only through the native clipboard bridge', () => {
  const source = readFileSync(new URL('./main.ts', import.meta.url), 'utf8');

  assert.equal(source.includes('GetARRAccessStatuses'), true);
  assert.equal(source.includes('CopyARRPassword(target.id)'), true);
  assert.equal(source.includes('arrAccesses.get(application.id)?.available'), true);
});

test('managed credentials keep the username visible and confirm password copy locally', () => {
  const source = readFileSync(new URL('./main.ts', import.meta.url), 'utf8');

  assert.match(source, /className = 'credential-access'/);
  assert.match(source, /usernameLabel\.textContent = 'Usuário'/);
  assert.match(source, /button\.textContent = '✓ Senha copiada'/);
  assert.doesNotMatch(source, /Use o usuário .* para entrar/);
});

test('installed and removal actions have explicit status semantics', () => {
  const source = readFileSync(new URL('./main.ts', import.meta.url), 'utf8');

  assert.match(source, /'✓ Instalado'/);
  assert.match(source, /classList\.add\('installed-status'\)/);
  assert.match(source, /classList\.add\('danger-button'\)/);
  assert.match(source, /Para remover \$\{target\.name\}, remova primeiro/);
});

test('application cards keep operational metadata out of the description area', () => {
  const source = readFileSync(new URL('./main.ts', import.meta.url), 'utf8');
  const cardSource = source.slice(
    source.indexOf('function createApplicationCard'),
    source.indexOf('function updateApplicationButton'),
  );

  assert.doesNotMatch(cardSource, /metadata\.className = 'metadata'/);
  assert.doesNotMatch(cardSource, /Remova primeiro/);
});

test('application cards render bundled logos with an initials fallback', () => {
  const source = readFileSync(new URL('./main.ts', import.meta.url), 'utf8');

  assert.match(source, /const applicationIconURLs: Record<string, string>/);
  assert.match(source, /image\.className = 'application-logo'/);
  assert.match(source, /image\.addEventListener\('error'/);
  assert.match(source, /createApplicationIcon\(application\)/);
  assert.match(source, /createApplicationIcon\(target\)/);
});

test('dashboard cards center a larger logo above complete application copy', () => {
  const styles = readFileSync(new URL('./style.css', import.meta.url), 'utf8');

  assert.match(
    styles,
    /grid-template-areas:\s*"\. actions"\s*"identity actions"\s*"information actions"\s*"\. actions"/,
  );
  assert.match(styles, /grid-template-rows: 1fr auto auto 1fr;/);
  assert.match(
    styles,
    /\.application-card > \.application-icon[^}]*width: 76px;[^}]*height: 76px;/,
  );
  assert.match(styles, /\.application-card \.application-logo[^}]*width: 58px;[^}]*height: 58px;/);
  assert.match(styles, /\.application-card \.application-info p[^}]*white-space: normal;/);
  assert.doesNotMatch(
    styles,
    /\.application-card \.application-info p[^}]*text-overflow: ellipsis;/,
  );
});
