import type { SourceTypeItem } from "@/config/source-types";
import { LocalDiscoveredConnection } from "@/store/providers";

const AWS_HOSTNAME_PATTERNS = [
    /\.rds\.amazonaws\.com$/i,
    /\.cache\.amazonaws\.com$/i,
    /\.docdb\.amazonaws\.com$/i,
    /\.redshift\.amazonaws\.com$/i,
    /\.neptune\.amazonaws\.com$/i,
    /\.timestream-influxdb\.\w+\.amazonaws\.com$/i,
];

export function isAwsHostname(hostname: string | undefined | null): boolean {
    if (!hostname) return false;
    return AWS_HOSTNAME_PATTERNS.some(pattern => pattern.test(hostname));
}

const AZURE_HOSTNAME_PATTERNS = [
    /\.postgres\.database\.azure\.com$/i,
    /\.mysql\.database\.azure\.com$/i,
    /\.redis\.cache\.windows\.net$/i,
    /\.mongo\.cosmos\.azure\.com$/i,
    /\.redisenterprise\.cache\.azure\.net$/i,
    /\.managedcassandra\.cosmos\.azure\.com$/i,
];

export function isAzureHostname(hostname: string | undefined | null): boolean {
    if (!hostname) return false;
    return AZURE_HOSTNAME_PATTERNS.some(pattern => pattern.test(hostname));
}

const GCP_HOSTNAME_PATTERNS = [
    /\.cloudsql\.goog$/i,
    /\.alloydb\.goog$/i,
    /\.memorystore\.goog$/i,
    /\.firestore\.goog$/i,
];

export function isGcpHostname(hostname: string | undefined | null): boolean {
    if (!hostname) return false;
    return GCP_HOSTNAME_PATTERNS.some(pattern => pattern.test(hostname));
}

export interface ConnectionPrefillData {
    databaseType: string;
    hostname: string;
    port?: string;
    advanced: Record<string, string>;
}

const getMetadataValue = (conn: LocalDiscoveredConnection, key: string): string | undefined => {
    if (!conn?.Metadata) return undefined;
    return conn.Metadata.find(m => m.Key === key)?.Value;
};

function providerMatches(allowedProviders: string[], providerType: string | undefined): boolean {
    if (allowedProviders.length === 0) {
        return true;
    }
    if (!providerType) {
        return false;
    }

    const normalizedProvider = providerType.toLowerCase();
    return allowedProviders.some(candidate => candidate.toLowerCase() === normalizedProvider);
}

function conditionsMatch(
    conn: LocalDiscoveredConnection,
    conditions: NonNullable<SourceTypeItem["discoveryPrefill"]>["AdvancedDefaults"][number]["Conditions"]
): boolean {
    return conditions.every(condition => getMetadataValue(conn, condition.Key) === condition.Value);
}

function resolveAdvancedDefaultValue(
    conn: LocalDiscoveredConnection,
    advancedDefault: NonNullable<SourceTypeItem["discoveryPrefill"]>["AdvancedDefaults"][number]
): string {
    if (advancedDefault.MetadataKey) {
        const metadataValue = getMetadataValue(conn, advancedDefault.MetadataKey);
        if (metadataValue) {
            return metadataValue;
        }
    }

    if (advancedDefault.Value) {
        return advancedDefault.Value;
    }

    return advancedDefault.DefaultValue;
}

function buildAdvancedPrefill(
    conn: LocalDiscoveredConnection,
    sourceType: SourceTypeItem | undefined
): Record<string, string> {
    const advanced: Record<string, string> = {};
    const port = getMetadataValue(conn, "port");
    if (port) {
        advanced["Port"] = port;
    }

    for (const advancedDefault of sourceType?.discoveryPrefill?.AdvancedDefaults ?? []) {
        if (!providerMatches(advancedDefault.ProviderTypes, conn.ProviderType)) {
            continue;
        }
        if (!conditionsMatch(conn, advancedDefault.Conditions)) {
            continue;
        }

        const value = resolveAdvancedDefaultValue(conn, advancedDefault);
        if (!value) {
            continue;
        }
        advanced[advancedDefault.Key] = value;
    }

    return advanced;
}

export function buildConnectionPrefill(
    conn: LocalDiscoveredConnection,
    sourceType?: SourceTypeItem
): ConnectionPrefillData {
    const endpoint = getMetadataValue(conn, "endpoint") || "";
    const port = getMetadataValue(conn, "port");

    return {
        databaseType: conn.DatabaseType,
        hostname: endpoint,
        port,
        advanced: buildAdvancedPrefill(conn, sourceType),
    };
}