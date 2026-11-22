import {createSlice, PayloadAction} from '@reduxjs/toolkit';

type HealthStatus = 'healthy' | 'error' | 'unavailable' | 'unknown';

type IHealthState = {
    serverStatus: HealthStatus;
    databaseStatus: HealthStatus;
    lastChecked: number | null;
    hasEverConnected: boolean;
    checksPerformed: number;
}

const initialState: IHealthState = {
    serverStatus: 'unknown',
    databaseStatus: 'unknown',
    lastChecked: null,
    hasEverConnected: false,
    checksPerformed: 0,
};

export const healthSlice = createSlice({
    name: 'health',
    initialState,
    reducers: {
        setServerStatus: (state, action: PayloadAction<HealthStatus>) => {
            state.serverStatus = action.payload;
            state.lastChecked = Date.now();
        },
        setDatabaseStatus: (state, action: PayloadAction<HealthStatus>) => {
            state.databaseStatus = action.payload;
            state.lastChecked = Date.now();
        },
        setHealthStatus: (state, action: PayloadAction<{server: HealthStatus, database: HealthStatus}>) => {
            state.serverStatus = action.payload.server;
            state.databaseStatus = action.payload.database;
            state.lastChecked = Date.now();
            state.checksPerformed += 1;

            if (action.payload.server === 'healthy') {
                state.hasEverConnected = true;
            }
        },
        resetHealth: (state) => {
            state.serverStatus = 'unknown';
            state.databaseStatus = 'unknown';
            state.lastChecked = null;
            state.hasEverConnected = false;
            state.checksPerformed = 0;
        },
    },
});

export const HealthActions = healthSlice.actions;
export const healthReducers = healthSlice.reducer;