import assert from 'node:assert/strict';
import { readFileSync } from 'node:fs';
import test from 'node:test';

const mainSource = readFileSync(new URL('./main.ts', import.meta.url), 'utf8');

test('quality onboarding omits implementation-version metadata', () => {
  const onboardingQuality = mainSource.match(
    /<article id="onboarding-quality"[\s\S]*?<\/article>/,
  )?.[0];
  assert.ok(onboardingQuality);
  assert.doesNotMatch(
    onboardingQuality,
    /Política Corsarr|Recyclarr 8\.7\.0|atualizações automáticas/,
  );
});

test('dashboard offers a dedicated product information view', () => {
  assert.match(mainSource, /id="show-info"/);
  assert.match(mainSource, /id="info-view"/);
  assert.match(mainSource, /GetProductInfo/);
  assert.match(mainSource, /corsarrVersion/);
  assert.match(mainSource, /qualityPolicyVersion/);
  assert.match(mainSource, /recyclarrVersion/);
  assert.match(mainSource, /trashGuidesCommit/);
  assert.match(mainSource, /automaticUpdates/);
});

test('quality component names open their official websites', () => {
  assert.match(mainSource, /id="info-open-recyclarr"/);
  assert.match(mainSource, /id="info-open-trash-guides"/);
  assert.match(mainSource, /openInfoComponentWebsite\(infoOpenRecyclarr, 'runtime-recyclarr'\)/);
  assert.match(mainSource, /openInfoComponentWebsite\(infoOpenTrashGuides, 'guide-trash'\)/);
});
