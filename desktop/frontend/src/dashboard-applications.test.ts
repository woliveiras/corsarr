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
