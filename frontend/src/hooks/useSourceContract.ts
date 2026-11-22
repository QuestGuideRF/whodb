import { useMemo } from "react";
import {
    SourceConnectionTransport,
    SourceHostInputMode,
    SourceHostInputUrlParser,
    SourceModel,
    SourceQueryExplainMode,
    SourceProfileLabelStrategy,
    SourceSchemaFidelity,
} from "@graphql";
import { SourceTypeItem } from "../config/source-types";
import { resolveSourceContractFlags, type SourceContractFlags } from "../utils/source-contract-flags";
import { useSourceTypeItem } from "./useSourceCatalog";

export interface SourceContractState extends SourceContractFlags {

    item?: SourceTypeItem;

    connector?: string;

    connectionTransport: SourceConnectionTransport;

    hostInputMode: SourceHostInputMode;

    hostInputUrlParser: SourceHostInputUrlParser;

    profileLabelStrategy: SourceProfileLabelStrategy;

    schemaFidelity: SourceSchemaFidelity;

    supportsAnalyze: boolean;

    explainMode: SourceQueryExplainMode;

    supportsMockDataRelations: boolean;

    loading: boolean;

    isNoSQL: boolean;

    storageUnitLabel: string;

    singularStorageUnitLabel: string;
}

export function useSourceContract(sourceType: string | undefined): SourceContractState {
    const { item, loading } = useSourceTypeItem(sourceType);

    return useMemo(() => {
        const connector = item?.connector ?? sourceType;
        const featureFlags = resolveSourceContractFlags(item);
        const contract = item?.contract;
        const defaultObjectKind = contract?.DefaultObjectKind;
        const defaultObjectType = contract?.ObjectTypes.find(objectType => objectType.Kind === defaultObjectKind);
        const model = contract?.Model;
        const traits = item?.traits;

        return {
            item,
            connector,
            connectionTransport: traits?.connection.transport ?? SourceConnectionTransport.Network,
            hostInputMode: traits?.connection.hostInputMode ?? SourceHostInputMode.None,
            hostInputUrlParser: traits?.connection.hostInputUrlParser ?? SourceHostInputUrlParser.None,
            profileLabelStrategy: traits?.presentation.profileLabelStrategy ?? SourceProfileLabelStrategy.Default,
            schemaFidelity: traits?.presentation.schemaFidelity ?? SourceSchemaFidelity.Exact,
            supportsAnalyze: traits?.query.supportsAnalyze ?? false,
            explainMode: traits?.query.explainMode ?? SourceQueryExplainMode.None,
            supportsMockDataRelations: traits?.mockData.supportsRelationalDependencies ?? true,
            loading,
            isNoSQL: model != null && model !== SourceModel.Relational,
            storageUnitLabel: item?.storageUnitLabel ?? defaultObjectType?.PluralLabel ?? "Storage Units",
            singularStorageUnitLabel: item?.singularStorageUnitLabel ?? defaultObjectType?.SingularLabel ?? "Storage Unit",
            ...featureFlags,
        };
    }, [sourceType, item, loading]);
}