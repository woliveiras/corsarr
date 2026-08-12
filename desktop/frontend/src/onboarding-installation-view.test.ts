import assert from 'node:assert/strict';
import { readFileSync } from 'node:fs';
import test from 'node:test';

const mainSource = readFileSync(new URL('./main.ts', import.meta.url), 'utf8');

test('confirmation opens a dedicated installation view before awaiting the backend', () => {
  assert.match(mainSource, /id="onboarding-installation"/);
  assert.match(mainSource, /id="onboarding-installation-progress-list"/);
  assert.match(mainSource, /id="onboarding-installation-completion"/);
  assert.match(
    mainSource,
    /showOnboardingStep\('installation'\);[\s\S]*await InstallSelectedApplications\(\)/,
  );
});

test('installation view offers retry only after an interrupted attempt', () => {
  assert.match(mainSource, /id="onboarding-installation-retry"[^>]*hidden/);
  assert.match(mainSource, /onboardingInstallationRetryButton\.hidden = false/);
});
