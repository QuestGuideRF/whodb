import { FC, ReactNode, useCallback, useEffect } from 'react';
import { useAppDispatch, useAppSelector } from '../../store/hooks';
import { TourActions } from '../../store/tour';
import { sampleDatabaseTour } from '../../config/tour-config';
import { Tour } from './tour';
import { markOnboardingComplete } from '../../utils/onboarding';
import { featureFlags } from '../../config/features';

interface TourProviderProps {
    children: ReactNode;
}

export const TourProvider: FC<TourProviderProps> = ({ children }) => {
    const dispatch = useAppDispatch();
    const tourState = useAppSelector(state => state.tour);

    useEffect(() => {
        if (tourState.shouldStartOnLoad && tourState.tourId && featureFlags.autoStartTourOnLogin) {
            setTimeout(() => {
                if (tourState.tourId) {
                    dispatch(TourActions.startTour(tourState.tourId));
                }
            }, 1000);
        }
    }, [tourState.shouldStartOnLoad, tourState.tourId, dispatch]);

    const handleComplete = useCallback(() => {
        markOnboardingComplete();
        dispatch(TourActions.stopTour());
    }, [dispatch]);

    const handleSkip = useCallback(() => {
        markOnboardingComplete();
        dispatch(TourActions.stopTour());
    }, [dispatch]);

    const getTourConfig = useCallback(() => {
        switch (tourState.tourId) {
            case 'sample-database-tour':
                return sampleDatabaseTour;
            default:
                return null;
        }
    }, [tourState.tourId]);

    const config = getTourConfig();

    return (
        <>
            {children}
            {config && (
                <Tour
                    config={config}
                    isActive={tourState.isActive}
                    onComplete={handleComplete}
                    onSkip={handleSkip}
                />
            )}
        </>
    );
};