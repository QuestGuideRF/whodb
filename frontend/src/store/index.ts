import { combineReducers, configureStore, Reducer } from '@reduxjs/toolkit';
import { persistReducer, persistStore, createTransform } from 'redux-persist';
import storage from 'redux-persist/lib/storage';
import { authReducers } from './auth';
import { databaseReducers } from './database';
import { settingsReducers } from "./settings";
import { houdiniReducers } from './chat';
import { aiModelsReducers } from './ai-models';
import { scratchpadReducers, IScratchpadState } from './scratchpad';
import { IChatState } from './chat';
import { tourReducers } from './tour';
import { providersReducers } from './providers';
import { healthReducers } from './health';
import { exploreConditionsReducers } from './explore-conditions';
import { runMigrations } from './migrations';

runMigrations();

if (typeof window !== 'undefined') {
  try {
    const scratchpadData = localStorage.getItem('persist:scratchpad');
    if (scratchpadData) {
      const parsed = JSON.parse(scratchpadData);
      if (parsed && parsed.cells) {

        const hasInvalidDates = Object.values(parsed.cells).some((cell: any) => {
          if (cell && cell.history && Array.isArray(cell.history)) {
            return cell.history.some((item: any) =>
              item.date && typeof item.date === 'string' && isNaN(new Date(item.date).getTime())
            );
          }
          return false;
        });

        if (hasInvalidDates) {
          console.warn('Clearing corrupted scratchpad data due to invalid dates');
          localStorage.removeItem('persist:scratchpad');
        }
      }
    }
  } catch (error) {
    console.warn('Error checking scratchpad data, clearing it:', error);
    localStorage.removeItem('persist:scratchpad');
  }
}

const scratchpadTransform = createTransform(

  (inboundState: IScratchpadState) => {
    return inboundState;
  },

  (outboundState: any) => {
    if (!outboundState || !outboundState.cells) {
      return outboundState;
    }

    const transformedCells: Record<string, any> = {};
    Object.keys(outboundState.cells).forEach(cellId => {
      const cell = outboundState.cells[cellId];
      if (cell && cell.history && Array.isArray(cell.history)) {
        transformedCells[cellId] = {
          ...cell,
          history: cell.history.map((historyItem: any) => {
            let date = historyItem.date;

            if (typeof date === 'string') {
              date = new Date(date);
            } else if (!(date instanceof Date)) {
              date = new Date();
            }

            if (isNaN(date.getTime())) {
              date = new Date();
            }
            return {
              ...historyItem,
              date
            };
          })
        };
      } else {
        transformedCells[cellId] = cell;
      }
    });

    return {
      ...outboundState,
      cells: transformedCells
    };
  },
  { whitelist: ['scratchpad'] }
);

const chatTransform = createTransform(

  (inboundState: IChatState) => {
    return inboundState;
  },

  (outboundState: any) => {
    if (!outboundState || !outboundState.sessions) {
      return outboundState;
    }

    const transformedSessions = outboundState.sessions.map((session: any) => {
      let createdAt = session.createdAt;

      if (typeof createdAt === 'string') {
        createdAt = new Date(createdAt);
      } else if (!(createdAt instanceof Date)) {
        createdAt = new Date();
      }

      if (isNaN(createdAt.getTime())) {
        createdAt = new Date();
      }
      return {
        ...session,
        createdAt
      };
    });

    return {
      ...outboundState,
      sessions: transformedSessions
    };
  },
  { whitelist: ['houdini'] }
);

const aiModelsPersistTransform = createTransform(
  (inboundState: any) => {
    if (!inboundState?.modelTypes) return inboundState;
    return {
      ...inboundState,
      modelTypes: inboundState.modelTypes.filter((m: any) => !m.isPlatformProvider),
    };
  },
  (outboundState: any) => outboundState,
  { whitelist: ['aiModels'] }
);

const ceReducerMap = {
  auth: persistReducer({ key: "auth", storage, }, authReducers),
  database: persistReducer({ key: "database", storage, }, databaseReducers),
  settings: persistReducer({ key: "settings", storage }, settingsReducers),
  houdini: persistReducer({
    key: "houdini",
    storage,
    transforms: [chatTransform]
  }, houdiniReducers),
  aiModels: persistReducer({ key: "aiModels", storage, transforms: [aiModelsPersistTransform] }, aiModelsReducers),
  scratchpad: persistReducer({
    key: "scratchpad",
    storage,
    transforms: [scratchpadTransform]
  }, scratchpadReducers),
  tour: persistReducer({ key: "tour", storage }, tourReducers),
  providers: persistReducer({ key: "providers", storage }, providersReducers),
  health: healthReducers,
  exploreConditions: persistReducer({ key: 'exploreConditions', storage }, exploreConditionsReducers),
};

const eeReducerMap: Record<string, Reducer> = {};

function buildRootReducer() {
  return combineReducers({ ...ceReducerMap, ...eeReducerMap });
}

export const reduxStore = configureStore({
  reducer: buildRootReducer(),
  middleware: (getDefaultMiddleware) => {
    return getDefaultMiddleware({
      serializableCheck: false,
    });
  },
});

export function registerReducer(key: string, reducer: Reducer): void {
  if (key in eeReducerMap) return;
  eeReducerMap[key] = reducer;

  reduxStore.replaceReducer(buildRootReducer() as any);
}

export const reduxStorePersistor = persistStore(reduxStore);

export type RootState = ReturnType<typeof reduxStore.getState>;
export type AppDispatch = typeof reduxStore.dispatch;