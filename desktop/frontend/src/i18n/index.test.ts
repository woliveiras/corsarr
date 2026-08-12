import assert from 'node:assert/strict';
import test from 'node:test';
import { detectLocale, normalizeLocale, supportedLocales } from './index.ts';
import { resources } from './resources.ts';

test('normalizes supported regional locales and rejects unsupported languages', () => {
  assert.equal(normalizeLocale('en-US'), 'en');
  assert.equal(normalizeLocale('es_MX'), 'es');
  assert.equal(normalizeLocale('pt-PT'), 'pt-BR');
  assert.equal(normalizeLocale('it-IT'), 'it');
  assert.equal(normalizeLocale('fr-FR'), undefined);
});

test('prefers the persisted locale and falls back to the browser then English', () => {
  assert.equal(detectLocale('it', ['es-ES']), 'it');
  assert.equal(detectLocale(undefined, ['fr-FR', 'es-ES']), 'es');
  assert.equal(detectLocale(undefined, ['fr-FR']), 'en');
});

test('all frontend locales expose exactly the same translation keys', () => {
  assert.deepEqual(Object.keys(resources), [...supportedLocales]);
  const expected = Object.keys(resources.en).sort();
  for (const locale of supportedLocales) {
    assert.deepEqual(Object.keys(resources[locale]).sort(), expected, locale);
  }
});
