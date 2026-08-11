import assert from 'node:assert/strict';
import { readFileSync } from 'node:fs';
import test from 'node:test';

const stylesheet = readFileSync(new URL('./style.css', import.meta.url), 'utf8');

test('onboarding uses viewport-bounded typography for large displays', () => {
  for (const token of [
    '--onboarding-font-micro: clamp(',
    '--onboarding-font-small: clamp(',
    '--onboarding-font-body: clamp(',
  ]) {
    assert.ok(stylesheet.includes(token), `expected responsive typography token ${token}`);
  }

  for (const selector of [
    '.onboarding .onboarding-message',
    '.onboarding .installation-progress-item strong',
    '.onboarding .application-info p',
    '.onboarding .metadata',
  ]) {
    assert.ok(stylesheet.includes(selector), `expected responsive onboarding rule ${selector}`);
  }
});
