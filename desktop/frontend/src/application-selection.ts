export interface SelectableApplication {
  id: string;
  dependencies?: readonly string[];
}

export interface MissingIntegration {
  consumerID: string;
  integrationID: string;
}

export function toggleApplicationSelection(
  current: Iterable<string>,
  targetID: string,
  applications: readonly SelectableApplication[],
): string[] {
  const selected = new Set(current);
  if (selected.has(targetID)) {
    selected.delete(targetID);
    return [...selected].sort();
  }

  const byID = new Map(applications.map((application) => [application.id, application]));
  const includeWithIntegrations = (applicationID: string): void => {
    if (selected.has(applicationID)) return;
    selected.add(applicationID);
    for (const integrationID of byID.get(applicationID)?.dependencies ?? []) {
      includeWithIntegrations(integrationID);
    }
  };
  includeWithIntegrations(targetID);
  return [...selected].sort();
}

export function missingSelectedIntegrations(
  selectedApplicationIDs: Iterable<string>,
  applications: readonly SelectableApplication[],
): MissingIntegration[] {
  const selected = new Set(selectedApplicationIDs);
  const missing: MissingIntegration[] = [];
  for (const application of applications) {
    if (!selected.has(application.id)) continue;
    for (const integrationID of application.dependencies ?? []) {
      if (!selected.has(integrationID)) {
        missing.push({ consumerID: application.id, integrationID });
      }
    }
  }
  return missing.sort(
    (left, right) =>
      left.consumerID.localeCompare(right.consumerID) ||
      left.integrationID.localeCompare(right.integrationID),
  );
}
