import { getSourceSessionMetadataState } from './source-session-metadata-cache';

export function getSourceOperators(sourceType: string): string[] {
    const metadataState = getSourceSessionMetadataState();

    if (
        metadataState.sourceType === sourceType &&
        metadataState.operators.length > 0
    ) {
        return metadataState.operators;
    }

    console.warn(
        `[source-operators] No operators found for ${sourceType}. ` +
            `Ensure SourceSessionMetadata query has completed.`
    );
    return [];
}