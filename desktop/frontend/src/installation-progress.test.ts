import assert from 'node:assert/strict';
import test from 'node:test';

import { applyInstallationProgress, createInstallationProgress } from './installation-progress.ts';

test('tracks every selected application from waiting through ready', () => {
  let items = createInstallationProgress(['radarr', 'qbittorrent', 'sonarr']);
  assert.deepEqual(
    items.map(({ applicationId, stage }) => ({ applicationId, stage })),
    [
      { applicationId: 'qbittorrent', stage: 'waiting' },
      { applicationId: 'radarr', stage: 'waiting' },
      { applicationId: 'sonarr', stage: 'waiting' },
    ],
  );

  items = applyInstallationProgress(items, {
    applicationId: 'qbittorrent',
    stage: 'installing',
    position: 1,
    total: 3,
  });
  items = applyInstallationProgress(items, {
    applicationId: 'qbittorrent',
    stage: 'provisioning',
    position: 1,
    total: 3,
  });
  items = applyInstallationProgress(items, {
    applicationId: 'qbittorrent',
    stage: 'ready',
    position: 1,
    total: 3,
  });

  assert.deepEqual(items[0], {
    applicationId: 'qbittorrent',
    position: 1,
    stage: 'ready',
  });
  assert.equal(items[1]?.stage, 'waiting');
});

test('keeps a failed application visible in its installation position', () => {
  const items = applyInstallationProgress(createInstallationProgress(['radarr']), {
    applicationId: 'radarr',
    stage: 'failed',
    position: 1,
    total: 1,
  });

  assert.deepEqual(items, [{ applicationId: 'radarr', position: 1, stage: 'failed' }]);
});
