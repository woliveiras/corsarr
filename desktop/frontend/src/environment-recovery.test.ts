import assert from 'node:assert/strict';
import test from 'node:test';

import { createEnvironmentRecoveryPoller } from './environment-recovery.ts';

test('rechecks a stopped environment until Docker becomes ready', async () => {
  const scheduled = new Map<number, () => void>();
  const delays: number[] = [];
  let nextTimer = 0;
  let runtimeState = 'stopped';
  let checks = 0;

  let poller: ReturnType<typeof createEnvironmentRecoveryPoller>;
  poller = createEnvironmentRecoveryPoller({
    recheck: async () => {
      checks++;
      poller.observe(runtimeState);
    },
    schedule: (callback, delay) => {
      const timer = ++nextTimer;
      scheduled.set(timer, callback);
      delays.push(delay);
      return timer;
    },
    cancel: (timer) => scheduled.delete(timer),
  });

  poller.observe(runtimeState);
  assert.deepEqual(delays, [5_000]);
  assert.equal(scheduled.size, 1);

  const firstRetry = scheduled.values().next().value;
  assert.ok(firstRetry);
  scheduled.clear();
  firstRetry();
  await Promise.resolve();

  assert.equal(checks, 1);
  assert.equal(scheduled.size, 1);

  runtimeState = 'ready';
  const secondRetry = scheduled.values().next().value;
  assert.ok(secondRetry);
  scheduled.clear();
  secondRetry();
  await Promise.resolve();

  assert.equal(checks, 2);
  assert.equal(scheduled.size, 0);
});

test('keeps at most one environment retry scheduled', () => {
  let scheduled = 0;
  const poller = createEnvironmentRecoveryPoller({
    recheck: async () => {},
    schedule: () => {
      scheduled++;
      return scheduled;
    },
    cancel: () => {},
  });

  poller.observe('stopped');
  poller.observe('error');
  poller.observe('unavailable');

  assert.equal(scheduled, 1);
});

test('cancels a stale retry when another check finds Docker ready', () => {
  const pending = new Set<number>();
  let nextTimer = 0;
  const poller = createEnvironmentRecoveryPoller({
    recheck: async () => {},
    schedule: () => {
      const timer = ++nextTimer;
      pending.add(timer);
      return timer;
    },
    cancel: (timer) => pending.delete(timer),
  });

  poller.observe('stopped');
  assert.equal(pending.size, 1);

  poller.observe('ready');
  assert.equal(pending.size, 0);
});
