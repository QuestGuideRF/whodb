import { IAIModelType } from '../store/ai-models';

export function ensureModelTypesArray(modelTypes: any): IAIModelType[] {
  if (Array.isArray(modelTypes)) {
    return modelTypes;
  }

  return [{
    id: "Ollama",
    modelType: "Ollama",
  }];
}

export function ensureModelsArray(models: any): string[] {
  if (Array.isArray(models)) {
    return models;
  }

  return [];
}