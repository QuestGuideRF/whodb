import { useLazyQuery } from '@apollo/client/react';
import { useEffect, useCallback } from 'react';
import { SourceSessionMetadataDocument } from '@graphql';
import { useAppSelector } from '../store/hooks';
import {
    clearSourceSessionMetadata,
    setSourceSessionMetadata,
    setSourceSessionMetadataLoading,
    shouldRefreshSourceSessionMetadata,
    useSourceSessionMetadataState,
} from '../utils/source-session-metadata-cache';

export const useSourceSessionMetadata = () => {
    const auth = useAppSelector(state => state.auth);
    const metadata = useSourceSessionMetadataState();
    const currentSourceType = auth.current?.Type;

    const [fetchMetadata, { data, error, loading }] = useLazyQuery(SourceSessionMetadataDocument, {
        fetchPolicy: 'network-only',
    });

    useEffect(() => {
        if (data?.SourceSessionMetadata) {
            setSourceSessionMetadata(data.SourceSessionMetadata);
        }
    }, [data?.SourceSessionMetadata]);

    useEffect(() => {
        if (error) {
            console.error('Failed to fetch source session metadata:', error);
            setSourceSessionMetadataLoading(false);
        }
    }, [error]);

    useEffect(() => {
        if (auth.status === 'logged-in' && currentSourceType) {
            if (shouldRefreshSourceSessionMetadata(currentSourceType)) {
                setSourceSessionMetadataLoading(true);
                void fetchMetadata();
            }
        }
    }, [auth.status, currentSourceType, fetchMetadata, metadata.lastFetched, metadata.sourceType]);

    useEffect(() => {
        if (auth.status === 'unauthorized') {
            clearSourceSessionMetadata();
        }
    }, [auth.status]);

    const refresh = useCallback(() => {
        if (auth.status === 'logged-in') {
            setSourceSessionMetadataLoading(true);
            void fetchMetadata();
        }
    }, [auth.status, fetchMetadata]);

    return {
        queryLanguages: metadata.queryLanguages,
        typeDefinitions: metadata.typeDefinitions,
        operators: metadata.operators,
        aliasMap: metadata.aliasMap,
        sourceType: metadata.sourceType,
        loading: loading || metadata.loading,
        hasFetched: metadata.lastFetched !== null,
        refresh,
    };
};