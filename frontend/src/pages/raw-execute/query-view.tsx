import {useLazyQuery} from "@apollo/client/react";
import React, {FC, useEffect} from "react";
import {StorageUnitTable} from "../../components/table";
import {RawExecuteDocument} from "../../generated/graphql";
import {CheckCircleIcon} from "../../components/heroicons";
import {useAppSelector} from "../../store/hooks";

type PromiseFunction = (code: string) => Promise<any>;

export type IPluginProps = {
    code: string;
    handleExecuteRef: React.MutableRefObject<PromiseFunction | null>;
    providerId?: string;
    modelType: string;
    token?: string;
    schema: string;
    containerWidth?: number;
}

function isSQLQueryAction(code?: string): boolean {
    if (code == null) {
        return true;
    }

    const cleaned = code
        .split("\n")
        .filter((text: string) => !text.trim().startsWith("--"))
        .join("\n")
        .trim()
        .toLowerCase();

    return /^(select|with|values|show|explain|describe)\b/.test(cleaned);
}

export const QueryView: FC<IPluginProps> = ({ code, handleExecuteRef, containerWidth }) => {
    const [rawExecute, { data }] = useLazyQuery(RawExecuteDocument, {
        fetchPolicy: 'network-only',
    });
    const current = useAppSelector(state => state.auth.current);

    useEffect(() => {
        handleExecuteRef.current = async (code: string) => {
            const result = await rawExecute({
                variables: { query: code },
            });
            if (result.error) {
                throw result.error;
            }
            const data = result.data;
            if (isSQLQueryAction(code) || (data?.RawExecute?.Rows?.length || 0) > 0) {
                return data?.RawExecute ?? null;
            }
            return null;
        };
    }, [rawExecute, handleExecuteRef, code]);

    if (data == null) {
        return null;
    }

    if (isSQLQueryAction(code) || data.RawExecute.Rows.length > 0) {
        return (
            <div className="flex flex-col w-full" data-testid="cell-query-output">
                {
                    data.RawExecute.Columns.length > 0 && (
                        <StorageUnitTable
                            key={containerWidth}
                            columns={data.RawExecute.Columns.map((c: any) => c.Name)}
                            columnTypes={data.RawExecute.Columns.map((c: any) => c.Type)}
                            rows={data.RawExecute.Rows}
                            disableEdit={true}
                            limitContextMenu={true}
                            height={250}
                            databaseType={current?.Type}
                            rawQuery={code}
                            totalCount={data.RawExecute.TotalCount}
                        />
                    )
                }
            </div>
        );
    }

    return (
        <div className="bg-white/10 text-neutral-800 dark:text-neutral-300 rounded-lg p-2 flex gap-sm self-start items-center my-4" data-testid="cell-action-output">
            Action Executed
            <CheckCircleIcon className="w-4 h-4" />
        </div>
    );
};