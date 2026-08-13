const defaultRetryDelay = 5_000;

type TimerHandle = number;

type EnvironmentRecoveryOptions = {
  recheck: () => Promise<void>;
  schedule?: (callback: () => void, delay: number) => TimerHandle;
  cancel?: (timer: TimerHandle) => void;
  retryDelay?: number;
};

export type EnvironmentRecoveryPoller = {
  observe: (runtimeState: string) => void;
  stop: () => void;
};

export function createEnvironmentRecoveryPoller({
  recheck,
  schedule = (callback, delay) => window.setTimeout(callback, delay),
  cancel = (timer) => window.clearTimeout(timer),
  retryDelay = defaultRetryDelay,
}: EnvironmentRecoveryOptions): EnvironmentRecoveryPoller {
  let pendingTimer: TimerHandle | undefined;

  const stop = (): void => {
    if (pendingTimer === undefined) return;
    cancel(pendingTimer);
    pendingTimer = undefined;
  };

  const observe = (runtimeState: string): void => {
    if (runtimeState === 'ready') {
      stop();
      return;
    }
    if (pendingTimer !== undefined) return;

    pendingTimer = schedule(() => {
      pendingTimer = undefined;
      void recheck().catch(() => observe('error'));
    }, retryDelay);
  };

  return { observe, stop };
}
