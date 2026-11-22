import { PayloadAction, createSlice } from '@reduxjs/toolkit';
import {
  AwsProvider,
  AzureProvider,
  GcpProvider,
  CloudProviderStatus,
  GetDiscoveredConnectionsQuery,
} from '@graphql';

export type LocalCloudProvider =
  Omit<AwsProvider, '__typename'> &
  Partial<Omit<AzureProvider, '__typename'>> &
  Partial<Omit<GcpProvider, '__typename'>> & {
  __typename?: 'AWSProvider' | 'AzureProvider' | 'GCPProvider';

  IsEnvironmentDefined?: boolean;
};

export type LocalDiscoveredConnection = GetDiscoveredConnectionsQuery['DiscoveredConnections'][number];

export interface IProvidersState {

  cloudProviders: LocalCloudProvider[];

  discoveredConnections: LocalDiscoveredConnection[];

  isProviderModalOpen: boolean;

  editingProviderId: string | null;

  isLoading: boolean;

  error: string | null;
}

const initialState: IProvidersState = {
  cloudProviders: [],
  discoveredConnections: [],
  isProviderModalOpen: false,
  editingProviderId: null,
  isLoading: false,
  error: null,
};

export const providersSlice = createSlice({
  name: 'providers',
  initialState,
  reducers: {

    setCloudProviders: (state, action: PayloadAction<LocalCloudProvider[]>) => {
      state.cloudProviders = action.payload;
      state.error = null;
    },

    addCloudProvider: (state, action: PayloadAction<LocalCloudProvider>) => {
      const existingIndex = state.cloudProviders.findIndex(p => p.Id === action.payload.Id);
      if (existingIndex >= 0) {
        state.cloudProviders[existingIndex] = action.payload;
      } else {
        state.cloudProviders.push(action.payload);
      }
      state.error = null;
    },

    updateCloudProvider: (state, action: PayloadAction<LocalCloudProvider>) => {
      const index = state.cloudProviders.findIndex(p => p.Id === action.payload.Id);
      if (index >= 0) {
        state.cloudProviders[index] = action.payload;
      }
      state.error = null;
    },

    removeCloudProvider: (state, action: PayloadAction<{ id: string }>) => {
      state.cloudProviders = state.cloudProviders.filter(p => p.Id !== action.payload.id);

      state.discoveredConnections = state.discoveredConnections.filter(
        c => c.ProviderID !== action.payload.id
      );
      state.error = null;
    },

    setProviderStatus: (state, action: PayloadAction<{ id: string; status: CloudProviderStatus; error?: string }>) => {
      const provider = state.cloudProviders.find(p => p.Id === action.payload.id);
      if (provider) {
        provider.Status = action.payload.status;
        if (action.payload.error) {
          state.error = action.payload.error;
        }
      }
    },

    updateProviderDiscovery: (state, action: PayloadAction<{
      id: string;
      discoveredCount: number;
      lastDiscoveryAt: string;
    }>) => {
      const provider = state.cloudProviders.find(p => p.Id === action.payload.id);
      if (provider) {
        provider.DiscoveredCount = action.payload.discoveredCount;
        provider.LastDiscoveryAt = action.payload.lastDiscoveryAt;
        provider.Status = CloudProviderStatus.Connected;
      }
    },

    setDiscoveredConnections: (state, action: PayloadAction<LocalDiscoveredConnection[]>) => {
      state.discoveredConnections = action.payload;
    },

    setProviderConnections: (state, action: PayloadAction<{
      providerId: string;
      connections: LocalDiscoveredConnection[];
    }>) => {

      state.discoveredConnections = state.discoveredConnections.filter(
        c => c.ProviderID !== action.payload.providerId
      );

      state.discoveredConnections.push(...action.payload.connections);
    },

    openAddProviderModal: (state) => {
      state.isProviderModalOpen = true;
      state.editingProviderId = null;
      state.error = null;
    },

    openEditProviderModal: (state, action: PayloadAction<{ id: string }>) => {
      state.isProviderModalOpen = true;
      state.editingProviderId = action.payload.id;
      state.error = null;
    },

    closeProviderModal: (state) => {
      state.isProviderModalOpen = false;
      state.editingProviderId = null;
    },

    setLoading: (state, action: PayloadAction<boolean>) => {
      state.isLoading = action.payload;
    },

    setError: (state, action: PayloadAction<string | null>) => {
      state.error = action.payload;
    },

    clearProviders: (state) => {
      state.cloudProviders = [];
      state.discoveredConnections = [];
      state.isProviderModalOpen = false;
      state.editingProviderId = null;
      state.isLoading = false;
      state.error = null;
    },
  },
});

export const ProvidersActions = providersSlice.actions;
export const providersReducers = providersSlice.reducer;