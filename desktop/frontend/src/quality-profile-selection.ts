export function requiresQualityProfile(applications: Iterable<string>): boolean {
  for (const application of applications) {
    if (application === 'radarr' || application === 'sonarr') return true;
  }
  return false;
}

export function shouldManageQualityProfile(preset: string): boolean {
  return preset !== '' && preset !== 'unmanaged';
}

export function onboardingStepTotal(qualityRequired: boolean): 4 | 5 {
  return qualityRequired ? 5 : 4;
}
