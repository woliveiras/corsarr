import assert from 'node:assert/strict';
import test from 'node:test';
import {
  onboardingStepTotal,
  requiresQualityProfile,
  shouldManageQualityProfile,
} from './quality-profile-selection.ts';

test('quality step appears when either supported Arr application is selected', () => {
  assert.equal(requiresQualityProfile(['radarr']), true);
  assert.equal(requiresQualityProfile(['sonarr']), true);
  assert.equal(requiresQualityProfile(['radarr', 'sonarr']), true);
  assert.equal(requiresQualityProfile(['jellyfin', 'qbittorrent']), false);
});

test('unmanaged selection keeps the step but disables Recyclarr ownership', () => {
  assert.equal(shouldManageQualityProfile('balanced-1080p'), true);
  assert.equal(shouldManageQualityProfile('unmanaged'), false);
  assert.equal(onboardingStepTotal(true), 5);
  assert.equal(onboardingStepTotal(false), 4);
});
