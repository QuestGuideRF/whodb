import { SourceObjectKind, type SourceObjectRefInput } from '@graphql';
import type { SourceTypeItem } from '../config/source-types';
import type { LocalLoginProfile } from '../store/auth';

function buildPathThroughIndex(
    item: SourceTypeItem | undefined,
    profile: LocalLoginProfile | undefined,
    endIndex: number,
    schema?: string,
): string[] | undefined {
    if (!item?.contract || endIndex < 0) {
        return [];
    }

    const path: string[] = [];
    for (let index = 0; index <= endIndex; index += 1) {
        const kind = item.contract.BrowsePath[index];
        if (kind == null) {
            return undefined;
        }

        if (kind === SourceObjectKind.Database) {
            if (!profile?.Database) {
                return undefined;
            }
            path.push(profile.Database);
            continue;
        }

        if (kind === SourceObjectKind.Schema) {
            if (!schema) {
                return undefined;
            }
            path.push(schema);
            continue;
        }

        return undefined;
    }

    return path;
}

function buildRefForKind(
    item: SourceTypeItem | undefined,
    profile: LocalLoginProfile | undefined,
    kind: SourceObjectKind | undefined | null,
    schema?: string,
): SourceObjectRefInput | undefined {
    if (!item?.contract || kind == null) {
        return undefined;
    }

    const targetIndex = item.contract.BrowsePath.indexOf(kind);
    if (targetIndex < 0) {
        return undefined;
    }

    const path = buildPathThroughIndex(item, profile, targetIndex, schema);
    if (path == null) {
        return undefined;
    }

    return {
        Kind: kind,
        Path: path,
    };
}

function buildParentKind(
    item: SourceTypeItem | undefined,
    pathLength: number
): SourceObjectKind | undefined {
    if (!item?.contract || pathLength <= 0) {
        return undefined;
    }

    const maxParentIndex = item.contract.BrowsePath.length - 2;
    if (maxParentIndex < 0) {
        return undefined;
    }

    let parentIndex = pathLength - 1;
    if (parentIndex > maxParentIndex) {
        parentIndex = maxParentIndex;
    }
    if (parentIndex < 0 || parentIndex >= item.contract.BrowsePath.length) {
        return undefined;
    }

    return item.contract.BrowsePath[parentIndex];
}

export function buildSourceParentRef(
    item: SourceTypeItem | undefined,
    profile: LocalLoginProfile | undefined,
    schema?: string,
): SourceObjectRefInput | undefined {
    if (!item?.contract) {
        return undefined;
    }

    const defaultIndex = item.contract.BrowsePath.indexOf(item.contract.DefaultObjectKind);
    if (defaultIndex <= 0) {
        return undefined;
    }

    return buildRefForKind(item, profile, item.contract.BrowsePath[defaultIndex - 1], schema);
}

export function buildSourceObjectRef(
    item: SourceTypeItem | undefined,
    profile: LocalLoginProfile | undefined,
    schema: string | undefined,
    objectName: string,
): SourceObjectRefInput {
    const parent = buildSourceParentRef(item, profile, schema);
    return {
        Kind: item?.contract?.DefaultObjectKind ?? SourceObjectKind.Table,
        Path: [...(parent?.Path ?? []), objectName],
    };
}

export function buildSourceParentObjectRef(
    item: SourceTypeItem | undefined,
    ref: Pick<SourceObjectRefInput, 'Kind' | 'Path' | 'Locator'>
): SourceObjectRefInput | undefined {
    if (ref.Path.length <= 1) {
        return undefined;
    }

    const parentPath = ref.Path.slice(0, -1);
    return {
        Kind: buildParentKind(item, parentPath.length) ?? ref.Kind,
        Path: parentPath,
    };
}

export function buildSourceScopeRef(
    item: SourceTypeItem | undefined,
    profile: LocalLoginProfile | undefined,
    schema?: string,
): SourceObjectRefInput | undefined {
    return buildRefForKind(item, profile, item?.contract?.GraphScopeKind, schema)
        ?? buildSourceParentRef(item, profile, schema);
}

export function buildSourceSchemaQuery(
    item: SourceTypeItem | undefined,
    profile: LocalLoginProfile | undefined,
): { parent?: SourceObjectRefInput; kinds: SourceObjectKind[] } {
    if (!item?.contract) {
        return { kinds: [] };
    }

    const schemaIndex = item.contract.BrowsePath.indexOf(SourceObjectKind.Schema);
    if (schemaIndex < 0) {
        return { kinds: [] };
    }
    if (schemaIndex === 0) {
        return { kinds: [SourceObjectKind.Schema] };
    }

    const path = buildPathThroughIndex(item, profile, schemaIndex - 1);
    if (path == null) {
        return { kinds: [] };
    }

    return {
        parent: {
            Kind: item.contract.BrowsePath[schemaIndex - 1] as SourceObjectKind,
            Path: path,
        },
        kinds: [SourceObjectKind.Schema],
    };
}

export function getObjectNameFromRef(ref: { Path: string[] }): string {
    return ref.Path[ref.Path.length - 1] ?? '';
}