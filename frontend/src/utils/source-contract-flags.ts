import { SourceAction } from '@graphql';
import type { SourceTypeItem } from '../config/source-types';

export interface SourceContractFlags {
    supportsChat: boolean;
    supportsGraph: boolean;
    supportsScratchpad: boolean;
    supportsSchema: boolean;
    supportsDatabaseSwitching: boolean;
    usesSchemaForGraph: boolean;
    usesDatabaseInsteadOfSchema: boolean;
    supportsMockData: boolean;
    supportsImportData: boolean;
    supportsModifiers: boolean;
}

export function resolveSourceContractFlags(
    item: SourceTypeItem | undefined
): SourceContractFlags {
    return {
        supportsChat: item?.supportsChat ?? false,
        supportsGraph: item?.supportsGraph ?? false,
        supportsScratchpad: item?.supportsScratchpad ?? false,
        supportsSchema: item?.supportsSchema ?? false,
        supportsDatabaseSwitching: item?.supportsDatabaseSwitching ?? false,
        usesSchemaForGraph: item?.usesSchemaForGraph ?? true,
        usesDatabaseInsteadOfSchema: item?.usesDatabaseInsteadOfSchema ?? false,
        supportsMockData: item?.supportsMockData ?? false,
        supportsImportData: item?.contract?.ObjectTypes.some(objectType => objectType.Actions.includes(SourceAction.ImportData)) ?? false,
        supportsModifiers: item?.supportsModifiers ?? false,
    };
}