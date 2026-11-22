import {createSlice, PayloadAction} from '@reduxjs/toolkit';
import {settingsDefaults} from '../config/features';
import {type SupportedLanguage, DEFAULT_LANGUAGE} from '../utils/languages';

const ANALYTICS_CONSENT_KEY = 'whodb.analytics.consent';

type ISettingsState = {
    metricsEnabled: boolean;
    cloudProvidersEnabled: boolean;
    storageUnitView: 'list' | 'card';
    fontSize: 'small' | 'medium' | 'large';
    borderRadius: 'none' | 'small' | 'medium' | 'large';
    spacing: 'compact' | 'comfortable' | 'spacious';
    whereConditionMode: 'popover' | 'sheet';
    defaultPageSize: number;
    maxPageSize: number;
    language: SupportedLanguage;
    databaseSchemaTerminology: 'database' | 'schema';
    disableAnimations: boolean;

    appTheme: 'default';

    os: 'linux' | 'macos' | 'windows' | undefined;
}

const getInitialMetricsEnabled = (): boolean => {
    if (typeof window === 'undefined') {
        return true;
    }

    const consent = window.localStorage.getItem(ANALYTICS_CONSENT_KEY);
    if (consent === 'denied') {
        return false;
    }
    return true;
};

const getInitialState = (): ISettingsState => {
    return {
        metricsEnabled: getInitialMetricsEnabled(),
        cloudProvidersEnabled: false,
        storageUnitView: 'card',
        fontSize: 'medium',
        borderRadius: 'medium',
        spacing: 'comfortable',

        whereConditionMode: settingsDefaults.whereConditionMode ?? 'popover',
        defaultPageSize: 100,
        maxPageSize: 10000,
        language: DEFAULT_LANGUAGE,
        databaseSchemaTerminology: 'database',

        disableAnimations: settingsDefaults.disableAnimations ?? false,
        appTheme: 'default',
        os: undefined,
    };
};

const initialState = getInitialState();

export const settingsSlice = createSlice({
    name: 'settings',
    initialState,
    reducers: {
        setMetricsEnabled: (state, action: PayloadAction<ISettingsState["metricsEnabled"]>) => {
            state.metricsEnabled = action.payload;
        },
        setCloudProvidersEnabled: (state, action: PayloadAction<ISettingsState["cloudProvidersEnabled"]>) => {
            state.cloudProvidersEnabled = action.payload;
        },
        setStorageUnitView: (state, action: PayloadAction<ISettingsState["storageUnitView"]>) => {
            state.storageUnitView = action.payload;
        },

        setFontSize: (state, action: PayloadAction<ISettingsState["fontSize"]>) => {
            state.fontSize = action.payload;
        },
        setBorderRadius: (state, action: PayloadAction<ISettingsState["borderRadius"]>) => {
            state.borderRadius = action.payload;
        },
        setSpacing: (state, action: PayloadAction<ISettingsState["spacing"]>) => {
            state.spacing = action.payload;
        },
        setWhereConditionMode: (state, action: PayloadAction<ISettingsState["whereConditionMode"]>) => {
            state.whereConditionMode = action.payload;
        },
        setDefaultPageSize: (state, action: PayloadAction<ISettingsState["defaultPageSize"]>) => {
            state.defaultPageSize = action.payload;
        },
        setMaxPageSize: (state, action: PayloadAction<ISettingsState["maxPageSize"]>) => {
            state.maxPageSize = action.payload;
        },
        setLanguage: (state, action: PayloadAction<ISettingsState["language"]>) => {
            state.language = action.payload;
        },
        setDatabaseSchemaTerminology: (state, action: PayloadAction<ISettingsState["databaseSchemaTerminology"]>) => {
            state.databaseSchemaTerminology = action.payload;
        },
        setDisableAnimations: (state, action: PayloadAction<ISettingsState["disableAnimations"]>) => {
            state.disableAnimations = action.payload;
        },
        setAppTheme: (state, action: PayloadAction<ISettingsState["appTheme"]>) => {
            state.appTheme = action.payload;
        },
        setOS: (state, action: PayloadAction<ISettingsState["os"]>) => {
            state.os = action.payload;
        },
    },
});

export const SettingsActions = settingsSlice.actions;
export const settingsReducers = settingsSlice.reducer;