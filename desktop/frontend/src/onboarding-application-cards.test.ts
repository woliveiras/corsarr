import assert from 'node:assert/strict';
import { readFileSync } from 'node:fs';
import test from 'node:test';

const source = readFileSync(new URL('./main.ts', import.meta.url), 'utf8');

test('onboarding application cards show purpose without technical metadata', () => {
  const cardSource = source.slice(
    source.indexOf('function createOnboardingApplicationCard'),
    source.indexOf('function renderApplicationCatalogError'),
  );

  assert.match(cardSource, /description\.textContent = target\.description/);
  assert.doesNotMatch(cardSource, /metadata/);
  assert.doesNotMatch(cardSource, /onboarding\.ownContainer/);
  assert.doesNotMatch(cardSource, /onboarding\.autoUnavailable/);
  assert.doesNotMatch(cardSource, /onboarding\.recommends/);
});
