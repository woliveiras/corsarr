import assert from 'node:assert/strict';
import { readFileSync } from 'node:fs';
import test from 'node:test';

const mainSource = readFileSync(new URL('./main.ts', import.meta.url), 'utf8');

test('runtime storage denial returns onboarding to folder selection', () => {
  assert.match(mainSource, /failed\?\.issue\?\.code === 'runtime_storage_access_denied'/);
  assert.match(mainSource, /showOnboardingStep\('storage'\)/);
  assert.match(mainSource, /onboardingStorageMessage\.textContent = issue\.nextAction/);
  assert.match(
    mainSource,
    /showOnboardingStep\('storage'\);[\s\S]*onboardingInstallationRetryButton\.hidden = true/,
  );
});
