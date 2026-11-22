import { skipToken, useQuery } from "@apollo/client/react";
import {
    SearchInput,
    Sidebar as SidebarComponent,
    SidebarContent,
    SidebarGroup,
    SidebarHeader,
    toTitleCase,
    Tree,
    TreeDataItem,
} from "@clidey/ux";
import {GetStorageUnitsDocument, GetStorageUnitsQuery, SourceAction, type SourceObjectRefInput} from "@graphql";
import {FolderIcon, TableCellsIcon} from "./heroicons";
import {FC, useCallback, useEffect, useMemo, useState} from "react";
import {useLocation, useNavigate} from "react-router-dom";
import {InternalRoutes} from "../config/routes";
import {useSourceContract} from "../hooks/useSourceContract";
import {useAppSelector} from "../store/hooks";
import {Loading} from "./loading";
import {useTranslation} from "@/hooks/use-translation";
import {buildSourceParentRef} from "@/utils/source-refs";

type SourceBrowserObject = GetStorageUnitsQuery['StorageUnit'][number];

type SchemaViewerProps = {
    parentRef?: SourceObjectRefInput;
    selectedName?: string;
    trail?: SourceBrowserObject[];
};

function groupByType(units: SourceBrowserObject[]) {
    const groups: Record<string, any[]> = {};
    for (const unit of units) {
        const type = toTitleCase(unit.Attributes.find(a => a.Key === "Type")?.Value ?? "");
        if (type === "") continue;
        if (!groups[type]) groups[type] = [];
        groups[type].push(unit);
    }
    return groups;
}

export const SchemaViewer: FC<SchemaViewerProps> = ({ parentRef: explicitParentRef, selectedName, trail = [] }) => {
    const { t } = useTranslation('components/schema-viewer');
    const current = useAppSelector(state => state.auth.current);
    const selectedSchema = useAppSelector(state => state.database.schema);
    const { item, storageUnitLabel, supportsSchema } = useSourceContract(current?.Type);
    const navigate = useNavigate();
    const state = useLocation().state as { unit?: SourceBrowserObject } | undefined;

    const [search, setSearch] = useState("");

    const parentRef = explicitParentRef ?? buildSourceParentRef(item, current, selectedSchema);
    const storageUnitsQueryOptions = current && (explicitParentRef != null || !supportsSchema || selectedSchema !== "")
        ? {
            variables: {
                parent: parentRef,
            },
        }
        : skipToken;

    const {data, loading, refetch} = useQuery(GetStorageUnitsDocument, storageUnitsQueryOptions);

    const currentProfileId = current?.Id;
    const currentDatabase = current?.Database;
    useEffect(() => {
        if (currentProfileId) {
            refetch();
        }
    }, [currentProfileId, currentDatabase, refetch]);

    const storageUnits = useMemo(() => {
        return (data?.StorageUnit ?? []) as SourceBrowserObject[];
    }, [data?.StorageUnit]);

    const treeData: TreeDataItem[] = useMemo(() => {
        if (storageUnits.length === 0) return [];
        const grouped = groupByType(storageUnits);

        if (search.trim() !== "") {
            const searchLower = search.trim().toLowerCase();

            const filteredUnits = storageUnits.filter(unit =>
                (unit.Name ?? "").toLowerCase().includes(searchLower)
            );
            const filteredGrouped = groupByType(filteredUnits);
            return Object.entries(filteredGrouped).map(([type, units]) => ({
                id: type,
                name: type,
                icon: FolderIcon as TreeDataItem["icon"],
                children: units.map(unit => ({
                    id: unit.Name,
                    name: unit.Name,
                    icon: TableCellsIcon as TreeDataItem["icon"],
                })),
            }));
        }

        return Object.entries(grouped).map(([type, units]) => ({
            id: type,
            name: type,
            icon: FolderIcon as TreeDataItem["icon"],
            children: units.map(unit => ({
                id: unit.Name,
                name: unit.Name,
                icon: TableCellsIcon as TreeDataItem["icon"],
            })),
        }));
    }, [search, storageUnits]);

    const handleSelect = useCallback((item: TreeDataItem | undefined) => {

        const tableId = item?.id;
        if (tableId == null || tableId === (selectedName ?? state?.unit?.Name)) {
            return
        }
        const unit = storageUnits.find(u => u.Name === tableId);
        if (unit == null) {
            return;
        }
        if (unit.Actions.includes(SourceAction.Browse) && unit.HasChildren) {
            navigate(InternalRoutes.Dashboard.StorageUnit.path, {
                state: {
                    parent: unit,
                    trail: [...trail, unit],
                },
            });
            return;
        }
        navigate(InternalRoutes.Dashboard.ExploreStorageUnit.path, {
            state: {
                unit,
                parentRef,
                trail,
            },
        });
    }, [navigate, parentRef, selectedName, state?.unit?.Name, storageUnits, trail]);

    if (storageUnits.length === 0 && search.trim() === "") {
        return null;
    }

    return (
        <div className="flex h-full dark" data-testid="schema-viewer">
            <SidebarComponent variant="embed" className="w-64 h-full flex flex-col">
                <SidebarContent>
                    <SidebarHeader>
                        <h1 className="text-lg font-semibold pt-8 px-4">
                            {storageUnitLabel}
                        </h1>
                    </SidebarHeader>
                    <div className="px-4">
                        <SearchInput
                            value={search}
                            onChange={e => setSearch(e.target.value)}
                            placeholder={t('searchTables')}
                            aria-label={t('searchTables')}
                        />
                    </div>
                    <SidebarGroup>
                        {
                            loading ? (
                                <div className="flex-1 flex items-center justify-center">
                                    <Loading />
                                </div>
                            ) : treeData.length === 0 ? (
                                <div className="flex-1 flex items-center justify-center px-4 text-center text-sm text-muted-foreground mt-4">
                                    {t('noResults')}
                                </div>
                            ) : (
                                <Tree
                                    className="flex-1 overflow-y-auto"
                                    data={treeData}
                                    initialSelectedItemId={selectedName ?? state?.unit?.Name}
                                    onSelectChange={handleSelect}
                                    expandAll
                                />
                            )
                        }
                    </SidebarGroup>
                </SidebarContent>
            </SidebarComponent>
        </div>
    );
};