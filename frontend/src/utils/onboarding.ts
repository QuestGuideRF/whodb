const ONBOARDING_STORAGE_KEY = '@clidey/whodb/onboarding-completed';

export const hasCompletedOnboarding = (): boolean => {
    return localStorage.getItem(ONBOARDING_STORAGE_KEY) === 'true';
};

export const markOnboardingComplete = (): void => {
    try {
        localStorage.setItem(ONBOARDING_STORAGE_KEY, 'true');
    } catch (error) {
        console.error('Failed to save onboarding state:', error);
    }
};