import type { ApolloCache } from "@apollo/client";
import {
    GetAzureProvidersDocument,
    GetAzureProvidersQuery,
    GetCloudProvidersDocument,
    GetCloudProvidersQuery,
} from "@graphql";

type WithId = {
    Id: string;
};

function upsertItemById<T extends WithId>(items: readonly T[] | undefined, item: T): T[] {
    const next = items ? [...items] : [];
    const existingIndex = next.findIndex(existing => existing.Id === item.Id);

    if (existingIndex === -1) {
        next.push(item);
        return next;
    }

    next[existingIndex] = item;
    return next;
}

function removeItemById<T extends WithId>(items: readonly T[] | undefined, id: string): T[] {
    return (items ?? []).filter(item => item.Id !== id);
}

export function upsertCloudProviderCache(
    cache: ApolloCache,
    provider: GetCloudProvidersQuery["CloudProviders"][number]
): void {
    const existing = cache.readQuery<GetCloudProvidersQuery>({
        query: GetCloudProvidersDocument,
    });

    if (!existing) {
        return;
    }

    cache.writeQuery<GetCloudProvidersQuery>({
        query: GetCloudProvidersDocument,
        data: {
            CloudProviders: upsertItemById(existing.CloudProviders, provider),
        },
    });
}

export function removeCloudProviderCache(
    cache: ApolloCache,
    id: string
): void {
    const existing = cache.readQuery<GetCloudProvidersQuery>({
        query: GetCloudProvidersDocument,
    });

    if (!existing) {
        return;
    }

    cache.writeQuery<GetCloudProvidersQuery>({
        query: GetCloudProvidersDocument,
        data: {
            CloudProviders: removeItemById(existing.CloudProviders, id),
        },
    });
}

export function upsertAzureProviderCache(
    cache: ApolloCache,
    provider: GetAzureProvidersQuery["AzureProviders"][number]
): void {
    const existing = cache.readQuery<GetAzureProvidersQuery>({
        query: GetAzureProvidersDocument,
    });

    if (!existing) {
        return;
    }

    cache.writeQuery<GetAzureProvidersQuery>({
        query: GetAzureProvidersDocument,
        data: {
            AzureProviders: upsertItemById(existing.AzureProviders, provider),
        },
    });
}

export function removeAzureProviderCache(
    cache: ApolloCache,
    id: string
): void {
    const existing = cache.readQuery<GetAzureProvidersQuery>({
        query: GetAzureProvidersDocument,
    });

    if (!existing) {
        return;
    }

    cache.writeQuery<GetAzureProvidersQuery>({
        query: GetAzureProvidersDocument,
        data: {
            AzureProviders: removeItemById(existing.AzureProviders, id),
        },
    });
}