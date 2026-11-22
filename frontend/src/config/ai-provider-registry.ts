import { IAIModelType } from "../store/ai-models";

export interface AIProviderOverrides {
    isActive: () => boolean;
    loadProviders: () => Promise<IAIModelType[]>;
    addProvider: (provider: { modelType: string; name: string; token: string }) => Promise<{ id: string } | null>;
    deleteProvider: (id: string) => Promise<boolean>;
    onActivate?: (callback: () => void) => () => void;
}

let overrides: AIProviderOverrides | null = null;

export function registerAIProviderOverrides(o: AIProviderOverrides): void {
    overrides = o;
}

export function getAIProviderOverrides(): AIProviderOverrides | null {
    return overrides;
}