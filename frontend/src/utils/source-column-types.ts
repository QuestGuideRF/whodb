import { TypeDefinition } from '../config/source-types';
import { getSourceSessionMetadataState } from './source-session-metadata-cache';

export function getSourceColumnTypeDefinitions(sourceType: string): TypeDefinition[] {
    const metadataState = getSourceSessionMetadataState();

    if (
        metadataState.sourceType === sourceType &&
        metadataState.typeDefinitions.length > 0
    ) {
        return metadataState.typeDefinitions.map(td => ({
            id: td.id,
            label: td.label,
            hasLength: td.hasLength || undefined,
            hasPrecision: td.hasPrecision || undefined,
            defaultLength: td.defaultLength ?? undefined,
            defaultPrecision: td.defaultPrecision ?? undefined,
            category: td.category,
        }));
    }

    console.warn(
        `[source-column-types] No type definitions found for ${sourceType}. ` +
            `Ensure SourceSessionMetadata query has completed.`
    );
    return [];
}

export function getSourceColumnTypeAliasMap(sourceType: string): Record<string, string> {
    const metadataState = getSourceSessionMetadataState();

    if (
        metadataState.sourceType === sourceType &&
        Object.keys(metadataState.aliasMap).length > 0
    ) {
        return metadataState.aliasMap;
    }

    if (metadataState.sourceType !== sourceType) {
        console.warn(
            `[source-column-types] No alias map found for ${sourceType}. ` +
                `Ensure SourceSessionMetadata query has completed.`
        );
    }
    return {};
}

export function normalizeColumnTypeName(typeName: string, sourceType: string): string {
    const baseType = typeName.replace(/\(.*\)$/, '').trim().toUpperCase();
    const aliasMap = getSourceColumnTypeAliasMap(sourceType);

    return aliasMap[baseType] ?? baseType;
}

export function findColumnTypeDefinition(typeId: string, sourceType: string): TypeDefinition | undefined {
    const typeDefs = getSourceColumnTypeDefinitions(sourceType);
    const normalizedType = normalizeColumnTypeName(typeId, sourceType);

    let typeDef = typeDefs.find(t => t.id.toUpperCase() === normalizedType);

    if (!typeDef) {
        typeDef = typeDefs.find(t => t.id.toUpperCase() === typeId.toUpperCase());
    }

    return typeDef;
}

export function parseTypeSpec(fullType: string): {
    baseType: string;
    length?: number;
    precision?: number;
    scale?: number;
} {
    const match = fullType.match(/^([A-Za-z0-9_ ]+)(?:\((\d+)(?:,\s*(\d+))?\))?$/);

    if (!match) {
        return { baseType: fullType.trim() };
    }

    const baseType = match[1].trim();
    const firstNum = match[2] ? parseInt(match[2], 10) : undefined;
    const secondNum = match[3] ? parseInt(match[3], 10) : undefined;

    if (secondNum !== undefined) {
        return { baseType, precision: firstNum, scale: secondNum };
    }

    return { baseType, length: firstNum };
}

export function formatTypeSpec(
    baseType: string,
    length?: number,
    precision?: number,
    scale?: number
): string {
    if (precision !== undefined) {
        if (scale !== undefined) {
            return `${baseType}(${precision},${scale})`;
        }
        return `${baseType}(${precision})`;
    }
    if (length !== undefined) {
        return `${baseType}(${length})`;
    }
    return baseType;
}