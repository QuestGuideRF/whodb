import { IAIModelType } from './ai-models';
import { ensureModelTypesArray, ensureModelsArray } from '../utils/ai-models-helper';
import { featureFlags } from '../config/features';

interface OldDatabaseState {
  schema: string;
  current?: IAIModelType;
  modelTypes?: IAIModelType[];
  currentModel?: string;
  models?: string[];
}

export function migrateAIModelsFromDatabase(): void {
  try {

    const migrationKey = 'aiModels_migration_v1_completed';
    if (localStorage.getItem(migrationKey) === 'true') {
      return;
    }

    const persistedDatabaseState = localStorage.getItem('persist:database');
    if (!persistedDatabaseState) {

      localStorage.setItem(migrationKey, 'true');
      return;
    }

    const databaseState = JSON.parse(persistedDatabaseState);

    let current, modelTypes, currentModel, models;

    try {
      current = databaseState.current ? JSON.parse(databaseState.current) : undefined;
      modelTypes = databaseState.modelTypes ? JSON.parse(databaseState.modelTypes) : undefined;
      currentModel = databaseState.currentModel ? JSON.parse(databaseState.currentModel) : undefined;
      models = databaseState.models ? JSON.parse(databaseState.models) : undefined;
    } catch (e) {

      current = databaseState.current;
      modelTypes = databaseState.modelTypes;
      currentModel = databaseState.currentModel;
      models = databaseState.models;
    }

    if (!modelTypes && !current && !currentModel && !models) {

      localStorage.setItem(migrationKey, 'true');
      return;
    }

    const aiModelsState: Record<string, string> = {};

    const defaultModelTypes = ["Ollama"].map(modelType => ({
      id: modelType,
      modelType,
    }));

    if (current !== undefined) {
      aiModelsState.current = JSON.stringify(current);
    }

    if (modelTypes !== undefined && Array.isArray(modelTypes)) {
      aiModelsState.modelTypes = JSON.stringify(modelTypes);
    } else {

      aiModelsState.modelTypes = JSON.stringify(defaultModelTypes);
    }

    if (currentModel !== undefined) {
      aiModelsState.currentModel = JSON.stringify(currentModel);
    }

    if (models !== undefined && Array.isArray(models)) {
      aiModelsState.models = JSON.stringify(models);
    } else {
      aiModelsState.models = JSON.stringify([]);
    }

    aiModelsState._persist = JSON.stringify({ version: -1, rehydrated: true });

    localStorage.setItem('persist:aiModels', JSON.stringify(aiModelsState));

    const cleanedDatabaseState = {
      schema: databaseState.schema || '""',
      _persist: JSON.stringify({ version: -1, rehydrated: true })
    };
    localStorage.setItem('persist:database', JSON.stringify(cleanedDatabaseState));

    localStorage.setItem(migrationKey, 'true');
      } catch (error) {
    console.error('Error during AI models migration:', error);

  }
}

function ensureValidAIModelsState(): void {
  try {
    const persistedAIModelsState = localStorage.getItem('persist:aiModels');
    if (!persistedAIModelsState) {
      return;
    }

    const aiModelsState = JSON.parse(persistedAIModelsState);
    let needsUpdate = false;

    if (aiModelsState.modelTypes) {
      try {
        let modelTypes = JSON.parse(aiModelsState.modelTypes);
        if (!Array.isArray(modelTypes)) {
          modelTypes = ensureModelTypesArray(modelTypes);
          aiModelsState.modelTypes = JSON.stringify(modelTypes);
          needsUpdate = true;
        } else {

          const validModelTypes = modelTypes.filter((mt: IAIModelType | null) => mt && mt.id && mt.modelType);
          if (validModelTypes.length !== modelTypes.length) {
            modelTypes = validModelTypes.length > 0 ? validModelTypes : ensureModelTypesArray(null);
            aiModelsState.modelTypes = JSON.stringify(modelTypes);
            needsUpdate = true;
          }
        }
      } catch (e) {
        aiModelsState.modelTypes = JSON.stringify(ensureModelTypesArray(null));
        needsUpdate = true;
      }
    } else {

      aiModelsState.modelTypes = JSON.stringify(ensureModelTypesArray(null));
      needsUpdate = true;
    }

    if (aiModelsState.models) {
      try {
        const models = JSON.parse(aiModelsState.models);
        if (!Array.isArray(models)) {
          aiModelsState.models = JSON.stringify(ensureModelsArray(models));
          needsUpdate = true;
        }
      } catch (e) {
        aiModelsState.models = JSON.stringify(ensureModelsArray(null));
        needsUpdate = true;
      }
    } else {

      aiModelsState.models = JSON.stringify([]);
      needsUpdate = true;
    }

    if (!aiModelsState._persist) {
      aiModelsState._persist = JSON.stringify({ version: -1, rehydrated: true });
      needsUpdate = true;
    }

    if (needsUpdate) {
      localStorage.setItem('persist:aiModels', JSON.stringify(aiModelsState));
    }
  } catch (error) {
    console.error('Error ensuring valid AI models state:', error);
  }
}

function getMigrationVersion(): number {
  const version = localStorage.getItem('whodb_migration_version');
  return version ? parseInt(version, 10) : 0;
}

function setMigrationVersion(version: number): void {
  localStorage.setItem('whodb_migration_version', version.toString());
}

function clearChatStateV2(): void {
  try {

    localStorage.removeItem('persist:houdini');
  } catch (error) {
    console.error('Error clearing chat state:', error);
  }
}

function applyEESettingsDefaultsV4(): void {
  try {

    if (featureFlags.settingsPage) {
      return;
    }

    const persistedSettingsState = localStorage.getItem('persist:settings');
    if (!persistedSettingsState) {
      return;
    }

    const settingsState = JSON.parse(persistedSettingsState);
    let needsUpdate = false;

    if (settingsState.whereConditionMode) {
      try {
        const whereConditionMode = JSON.parse(settingsState.whereConditionMode);
        if (whereConditionMode !== 'sheet') {
          settingsState.whereConditionMode = JSON.stringify('sheet');
          needsUpdate = true;
        }
      } catch (e) {

        settingsState.whereConditionMode = JSON.stringify('sheet');
        needsUpdate = true;
      }
    } else {

      settingsState.whereConditionMode = JSON.stringify('sheet');
      needsUpdate = true;
    }

    if (settingsState.disableAnimations) {
      try {
        const disableAnimations = JSON.parse(settingsState.disableAnimations);
        if (disableAnimations !== true) {
          settingsState.disableAnimations = JSON.stringify(true);
          needsUpdate = true;
        }
      } catch (e) {

        settingsState.disableAnimations = JSON.stringify(true);
        needsUpdate = true;
      }
    } else {

      settingsState.disableAnimations = JSON.stringify(true);
      needsUpdate = true;
    }

    if (needsUpdate) {
      localStorage.setItem('persist:settings', JSON.stringify(settingsState));
    }
  } catch (error) {
    console.error('Error applying settings defaults:', error);
  }
}

function migrateLanguageCodesV5(): void {
  try {
    const persistedSettingsState = localStorage.getItem('persist:settings');
    if (!persistedSettingsState) {
      return;
    }

    const settingsState = JSON.parse(persistedSettingsState);
    if (!settingsState.language) {
      return;
    }

    const languageMap: Record<string, string> = {
      'en': 'en_US',
      'es': 'es_ES',
      'fr': 'fr_FR',
      'de': 'de_DE',
    };

    const currentLanguage = JSON.parse(settingsState.language);
    const mapped = languageMap[currentLanguage];
    if (mapped) {
      settingsState.language = JSON.stringify(mapped);
      localStorage.setItem('persist:settings', JSON.stringify(settingsState));
    }
  } catch (error) {
    console.error('Error migrating language codes:', error);
  }
}

function clearDatabaseMetadataStateV6(): void {
  try {
    localStorage.removeItem('persist:databaseMetadata');
  } catch (error) {
    console.error('Error clearing database metadata state:', error);
  }
}

function clearAuthStateV8(): void {
  try {
    localStorage.removeItem('persist:auth');
  } catch (error) {
    console.error('Error clearing auth state:', error);
  }
}

export function runMigrations(): void {
  const currentVersion = getMigrationVersion();

  if (currentVersion < 1) {
    migrateAIModelsFromDatabase();
    setMigrationVersion(1);
  }

  if (currentVersion < 2) {
    clearChatStateV2();
    setMigrationVersion(2);
  }

  if (currentVersion < 4) {
    applyEESettingsDefaultsV4();
    setMigrationVersion(4);
  }

  if (currentVersion < 5) {
    migrateLanguageCodesV5();
    setMigrationVersion(5);
  }

  if (currentVersion < 6) {
    clearDatabaseMetadataStateV6();
    setMigrationVersion(6);
  }

  if (currentVersion < 8) {
    clearAuthStateV8();
    setMigrationVersion(8);
  }

  ensureValidAIModelsState();
}