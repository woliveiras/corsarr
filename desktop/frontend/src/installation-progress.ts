export type InstallationStage = 'installing' | 'provisioning' | 'ready' | 'failed';
export type TrackedInstallationStage = 'waiting' | InstallationStage;

export interface InstallationProgressEvent {
  applicationId: string;
  stage: InstallationStage;
  position: number;
  total: number;
}

export interface InstallationProgressItem {
  applicationId: string;
  position?: number;
  stage: TrackedInstallationStage;
}

export function createInstallationProgress(
  selectedApplicationIDs: Iterable<string>,
): InstallationProgressItem[] {
  return [...new Set(selectedApplicationIDs)]
    .sort()
    .map((applicationId) => ({ applicationId, stage: 'waiting' }));
}

export function applyInstallationProgress(
  current: readonly InstallationProgressItem[],
  progress: InstallationProgressEvent,
): InstallationProgressItem[] {
  const items = current.map((item) =>
    item.applicationId === progress.applicationId
      ? { applicationId: item.applicationId, position: progress.position, stage: progress.stage }
      : item,
  );
  if (!items.some(({ applicationId }) => applicationId === progress.applicationId)) {
    items.push({
      applicationId: progress.applicationId,
      position: progress.position,
      stage: progress.stage,
    });
  }
  return items.sort((left, right) => {
    if (left.position !== undefined && right.position !== undefined) {
      return left.position - right.position;
    }
    if (left.position !== undefined) return -1;
    if (right.position !== undefined) return 1;
    return left.applicationId.localeCompare(right.applicationId);
  });
}
