type ApplicationLike = { id: string };
type ManagedStatusLike = { applicationId: string; state: string };

function isInstalled(state: string | undefined): boolean {
  return state !== undefined && state !== 'not_installed';
}

export function runningServicesSummary(statuses: readonly ManagedStatusLike[]): {
  installed: number;
  running: number;
} {
  return statuses.reduce(
    (summary, status) => {
      if (isInstalled(status.state)) summary.installed += 1;
      if (status.state === 'running') summary.running += 1;
      return summary;
    },
    { installed: 0, running: 0 },
  );
}

export function sortApplicationsByInstallation<T extends ApplicationLike>(
  applications: readonly T[],
  statuses: readonly ManagedStatusLike[],
): T[] {
  const stateByApplication = new Map(
    statuses.map(({ applicationId, state }) => [applicationId, state]),
  );

  return applications
    .map((application, catalogIndex) => ({ application, catalogIndex }))
    .sort((left, right) => {
      const leftInstalled = isInstalled(stateByApplication.get(left.application.id));
      const rightInstalled = isInstalled(stateByApplication.get(right.application.id));
      if (leftInstalled !== rightInstalled) return leftInstalled ? -1 : 1;
      return left.catalogIndex - right.catalogIndex;
    })
    .map(({ application }) => application);
}
