import assert from 'node:assert/strict';
import { readFileSync } from 'node:fs';
import test from 'node:test';

const source = readFileSync(new URL('./main.ts', import.meta.url), 'utf8');

test('installation failures expose a native support report copy action', () => {
  assert.match(source, /id="onboarding-copy-support-report"/);
  assert.match(source, /t\('onboarding\.copyErrors'\)/);
  assert.match(source, /await CopyLastInstallationSupportReport\(\)/);
  assert.match(source, /showOnboardingNotification\(t\('onboarding\.logCopied'\), 'success'\)/);
  assert.match(source, /onboardingCopySupportReport\.hidden = false/);
});

test('retrying a finalization failure immediately restores visible progress', () => {
  assert.match(
    source,
    /onboardingInstallationProgress\.every\(\(\{ stage \}\) => stage === 'ready'\)/,
  );
  assert.match(source, /onboardingInstallationCompletionStage = 'active'/);
  assert.match(source, /t\('onboarding\.finalizeAgain'\)/);
});
