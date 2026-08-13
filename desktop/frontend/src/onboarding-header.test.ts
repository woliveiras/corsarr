import assert from 'node:assert/strict';
import { readFileSync } from 'node:fs';
import test from 'node:test';

const mainSource = readFileSync(new URL('./main.ts', import.meta.url), 'utf8');
const stylesheet = readFileSync(new URL('./style.css', import.meta.url), 'utf8');

test('onboarding descriptions are localized without replacing their step labels', () => {
  for (const step of ['environment', 'storage', 'quality']) {
    assert.ok(
      mainSource.includes(`#onboarding-${step} .onboarding-step-copy > p:not(.eyebrow)`),
      `expected ${step} description to exclude the eyebrow`,
    );
  }
});

test('onboarding header keeps progress central and its label on one line', () => {
  assert.match(
    stylesheet,
    /\.onboarding-header\s*\{[^}]*display:\s*grid;[^}]*grid-template-columns:\s*minmax\(0, 1fr\) auto minmax\(0, 1fr\);/s,
  );
  assert.match(stylesheet, /\.onboarding-progress small\s*\{[^}]*white-space:\s*nowrap;/s);
  assert.match(
    stylesheet,
    /\.onboarding-header \.language-control\s*\{[^}]*grid-column:\s*3;[^}]*justify-self:\s*end;/s,
  );
});
