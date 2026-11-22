import { updateDocumentMeta } from './meta';

export interface FeatureFlags {
    analyzeView: boolean;
    explainView: boolean;
    generateView: boolean;
    customTheme: boolean;
    dataVisualization: boolean;
    aiChat: boolean;
    multiProfile: boolean;
    advancedDatabases: boolean;
    contactUsPage: boolean;
    settingsPage: boolean;
    sampleDatabaseTour: boolean;
    autoStartTourOnLogin: boolean;
    sqlAgent: boolean;
}

const defaultFeatures: FeatureFlags = {
    analyzeView: false,
    explainView: false,
    generateView: false,
    customTheme: false,
    dataVisualization: false,
    aiChat: false,
    multiProfile: false,
    advancedDatabases: false,
    contactUsPage: true,
    settingsPage: true,
    sampleDatabaseTour: true,
    autoStartTourOnLogin: true,
    sqlAgent: false,
};

export let featureFlags: FeatureFlags = {} as FeatureFlags;
export let extensions: Record<string, any> = {};
export let sources: Record<string, any> = {};
export let settingsDefaults: Record<string, any> = {};

export const getAppName = (): string => extensions.AppName || "WhoDB";

export const initialize = () => {
    featureFlags = defaultFeatures;
    updateDocumentMeta({});
};

initialize();