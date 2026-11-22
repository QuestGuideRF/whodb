import type { ErrorLike } from "@apollo/client";
import { useQuery } from "@apollo/client/react";
import { useEffect, useMemo, useState } from "react";
import { SourceTypesDocument } from "@graphql";
import {
    BackendSourceType,
    findSourceTypeItem,
    readCachedSourceCatalog,
    resolveSourceConnector,
    resolveSourceTypeItems,
    SourceTypeFilterOptions,
    SourceTypeItem,
    writeCachedSourceCatalog,
} from "../config/source-types";

export interface UseSourceTypeItemsResult {

    items: SourceTypeItem[];

    loading: boolean;

    error?: ErrorLike;
}

export interface UseSourceTypeItemResult extends UseSourceTypeItemsResult {

    item?: SourceTypeItem;
}

export function useSourceTypeItems(
    options: SourceTypeFilterOptions = {}
): UseSourceTypeItemsResult {
    const [cachedCatalog] = useState<BackendSourceType[]>(() => readCachedSourceCatalog());
    const { data, loading, error } = useQuery(SourceTypesDocument, {

        fetchPolicy: import.meta.env.DEV ? "network-only" : "cache-and-network",
    });
    const cloudProvidersEnabled = options.cloudProvidersEnabled;

    useEffect(() => {
        if (data?.SourceTypes) {
            writeCachedSourceCatalog(data.SourceTypes);
        }
    }, [data?.SourceTypes]);

    const items = useMemo(() => {
        const catalog = data?.SourceTypes ?? cachedCatalog;
        return resolveSourceTypeItems(catalog, { cloudProvidersEnabled });
    }, [cachedCatalog, cloudProvidersEnabled, data?.SourceTypes]);

    return {
        items,
        loading: loading && data?.SourceTypes == null && cachedCatalog.length === 0,
        error: error ?? undefined,
    };
}

export function useSourceTypeItem(
    sourceType: string | undefined,
    options: SourceTypeFilterOptions = {}
): UseSourceTypeItemResult {
    const result = useSourceTypeItems(options);

    const item = useMemo(() => {
        return findSourceTypeItem(result.items, sourceType);
    }, [result.items, sourceType]);

    return {
        ...result,
        item,
    };
}

export function useResolvedSourceConnector(
    sourceType: string | undefined,
    options: SourceTypeFilterOptions = {}
): string | undefined {
    const { item } = useSourceTypeItem(sourceType, options);
    return resolveSourceConnector(sourceType, item);
}