import { ReactElement } from 'react';
import { TooltipPosition } from './tour-tooltip';

export interface TourStep {
    target: string;
    title: string;
    description: string;
    icon?: ReactElement;
    position?: TooltipPosition;
    beforeShow?: () => void;
    path?: string;
}

export interface TourConfig {
    id: string;
    steps: TourStep[];
}