import { ComponentType, ReactElement, createElement } from "react";
import {
    SourceTypesQuery,
    SourceAction,
    SourceConnectionFieldSection,
    SourceModel,
    SourceObjectKind,
    SourceSurface,
    SourceView,
} from "@graphql";
import { Icons } from "../components/icons";
import { getEdition } from "./edition";
import { getRegisteredSourceTypeOverrides } from "./source-registry";

export type TypeCategory = 'numeric' | 'text' | 'binary' | 'datetime' | 'boolean' | 'json' | 'other';

export interface SSLModeOption {

    value: string;

    aliases?: string[];
}

export interface SourceConnectionFieldDescriptor {
    Key: string;
    Kind: SourceTypesQuery['SourceTypes'][number]['connectionFields'][number]['Kind'];
    Section: SourceTypesQuery['SourceTypes'][number]['connectionFields'][number]['Section'];
    Required: boolean;
    LabelKey: string;
    PlaceholderKey?: string | null;
    DefaultValue?: string | null;
    SupportsOptions: boolean;
}

export interface SourceObjectTypeDescriptor {
    Kind: SourceObjectKind;
    DataShape: SourceTypesQuery['SourceTypes'][number]['contract']['ObjectTypes'][number]['DataShape'];
    Actions: SourceAction[];
    Views: SourceView[];
    SingularLabel: string;
    PluralLabel: string;
}

export interface SourceContractDescriptor {
    Model: SourceModel;
    Surfaces: SourceSurface[];
    RootActions: SourceAction[];
    BrowsePath: SourceObjectKind[];
    DefaultObjectKind: SourceObjectKind;
    GraphScopeKind?: SourceObjectKind | null;
    ObjectTypes: SourceObjectTypeDescriptor[];
}

export interface SourceConnectionTraitsDescriptor {
    transport: NonNullable<SourceTypesQuery['SourceTypes'][number]['traits']>['connection']['transport'];
    hostInputMode: NonNullable<SourceTypesQuery['SourceTypes'][number]['traits']>['connection']['hostInputMode'];
    hostInputUrlParser: NonNullable<SourceTypesQuery['SourceTypes'][number]['traits']>['connection']['hostInputUrlParser'];
    supportsCustomCAContent: NonNullable<SourceTypesQuery['SourceTypes'][number]['traits']>['connection']['supportsCustomCAContent'];
}

export interface SourcePresentationTraitsDescriptor {
    profileLabelStrategy: NonNullable<SourceTypesQuery['SourceTypes'][number]['traits']>['presentation']['profileLabelStrategy'];
    schemaFidelity: NonNullable<SourceTypesQuery['SourceTypes'][number]['traits']>['presentation']['schemaFidelity'];
}

export interface SourceQueryTraitsDescriptor {
    supportsAnalyze: NonNullable<SourceTypesQuery['SourceTypes'][number]['traits']>['query']['supportsAnalyze'];
    explainMode: NonNullable<SourceTypesQuery['SourceTypes'][number]['traits']>['query']['explainMode'];
}

export interface SourceMockDataTraitsDescriptor {
    supportsRelationalDependencies: NonNullable<SourceTypesQuery['SourceTypes'][number]['traits']>['mockData']['supportsRelationalDependencies'];
}

export interface SourceTraitsDescriptor {
    connection: SourceConnectionTraitsDescriptor;
    presentation: SourcePresentationTraitsDescriptor;
    query: SourceQueryTraitsDescriptor;
    mockData: SourceMockDataTraitsDescriptor;
}

export interface SourceDiscoveryMetadataConditionDescriptor {
    Key: string;
    Value: string;
}

export interface SourceDiscoveryAdvancedDefaultDescriptor {
    Key: string;
    Value: string;
    MetadataKey: string;
    DefaultValue: string;
    ProviderTypes: string[];
    Conditions: SourceDiscoveryMetadataConditionDescriptor[];
}

export interface SourceDiscoveryPrefillDescriptor {
    AdvancedDefaults: SourceDiscoveryAdvancedDefaultDescriptor[];
}

export interface TypeDefinition {

    id: string;

    label: string;

    hasLength?: boolean;

    hasPrecision?: boolean;

    defaultLength?: number;

    defaultPrecision?: number;

    category: TypeCategory;

    insertFunc?: string;

    tableModel?: string;
}

export interface CustomLoginFormProps {
    hostName: string;
    setHostName: (value: string) => void;
    username: string;
    setUsername: (value: string) => void;
    password: string;
    setPassword: (value: string) => void;
    advancedForm: Record<string, string>;
    setAdvancedForm: (value: Record<string, string>) => void;
}

export interface SourceTypeItem {
    id: string;
    label: string;
    connector: string;
    icon: ReactElement;
    extra: Record<string, string>;
    category?: SourceTypesQuery['SourceTypes'][number]['category'];
    traits?: SourceTraitsDescriptor;
    connectionFields?: SourceConnectionFieldDescriptor[];
    contract?: SourceContractDescriptor;
    discoveryPrefill?: SourceDiscoveryPrefillDescriptor;
    fields?: {
        hostname?: boolean;
        username?: boolean;
        password?: boolean;
        database?: boolean;
        searchPath?: boolean;
    };
    requiredFields?: {
        hostname?: boolean;
        username?: boolean;
        password?: boolean;
        database?: boolean;
    };
    operators?: string[];

    typeDefinitions?: TypeDefinition[];

    aliasMap?: Record<string, string>;

    supportsModifiers?: boolean;

    supportsScratchpad?: boolean;

    supportsChat?: boolean;

    supportsGraph?: boolean;

    supportsSchema?: boolean;

    supportsDatabaseSwitching?: boolean;

    usesSchemaForGraph?: boolean;

    usesDatabaseInsteadOfSchema?: boolean;

    supportsMockData?: boolean;

    singularStorageUnitLabel?: string;

    storageUnitLabel?: string;

    isAwsManaged?: boolean;

    sslModes?: SSLModeOption[];

    customFormRenderer?: ComponentType<CustomLoginFormProps>;
}

export type SourceTypeOverride = Pick<SourceTypeItem, 'id'> &
    Partial<Omit<SourceTypeItem, 'id'>>;

export interface SourceTypeFilterOptions {

    cloudProvidersEnabled?: boolean;
}

export type BackendSourceType = SourceTypesQuery['SourceTypes'][number];

const shouldUseCatalogCache = !import.meta.env.DEV;

function getCatalogCacheKey(): string {
    return `whodb_source_catalog_${getEdition()}_${__APP_VERSION__}`;
}

export function readCachedSourceCatalog(): BackendSourceType[] {
    if (!shouldUseCatalogCache) {
        return [];
    }

    try {
        const raw = localStorage.getItem(getCatalogCacheKey());
        if (!raw) {
            return [];
        }

        const parsed = JSON.parse(raw) as BackendSourceType[];
        return Array.isArray(parsed) ? parsed : [];
    } catch {
        return [];
    }
}

export function writeCachedSourceCatalog(items: BackendSourceType[]): void {
    if (!shouldUseCatalogCache) {
        return;
    }

    try {
        localStorage.setItem(getCatalogCacheKey(), JSON.stringify(items));
    } catch {

    }
}

function findConnectionField(
    connectionFields: BackendSourceType['connectionFields'],
    key: string
): BackendSourceType['connectionFields'][number] | undefined {
    return connectionFields.find(field => field.Key.toLowerCase() === key.toLowerCase());
}

function mapAdvancedDefaults(
    connectionFields: BackendSourceType['connectionFields']
): Record<string, string> {
    return connectionFields.reduce<Record<string, string>>((acc, field) => {
        if (field.Section === SourceConnectionFieldSection.Advanced) {
            acc[field.Key] = field.DefaultValue ?? "";
        }
        return acc;
    }, {});
}

function resolveIcon(sourceType: string, connector: string): ReactElement {
    const logos = Icons.Logos as Record<string, ReactElement>;
    return logos[sourceType] ?? logos[connector] ?? createElement("span", { className: "w-6 h-6" });
}

function decorateSourceType(item: BackendSourceType): SourceTypeItem {
    const hostnameField = findConnectionField(item.connectionFields, "Hostname");
    const usernameField = findConnectionField(item.connectionFields, "Username");
    const passwordField = findConnectionField(item.connectionFields, "Password");
    const databaseField = findConnectionField(item.connectionFields, "Database");
    const searchPathField = findConnectionField(item.connectionFields, "Search Path");
    const browsePath = item.contract.BrowsePath;
    const objectTypes = item.contract.ObjectTypes;
    const defaultObjectType = objectTypes.find(objectType => objectType.Kind === item.contract.DefaultObjectKind);

    return {
        id: item.id,
        label: item.label,
        connector: item.connector,
        icon: resolveIcon(item.id, item.connector),
        extra: mapAdvancedDefaults(item.connectionFields),
        category: item.category,
        traits: item.traits,
        connectionFields: item.connectionFields,
        contract: item.contract,
        discoveryPrefill: item.discoveryPrefill,
        fields: {
            hostname: hostnameField != null,
            username: usernameField != null,
            password: passwordField != null,
            database: databaseField != null,
            searchPath: searchPathField != null,
        },
        requiredFields: {
            hostname: hostnameField?.Required ?? false,
            username: usernameField?.Required ?? false,
            password: passwordField?.Required ?? false,
            database: databaseField?.Required ?? false,
        },
        supportsModifiers: item.contract.Model === SourceModel.Relational,
        supportsChat: item.contract.Surfaces.includes(SourceSurface.Chat),
        supportsGraph: item.contract.Surfaces.includes(SourceSurface.Graph),
        supportsScratchpad: item.contract.Surfaces.includes(SourceSurface.Query),
        supportsSchema: browsePath.includes(SourceObjectKind.Schema),
        supportsDatabaseSwitching: browsePath.includes(SourceObjectKind.Database),
        usesSchemaForGraph: item.contract.GraphScopeKind === SourceObjectKind.Schema,
        usesDatabaseInsteadOfSchema:
            item.contract.GraphScopeKind === SourceObjectKind.Database ||
            (!browsePath.includes(SourceObjectKind.Schema) && browsePath.includes(SourceObjectKind.Database)),
        supportsMockData: objectTypes.some(objectType => objectType.Actions.includes(SourceAction.GenerateMockData)),
        isAwsManaged: item.isAwsManaged,
        singularStorageUnitLabel: defaultObjectType?.SingularLabel,
        storageUnitLabel: defaultObjectType?.PluralLabel,
        sslModes: item.sslModes.map(mode => ({
            value: mode.value,
            aliases: mode.aliases.length > 0 ? mode.aliases : undefined,
        })),
    };
}

function filterSourceTypes(
    items: SourceTypeItem[],
    options: SourceTypeFilterOptions = {}
): SourceTypeItem[] {
    const { cloudProvidersEnabled = true } = options;

    if (cloudProvidersEnabled) {
        return items;
    }

    return items.filter(item => !item.isAwsManaged);
}

function mergeSourceTypeOverride(
    item: SourceTypeItem,
    override: SourceTypeOverride
): SourceTypeItem {
    return {
        ...item,
        ...override,
        extra: override.extra ? { ...item.extra, ...override.extra } : item.extra,
        traits: override.traits
            ? {
                connection: { ...(item.traits?.connection ?? {}), ...override.traits.connection },
                presentation: { ...(item.traits?.presentation ?? {}), ...override.traits.presentation },
                query: { ...(item.traits?.query ?? {}), ...override.traits.query },
                mockData: { ...(item.traits?.mockData ?? {}), ...override.traits.mockData },
            }
            : item.traits,
        fields: override.fields ? { ...item.fields, ...override.fields } : item.fields,
        requiredFields: override.requiredFields ? { ...item.requiredFields, ...override.requiredFields } : item.requiredFields,
    };
}

function withRegisteredSourceTypes(baseTypes: SourceTypeItem[]): SourceTypeItem[] {
    const overrides = getRegisteredSourceTypeOverrides();
    if (overrides.length === 0) {
        return baseTypes;
    }

    return baseTypes.map(item => {
        const override = overrides.find(candidate => candidate.id === item.id);
        return override ? mergeSourceTypeOverride(item, override) : item;
    });
}

export function resolveSourceTypeItems(
    catalog: BackendSourceType[],
    options: SourceTypeFilterOptions = {}
): SourceTypeItem[] {
    const items = withRegisteredSourceTypes(catalog.map(decorateSourceType));
    return filterSourceTypes(items, options);
}

export function findSourceTypeItem(
    items: SourceTypeItem[],
    sourceType: string | undefined
): SourceTypeItem | undefined {
    if (!sourceType) {
        return undefined;
    }

    return items.find(item => item.id === sourceType);
}

export function resolveSourceConnector(
    sourceType: string | undefined,
    item: SourceTypeItem | undefined
): string | undefined {
    if (!sourceType) {
        return undefined;
    }

    return item?.connector ?? sourceType;
}

export function findSourceObjectType(
    item: SourceTypeItem | undefined,
    kind: SourceObjectKind | undefined | null
): SourceObjectTypeDescriptor | undefined {
    if (!item?.contract || kind == null) {
        return undefined;
    }

    return item.contract.ObjectTypes.find(objectType => objectType.Kind === kind);
}

export function getDefaultSourceObjectType(
    item: SourceTypeItem | undefined
): SourceObjectTypeDescriptor | undefined {
    return findSourceObjectType(item, item?.contract?.DefaultObjectKind);
}