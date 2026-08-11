import assert from 'node:assert/strict';
import { readFileSync } from 'node:fs';
import test from 'node:test';

const source = readFileSync(new URL('./main.ts', import.meta.url), 'utf8');

test('installation failures expose a native support report copy action', () => {
  assert.match(source, /id="onboarding-copy-support-report"/);
  assert.match(source, /Copiar relatório técnico/);
  assert.match(source, /await CopyLastInstallationSupportReport\(\)/);
  assert.match(source, /colá-lo diretamente na issue do GitHub/);
});
