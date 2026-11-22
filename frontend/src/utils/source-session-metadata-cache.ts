import { makeVar } from '@apollo/client';
import { useReactiveVar } from '@apollo/client/react';
import { SourceSessionMetadataQuery } from '@graphql';
import { TypeDefinition } from '../config/source-types';

type SourceSessionMetadataPayload = NonNullable<SourceSessionMetadataQuery['SourceSessionMetadata']>;

export interface SourceSessionMetadataState {

    sourceType: string | null;

    typeDefinitions: TypeDefinition[];

    operators: string[];

    aliasMap: Record<string, string>;

    queryLanguages: string[];

    lastFetched: number | null;

    loading: boolean;
}

export const METADATA_CACHE_DURATION = 5 * 60 * 1000;

function createInitialSourceSessionMetadataState(): SourceSessionMetadataState {
    return {
        sourceType: null,
        typeDefinitions: [],
        operators: [],
        aliasMap: {},
        queryLanguages: [],
        lastFetched: null,
        loading: false,
    };
}

function mapTypeDefinitions(typeDefinitions: SourceSessionMetadataPayload['typeDefinitions']): TypeDefinition[] {
    return typeDefinitions.map(typeDefinition => ({
        id: typeDefinition.id,
        label: typeDefinition.label,
        hasLength: typeDefinition.hasLength || undefined,
        hasPrecision: typeDefinition.hasPrecision || undefined,
        defaultLength: typeDefinition.defaultLength ?? undefined,
        defaultPrecision: typeDefinition.defaultPrecision ?? undefined,
        category: typeDefinition.category,
    }));
}

function mapAliasMap(aliasMap: SourceSessionMetadataPayload['aliasMap']): Record<string, string> {
    return aliasMap.reduce((acc, item) => {
        acc[item.Key] = item.Value;
        return acc;
    }, {} as Record<string, string>);
}

const sourceSessionMetadataStateVar = makeVar<SourceSessionMetadataState>(createInitialSourceSessionMetadataState());

export function getSourceSessionMetadataState(): SourceSessionMetadataState {
    return sourceSessionMetadataStateVar();
}

export function useSourceSessionMetadataState(): SourceSessionMetadataState {
    return useReactiveVar(sourceSessionMetadataStateVar);
}

export function setSourceSessionMetadataLoading(loading: boolean): void {
    const currentState = sourceSessionMetadataStateVar();

    if (currentState.loading === loading) {
        return;
    }

    sourceSessionMetadataStateVar({
        ...currentState,
        loading,
    });
}

export function setSourceSessionMetadata(metadata: SourceSessionMetadataPayload): void {
    sourceSessionMetadataStateVar({
        sourceType: metadata.sourceType,
        queryLanguages: metadata.queryLanguages,
        typeDefinitions: mapTypeDefinitions(metadata.typeDefinitions),
        operators: metadata.operators,
        aliasMap: mapAliasMap(metadata.aliasMap),
        lastFetched: Date.now(),
        loading: false,
    });
}

export function clearSourceSessionMetadata(): void {
    sourceSessionMetadataStateVar(createInitialSourceSessionMetadataState());
}

export function shouldRefreshSourceSessionMetadata(currentSourceType: string): boolean {
    const state = sourceSessionMetadataStateVar();

    if (state.loading) {
        return false;
    }

    if (state.sourceType !== currentSourceType) {
        return true;
    }

    if (!state.lastFetched) {
        return true;
    }

    return Date.now() - state.lastFetched > METADATA_CACHE_DURATION;
}