import type { SourceTypeOverride } from './source-types';

let registeredSourceTypeOverrides: SourceTypeOverride[] = [];

export const registerSourceTypeOverrides = (overrides: SourceTypeOverride[]) => {
    registeredSourceTypeOverrides = overrides;
};

export const getRegisteredSourceTypeOverrides = (): SourceTypeOverride[] => {
    return registeredSourceTypeOverrides;
};