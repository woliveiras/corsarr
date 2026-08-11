import assert from 'node:assert/strict';
import test from 'node:test';

import {
  missingSelectedIntegrations,
  toggleApplicationSelection,
} from './application-selection.ts';

const applications = [
  { id: 'qbittorrent', dependencies: [] },
  { id: 'prowlarr', dependencies: [] },
  { id: 'radarr', dependencies: ['qbittorrent', 'prowlarr'] },
  { id: 'sonarr', dependencies: ['qbittorrent', 'prowlarr'] },
  { id: 'bazarr', dependencies: ['radarr', 'sonarr'] },
];

test('selecting an application also selects its recommended integrations', () => {
  assert.deepEqual(toggleApplicationSelection([], 'radarr', applications), [
    'prowlarr',
    'qbittorrent',
    'radarr',
  ]);
  assert.deepEqual(toggleApplicationSelection([], 'bazarr', applications), [
    'bazarr',
    'prowlarr',
    'qbittorrent',
    'radarr',
    'sonarr',
  ]);
});

test('deselecting an integration keeps its consumers selected', () => {
  const selection = toggleApplicationSelection(
    ['prowlarr', 'qbittorrent', 'radarr'],
    'qbittorrent',
    applications,
  );

  assert.deepEqual(selection, ['prowlarr', 'radarr']);
  assert.deepEqual(missingSelectedIntegrations(selection, applications), [
    { consumerID: 'radarr', integrationID: 'qbittorrent' },
  ]);
});

test('selecting another consumer restores a missing shared integration', () => {
  assert.deepEqual(toggleApplicationSelection(['prowlarr', 'radarr'], 'sonarr', applications), [
    'prowlarr',
    'qbittorrent',
    'radarr',
    'sonarr',
  ]);
});
