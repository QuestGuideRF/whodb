import {useLazyQuery, useMutation, useQuery} from "@apollo/client/react";
import {
    Alert,
    AlertDescription,
    AlertDialog,
    AlertDialogAction,
    AlertDialogCancel,
    AlertDialogContent,
    AlertDialogDescription,
    AlertDialogFooter,
    AlertDialogHeader,
    AlertDialogTitle,
    AlertTitle,
    Button,
    Checkbox,
    cn,
    ContextMenu,
    ContextMenuContent,
    ContextMenuItem,
    ContextMenuSeparator,
    ContextMenuShortcut,
    ContextMenuSub,
    ContextMenuSubContent,
    ContextMenuSubTrigger,
    ContextMenuTrigger,
    EmptyState,
    Input,
    Label,
    Pagination,
    PaginationContent,
    PaginationEllipsis,
    PaginationItem,
    PaginationLink,
    PaginationNext,
    PaginationPrevious,
    Select,
    SelectContent,
    SelectItem,
    SelectTrigger,
    SelectValue,
    Sheet,
    SheetContent,
    SheetFooter,
    SheetTitle,
    Spinner,
    Table as TableComponent,
    TableCell,
    TableHead,
    TableHeader,
    TableHeadRow,
    TableRow,
    TextArea,
    toast,
    VirtualizedTableBody
} from "@clidey/ux";
import {
    AnalyzeMockDataDependenciesDocument,
    DeleteRowDocument,
    GenerateMockDataDocument,
    MockDataMaxRowCountDocument,
    type SourceObjectRefInput,
} from '@graphql';
import {FC, Suspense, useCallback, useEffect, useMemo, useRef, useState} from "react";
import {Export} from "./export";
import {ImportData} from "./import-data";
import {useTranslation} from '@/hooks/use-translation';
import {copyToClipboard} from '@/services/clipboard';
import {useSourceContract} from "@/hooks/useSourceContract";
import {
    ArrowDownCircleIcon,
    ArrowDownTrayIcon,
    ArrowUpCircleIcon,
    CalculatorIcon,
    CalendarIcon,
    CheckCircleIcon,
    ChevronDownIcon,
    ChevronUpIcon,
    CircleStackIcon,
    ClockIcon,
    CodeBracketIcon,
    CursorArrowRaysIcon,
    DocumentDuplicateIcon,
    DocumentIcon,
    DocumentTextIcon,
    EllipsisVerticalIcon,
    GlobeAltIcon,
    HashtagIcon,
    KeyIcon,
    ListBulletIcon,
    MagnifyingGlassIcon,
    PencilSquareIcon,
    ShareIcon,
    Squares2X2Icon,
    TrashIcon,
    XMarkIcon
} from "./heroicons";
import {Tip} from "./tip";
import {formatShortcut} from "@/utils/platform";
import {matchesShortcut, SHORTCUTS} from "@/utils/shortcuts";
import {formatNumber} from "@/utils/functions";

const EEExport = null;

const DynamicExport: FC<{
    open: boolean;
    onOpenChange: (open: boolean) => void;
    storageUnit: string;
    objectRef?: SourceObjectRefInput;
    hasSelectedRows: boolean;
    selectedRowsData?: Record<string, any>[];
    checkedRowsCount: number;
    databaseType?: string;
    rawQuery?: string;
    preselectedFormat?: 'csv' | 'excel' | 'ndjson';
    forceExportAll?: boolean;
}> = (props) => {

    const ExportComponent = EEExport || Export;
    return <ExportComponent {...props} />;
};

const stringTypes = new Set([
    "TEXT", "STRING", "VARCHAR", "CHAR",
    "CHARACTER VARYING", "CHARACTER",
    "FIXEDSTRING",
]);
const intTypes = new Set([
    "INTEGER", "SMALLINT", "BIGINT", "INT", "TINYINT", "MEDIUMINT",
    "INT2", "INT4", "INT8",
    "INT16", "INT32", "INT64", "INT128", "INT256",
    "SERIAL", "BIGSERIAL", "SMALLSERIAL",
]);
const uintTypes = new Set([
    "TINYINT UNSIGNED", "SMALLINT UNSIGNED", "MEDIUMINT UNSIGNED", "BIGINT UNSIGNED",
    "UINT8", "UINT16", "UINT32", "UINT64", "UINT128", "UINT256",
]);
const floatTypes = new Set([
    "REAL", "NUMERIC", "DOUBLE PRECISION", "FLOAT", "NUMBER", "DOUBLE", "DECIMAL",
    "FLOAT4", "FLOAT8",
    "FLOAT32", "FLOAT64",
    "DECIMAL32", "DECIMAL64", "DECIMAL128", "DECIMAL256",
    "MONEY",
]);
const boolTypes = new Set([
    "BOOLEAN", "BIT", "BOOL",
]);
const dateTypes = new Set([
    "DATE",
    "DATE32",
]);
const dateTimeTypes = new Set([
    "DATETIME", "TIMESTAMP", "TIME",
    "TIMESTAMP WITH TIME ZONE", "TIMESTAMP WITHOUT TIME ZONE",
    "TIME WITH TIME ZONE", "TIME WITHOUT TIME ZONE",
    "DATETIME2", "SMALLDATETIME",
    "TIMETZ", "TIMESTAMPTZ",
    "INTERVAL",
    "DATETIME64",
    "YEAR",
]);
const uuidTypes = new Set([
    "UUID",
]);
const binaryTypes = new Set([
    "BLOB", "BYTEA", "VARBINARY", "BINARY", "IMAGE",
    "TINYBLOB", "MEDIUMBLOB", "LONGBLOB",
]);
const jsonTypes = new Set([
    "JSON", "JSONB",
]);
const networkTypes = new Set([
    "CIDR", "INET", "MACADDR", "MACADDR8",
    "IPV4", "IPV6",
]);
const geometryTypes = new Set([
    "POINT", "LINE", "LSEG", "BOX", "PATH", "POLYGON", "CIRCLE",
    "GEOMETRY", "GEOGRAPHY",
    "LINESTRING", "MULTIPOINT", "MULTILINESTRING", "MULTIPOLYGON", "GEOMETRYCOLLECTION",
]);
const xmlTypes = new Set([
    "XML",
]);

function stripTypeSuffix(type: string): string {
    return type.replace(/\(.*\)$/, '').trim();
}

export function getColumnIcons(columns: string[], columnTypes?: string[], t?: (key: string) => string) {
    return columns.map((col, idx) => {
        const rawType = columnTypes?.[idx] || "";

        const type = stripTypeSuffix(rawType).toUpperCase();

        if (intTypes.has(type) || uintTypes.has(type)) return <HashtagIcon className="w-4 h-4" aria-label={t?.('integerType') ?? 'Integer type'} />;
        if (floatTypes.has(type)) return <CalculatorIcon className="w-4 h-4" aria-label={t?.('decimalType') ?? 'Decimal type'} />;
        if (boolTypes.has(type)) return <CheckCircleIcon className="w-4 h-4" aria-label={t?.('booleanType') ?? 'Boolean type'} />;
        if (dateTypes.has(type)) return <CalendarIcon className="w-4 h-4" aria-label={t?.('dateType') ?? 'Date type'} />;
        if (dateTimeTypes.has(type)) return <ClockIcon className="w-4 h-4" aria-label={t?.('dateTimeType') ?? 'DateTime type'} />;
        if (uuidTypes.has(type)) return <KeyIcon className="w-4 h-4" aria-label={t?.('uuidType') ?? 'UUID type'} />;
        if (binaryTypes.has(type)) return <DocumentDuplicateIcon className="w-4 h-4" aria-label={t?.('binaryType') ?? 'Binary type'} />;
        if (jsonTypes.has(type)) return <CodeBracketIcon className="w-4 h-4" aria-label={t?.('jsonType') ?? 'JSON type'} />;
        if (networkTypes.has(type)) return <GlobeAltIcon className="w-4 h-4" aria-label={t?.('networkType') ?? 'Network type'} />;
        if (geometryTypes.has(type)) return <Squares2X2Icon className="w-4 h-4" aria-label={t?.('geometryType') ?? 'Geometry type'} />;
        if (xmlTypes.has(type)) return <CodeBracketIcon className="w-4 h-4" aria-label={t?.('xmlType') ?? 'XML type'} />;
        if (type.startsWith("ARRAY")) return <ListBulletIcon className="w-4 h-4" aria-label={t?.('arrayType') ?? 'Array type'} />;
        if (stringTypes.has(type)) return <DocumentTextIcon className="w-4 h-4" aria-label={t?.('textType') ?? 'Text type'} />;
        return <CircleStackIcon className="w-4 h-4" aria-label={t?.('dataType') ?? 'Data type'} />;
    });
}

export function getInputPropsForColumnType(rawType: string): {
    type?: React.HTMLInputTypeAttribute;
    step?: string;
    min?: string;
    inputMode?: 'text' | 'numeric' | 'decimal',
    onKeyDown?: (e: React.KeyboardEvent<HTMLInputElement>) => void;
} {
    const type = stripTypeSuffix(rawType).toUpperCase();

    const numOnKeyDown = (e: React.KeyboardEvent<HTMLInputElement>) => {if (e.key === "e" || e.key === "E" || e.key === "+") e.preventDefault();}

    if (intTypes.has(type)) {
        return { type: 'number', step: '1', inputMode: 'numeric',  onKeyDown: numOnKeyDown};
    }

    if (uintTypes.has(type)) {
        return { type: 'number', step: '1', min: '0', inputMode: 'numeric', onKeyDown: numOnKeyDown };
    }

    if (floatTypes.has(type)) {
        return { type: 'number', step: 'any', inputMode: 'decimal', onKeyDown: numOnKeyDown };
    }

    return { type: 'text', inputMode: 'text' };
}

interface TableProps {
    columns: string[];
    columnTypes?: string[];
    columnIsPrimary?: boolean[];
    columnIsForeignKey?: boolean[];
    rows: string[][];
    rowHeight?: number;
    height?: number;
    onRowUpdate?: (row: Record<string, string | number>, originalRow?: Record<string, string | number>) => Promise<void>;
    disableEdit?: boolean;
    allowRowUpdate?: boolean;
    allowRowDelete?: boolean;
    limitContextMenu?: boolean;
    schema?: string;
    storageUnit?: string;
    objectRef?: SourceObjectRefInput;
    onRefresh?: () => void;
    children?: React.ReactNode;
    onColumnSort?: (column: string) => void;
    sortedColumns?: Map<string, 'asc' | 'desc'>;
    searchRef?: React.MutableRefObject<(search: string) => void>;
    pageSize?: number;

    totalCount?: number;
    currentPage?: number;
    onPageChange?: (page: number) => void;
    showPagination?: boolean;

    isValidForeignKey?: (columnName: string) => boolean;
    onEntitySearch?: (columnName: string, value: string) => void;
    databaseType?: string;

    isMockDataGenerationAllowed?: boolean;

    allowImport?: boolean;
    rawQuery?: string;

    enforceMinHeight?: boolean;

    enableKeyboardShortcuts?: boolean;
}

export const StorageUnitTable: FC<TableProps> = ({
    columns,
    columnTypes,
    columnIsPrimary,
    columnIsForeignKey,
    rows,
    rowHeight = 48,
    height = 500,
    onRowUpdate,
    disableEdit = false,
    allowRowUpdate = true,
    allowRowDelete = true,
    limitContextMenu = false,
    schema,
    storageUnit,
    objectRef,
    onRefresh,
    children,
    onColumnSort,
    sortedColumns,
    searchRef,
    pageSize = 100,

    totalCount,
    currentPage: serverCurrentPage,
    onPageChange,
    showPagination = false,

    isValidForeignKey,
    onEntitySearch,
    databaseType,

    isMockDataGenerationAllowed = true,

    allowImport = false,
    rawQuery,
    enforceMinHeight = false,
    enableKeyboardShortcuts = false,
}) => {
    const { t, language } = useTranslation('components/table');
    const [editIndex, setEditIndex] = useState<number | null>(null);
    const [editRow, setEditRow] = useState<string[] | null>(null);
    const [editRowInitialLengths, setEditRowInitialLengths] = useState<number[]>([]);
    const [deleting, setDeleting] = useState(false);
    const [pendingDeleteIndexes, setPendingDeleteIndexes] = useState<number[] | null>(null);
    const [checked, setChecked] = useState<number[]>([]);
    const [showExportConfirm, setShowExportConfirm] = useState(false);
    const [showImport, setShowImport] = useState(false);
    const [preselectedFormat, setPreselectedFormat] = useState<'csv' | 'excel' | 'ndjson' | undefined>(undefined);
    const [forceExportAll, setForceExportAll] = useState(false);
    const tableRef = useRef<HTMLDivElement>(null);
    const [contextMenuCellIdx, setContextMenuCellIdx] = useState<number | null>(null);

    const [focusedRowIndex, setFocusedRowIndex] = useState<number | null>(null);

    const focusedColumnRef = useRef<string | null>(null);

    const [showMockDataSheet, setShowMockDataSheet] = useState(false);
    const [mockDataRowCount, setMockDataRowCount] = useState("100");
    const [mockDataMethod, setMockDataMethod] = useState("Normal");
    const [mockDataOverwriteExisting, setMockDataOverwriteExisting] = useState("append");
    const [mockDataFkDensityRatio, setMockDataFkDensityRatio] = useState("20");
    const [showMockDataConfirmation, setShowMockDataConfirmation] = useState(false);
    const { isNoSQL, supportsImportData, supportsMockData, supportsMockDataRelations } = useSourceContract(databaseType);
    const isMockDataSupported = supportsMockData && isMockDataGenerationAllowed;
    const isImportSupported = supportsImportData;
    const { data: maxRowData } = useQuery(MockDataMaxRowCountDocument);
    const maxRowCount = maxRowData?.MockDataMaxRowCount || 200;

    const currentPage = serverCurrentPage || 1;
    const totalRows = totalCount || 0;
    const totalPages = Math.ceil(totalRows / pageSize);

    const [generateMockData, { loading: generatingMockData }] = useMutation(GenerateMockDataDocument);
    const [analyzeDependencies, { data: depAnalysis, loading: analyzingDeps }] = useLazyQuery(AnalyzeMockDataDependenciesDocument);
    const [deleteRow, ] = useMutation(DeleteRowDocument);
    const [containerWidth, setContainerWidth] = useState<number>(0);
    const lastSearchState = useRef<{ search: string; matchIdx: number }>({ search: '', matchIdx: 0 });
    const canEditRows = !disableEdit && allowRowUpdate && onRowUpdate != null;
    const canDeleteRows = !disableEdit && allowRowDelete && objectRef != null;

    const handleEdit = (index: number) => {
        if (!canEditRows) {
            return;
        }
        setEditIndex(index);
        const rowData = [...rows[index]];
        setEditRow(rowData);

        setEditRowInitialLengths(rowData.map(cell => cell?.length || 0));
    };

    const handleInputChange = (value: string, idx: number) => {
        if (editRow) {
            const updated = [...editRow];
            updated[idx] = value;
            setEditRow(updated);
        }
    };

    const handleUpdate = useCallback(() => {
        if (!canEditRows) {
            return;
        }
        if (editIndex !== null && editRow) {
            const updatedRow: Record<string, string | number> = {};
            columns.forEach((col, idx) => {
                updatedRow[col] = editRow[idx];
            });

            const originalRow: Record<string, string | number> = {};
            if (rows[editIndex]) {
                columns.forEach((col, idx) => {
                    originalRow[col] = rows[editIndex][idx];
                });
            }
            onRowUpdate?.(updatedRow, originalRow)
                .then(() => {
                    setEditIndex(null);
                    setEditRow(null);
                    setEditRowInitialLengths([]);
                    toast.success(t('rowUpdated'));
                    onRefresh?.();
                })
                .catch(() => {
                    toast.error(t('errorUpdatingRow'));
                });
        }
    }, [canEditRows, editIndex, editRow, columns, onRowUpdate, rows, onRefresh, t]);

    const hasSelectedRows = checked.length > 0;
    const selectedRowsData = useMemo(() => {
        if (hasSelectedRows) {
            const validChecked = checked.filter(idx => idx < rows.length);
            return validChecked.map(idx => {
                const row = rows[idx];
                const rowObj: Record<string, any> = {};
                columns.forEach((col, colIdx) => {
                    rowObj[col] = row[colIdx];
                });
                return rowObj;
            });
        }
        if (rawQuery) {
            return rows.map(row => {
                const rowObj: Record<string, any> = {};
                columns.forEach((col, colIdx) => {
                    rowObj[col] = row[colIdx];
                });
                return rowObj;
            });
        }
        return undefined;
    }, [hasSelectedRows, checked, rows, columns, rawQuery]);

    const openExport = useCallback((format?: 'csv' | 'excel' | 'ndjson', exportAll?: boolean) => {
        setPreselectedFormat(format);
        setForceExportAll(exportAll ?? false);
        setShowExportConfirm(true);
    }, []);

    const doDeleteRows = useCallback(async (indexesToDelete: number[]) => {
        if (!canDeleteRows) {
            return;
        }
        if (!objectRef) {
            toast.error(t('storageUnitRequired'));
            return;
        }
        let unableToDeleteAll = false;
        toast.info(t('deletingRows', { count: indexesToDelete.length }));
        for (const index of indexesToDelete) {
            const row = rows[index];
            if (!row) continue;
            const values = columns.map((col, i) => ({
                Key: col,
                Value: row[i],
            }));
            try {
                await deleteRow({
                    variables: {
                        ref: objectRef,
                        values,
                    },
                });
            } catch (e: any) {
                toast.error(t('unableToDeleteRow', { message: e?.message || e }));
                unableToDeleteAll = true;
                break;
            }
        }
        if (!unableToDeleteAll) {
            toast.success(t('rowDeleted'));
        }
        onRefresh?.();
    }, [canDeleteRows, columns, deleteRow, objectRef, onRefresh, rows, t]);

    const handleDeleteRow = useCallback((rowIndex: number) => {
        if (!canDeleteRows) {
            return;
        }
        if (!rows || !columns) return;
        let indexesToDelete: number[] = [];
        if (Array.isArray(rowIndex)) {
            indexesToDelete = rowIndex;
        } else if (typeof rowIndex === "number") {
            indexesToDelete = [rowIndex];
        }
        if (checked.length > 0) {
            indexesToDelete = [...checked];
        }
        if (indexesToDelete.length === 0) return;
        setPendingDeleteIndexes(indexesToDelete);
    }, [canDeleteRows, rows, columns, checked]);

    const handleConfirmDelete = useCallback(async () => {
        if (pendingDeleteIndexes) {
            const indexes = pendingDeleteIndexes;
            setPendingDeleteIndexes(null);
            await doDeleteRows(indexes);
        }
    }, [pendingDeleteIndexes, doDeleteRows]);

    const handleCancelDelete = useCallback(() => {
        setPendingDeleteIndexes(null);
    }, []);

    const paginatedRows = useMemo(() => {

        return rows;
    }, [rows]);

    useEffect(() => {
        setFocusedRowIndex(null);
        setChecked([]);

        const columnToFocus = focusedColumnRef.current;
        if (columnToFocus) {

            const timeoutId = setTimeout(() => {
                const header = document.querySelector(
                    `[data-testid="column-header-${columnToFocus}"]`
                ) as HTMLElement;
                if (header) {
                    header.focus();
                }
            }, 50);
            return () => clearTimeout(timeoutId);
        }
    }, [rows]);

    const handlePageChange = useCallback((newPage: number) => {
        onPageChange?.(newPage);
    }, [onPageChange]);

    const renderPaginationLinks = () => {
        const links = [];

        const start = Math.max(1, currentPage - 2);
        const end = Math.min(totalPages, currentPage + 2);

        if (start > 1) {
            links.push(
                <PaginationItem key={1}>
                    <PaginationLink href="#" onClick={e => { e.preventDefault(); handlePageChange(1); }} size="sm" data-testid="table-page-number" data-page="1" data-active={currentPage === 1} aria-label={t('goToPage', { page: 1 })}>{formatNumber(1, language)}</PaginationLink>
                </PaginationItem>
            );
            if (start > 2) {
                links.push(<PaginationEllipsis key="start-ellipsis" aria-hidden="true" />);
            }
        }

        for (let i = start; i <= end; i++) {
            links.push(
                <PaginationItem key={i}>
                    <PaginationLink
                        href="#"
                        isActive={i === currentPage}
                        onClick={e => { e.preventDefault(); handlePageChange(i); }}
                        size="sm"
                        data-testid="table-page-number"
                        data-page={i}
                        data-active={i === currentPage}
                        aria-label={i === currentPage ? t('currentPage', { page: i }) : t('goToPage', { page: i })}
                        aria-current={i === currentPage ? 'page' : undefined}
                    >
                        {formatNumber(i, language)}
                    </PaginationLink>
                </PaginationItem>
            );
        }

        if (end < totalPages) {
            if (end < totalPages - 1) {
                links.push(<PaginationEllipsis key="end-ellipsis" aria-hidden="true" />);
            }
            links.push(
                <PaginationItem key={totalPages}>
                    <PaginationLink href="#" onClick={e => { e.preventDefault(); handlePageChange(totalPages); }} size="sm" data-testid="table-page-number" data-page={totalPages} data-active={currentPage === totalPages} aria-label={t('goToPage', { page: totalPages })}>{formatNumber(totalPages, language)}</PaginationLink>
                </PaginationItem>
            );
        }

        return links;
    };

    const handleSelectRow = useCallback((rowIndex: number) => {
        const isCurrentlySelected = checked.includes(rowIndex);
        const newChecked = isCurrentlySelected ? checked.filter(i => i !== rowIndex) : [...checked, rowIndex];
        setChecked(newChecked);
    }, [checked, t]);

    const clickTimeouts = useRef<Map<string, number>>(new Map());

    const handleCellClick = useCallback((rowIndex: number, cellIndex: number) => {
        const cellKey = `${rowIndex}-${cellIndex}`;

        const existingTimeout = clickTimeouts.current.get(cellKey);
        if (existingTimeout) {
            clearTimeout(existingTimeout);
        }

        const timeout = window.setTimeout(() => {
            const cell = paginatedRows[rowIndex][cellIndex];
            if (cell !== undefined && cell !== null) {
                copyToClipboard(String(cell)).then(success => {
                    if (success) toast.success(t('copiedToClipboard'));
                });
            }
            clickTimeouts.current.delete(cellKey);
        }, 200);

        clickTimeouts.current.set(cellKey, timeout);
    }, [paginatedRows, t]);

    const handleCellDoubleClick = useCallback((rowIndex: number) => {

        for (let cellIdx = 0; cellIdx < columns.length; cellIdx++) {
            const cellKey = `${rowIndex}-${cellIdx}`;
            const timeout = clickTimeouts.current.get(cellKey);
            if (timeout) {
                clearTimeout(timeout);
                clickTimeouts.current.delete(cellKey);
            }
        }

        const row = paginatedRows[rowIndex];
        if (row && Array.isArray(row)) {
            const rowString = row.map(cell => cell ?? "").join("\t");
            copyToClipboard(rowString).then(success => {
                if (success) toast.success(t('rowCopiedToClipboard'));
            });
        }
    }, [paginatedRows, columns.length, t]);

    const handleMockDataRowCountChange = useCallback((value: string) => {

        const numericValue = value.replace(/[^0-9]/g, '');
        const parsedValue = parseInt(numericValue) || 0;

        if (parsedValue > maxRowCount) {
            setMockDataRowCount(maxRowCount.toString());
            toast.error(t('maximumRowCount', { max: maxRowCount }));
        } else {
            setMockDataRowCount(numericValue);
        }
    }, [maxRowCount, t]);

    const handleMockDataGenerate = useCallback(async () => {

        if (!storageUnit || !objectRef) {
            toast.error(t('storageUnitRequired'));
            return;
        }

        if (mockDataOverwriteExisting === "overwrite" && !showMockDataConfirmation) {
            setShowMockDataConfirmation(true);
            return;
        }

        const count = parseInt(mockDataRowCount);

        if (isNaN(count) || count < 1) {
            toast.error(t('rowCountMustBePositive'));
            return;
        }

        if (count > maxRowCount) {
            toast.error(t('rowCountExceedsMax', { max: maxRowCount }));
            return;
        }

        try {
            const result = await generateMockData({
                variables: {
                    input: {
                        Ref: objectRef,
                        RowCount: count,
                        Method: mockDataMethod,
                        OverwriteExisting: mockDataOverwriteExisting === "overwrite",
                        FkDensityRatio: parseInt(mockDataFkDensityRatio) || 20,
                    }
                }
            });

            const data = result.data?.GenerateMockData;
            if (data?.AmountGenerated) {
                toast.success(t('successfullyGenerated', { count: data.AmountGenerated }));
                setShowMockDataSheet(false);
                setShowMockDataConfirmation(false);

                if (onRefresh) {
                    onRefresh();
                }
            } else {
                toast.error(t('failedToMockData'));
            }
        } catch (error: any) {
            if (error.message === "mock data generation is not allowed for this table") {
                toast.error(t('mockDataNotAllowed'));
            } else {
                toast.error(t('mockDataFailed', { message: error.message }));
            }
        }
    }, [generateMockData, maxRowCount, mockDataFkDensityRatio, mockDataMethod, mockDataOverwriteExisting, mockDataRowCount, objectRef, onRefresh, showMockDataConfirmation, storageUnit, t]);

    const columnIcons = useMemo(() => getColumnIcons(columns, columnTypes, t), [columns, columnTypes, t]);

    useEffect(() => {
        return () => {

            clickTimeouts.current.forEach(timeout => clearTimeout(timeout));
            clickTimeouts.current.clear();
        };
    }, []);

    useEffect(() => {

        if (showMockDataSheet && objectRef) {
            const rowCount = parseInt(mockDataRowCount) || 100;
            if (rowCount > 0 && rowCount <= maxRowCount) {
                analyzeDependencies({
                    variables: {
                        ref: objectRef,
                        rowCount,
                        fkDensityRatio: null,
                    },
                });
            }
        }
    }, [analyzeDependencies, maxRowCount, mockDataRowCount, objectRef, showMockDataSheet]);

    const adjustedDepAnalysis = useMemo(() => {
        const analysis = depAnalysis?.AnalyzeMockDataDependencies;
        if (!analysis || analysis.Error || !analysis.Tables || analysis.Tables.length <= 1) {
            return analysis;
        }

        const ratio = parseInt(mockDataFkDensityRatio) || 20;
        const requestedRows = parseInt(mockDataRowCount) || 100;
        const tables = [...analysis.Tables];

        const recalculated = tables.map((t, i) => ({ ...t }));

        let childRowCount = requestedRows;
        for (let i = recalculated.length - 1; i >= 0; i--) {
            if (i === recalculated.length - 1) {

                recalculated[i] = { ...recalculated[i], RowsToGenerate: requestedRows };
            } else {

                const parentRows = Math.max(1, Math.floor(childRowCount / ratio));
                recalculated[i] = { ...recalculated[i], RowsToGenerate: parentRows };
                childRowCount = parentRows;
            }
        }

        const totalRows = recalculated.reduce((sum, t) => sum + t.RowsToGenerate, 0);

        return {
            ...analysis,
            Tables: recalculated,
            TotalRows: totalRows,
        };
    }, [depAnalysis, mockDataFkDensityRatio, mockDataRowCount]);

    useEffect(() => {
        const handleExportTrigger = () => {
            openExport();
        };

        window.addEventListener('menu:trigger-export', handleExportTrigger);
        return () => {
            window.removeEventListener('menu:trigger-export', handleExportTrigger);
        };
    }, []);

    useEffect(() => {
        const handleImportTrigger = () => {
            if (isImportSupported && allowImport) {
                setShowImport(true);
            }
        };

        window.addEventListener('menu:trigger-import', handleImportTrigger);
        return () => {
            window.removeEventListener('menu:trigger-import', handleImportTrigger);
        };
    }, [isImportSupported, allowImport]);

    useEffect(() => {
        let resizeTimeout: ReturnType<typeof setTimeout> | null = null;

        const handleResize = () => {
            if (resizeTimeout) clearTimeout(resizeTimeout);
            resizeTimeout = setTimeout(() => {
                if (onRefresh) {
                    onRefresh();
                }
            }, 300);
        };

        window.addEventListener('resize', handleResize);

        return () => {
            window.removeEventListener('resize', handleResize);
            if (resizeTimeout) clearTimeout(resizeTimeout);
        };
    }, [onRefresh]);

    const scrollRowIntoView = useCallback((rowIndex: number) => {
        const container = document.querySelector<HTMLElement>('[data-slot="table-body"]');
        if (!container) return;

        const rowTop = rowIndex * rowHeight;
        const rowBottom = rowTop + rowHeight;
        const viewTop = container.scrollTop;
        const viewBottom = viewTop + container.clientHeight;

        let target: number | null = null;
        if (rowTop < viewTop) {
            target = rowTop;
        } else if (rowBottom > viewBottom) {
            target = rowBottom - container.clientHeight;
        }

        if (target === null) return;

        const distance = Math.abs(target - viewTop);
        const behavior = distance > container.clientHeight ? 'instant' : 'smooth';
        container.scrollTo({ top: target, behavior });
    }, [rowHeight]);

    const moveFocus = useCallback((newIndex: number, extendSelection: boolean = false) => {
        if (newIndex < 0 || newIndex >= paginatedRows.length) return;

        if (extendSelection && focusedRowIndex !== null) {

            const start = Math.min(focusedRowIndex, newIndex);
            const end = Math.max(focusedRowIndex, newIndex);
            const newChecked = new Set(checked);
            for (let i = start; i <= end; i++) {
                newChecked.add(i);
            }
            setChecked(Array.from(newChecked));
        }

        setFocusedRowIndex(newIndex);
        scrollRowIntoView(newIndex);
    }, [paginatedRows.length, focusedRowIndex, checked, scrollRowIntoView]);

    const visibleRowCount = useMemo(() => {
        return Math.floor(height / rowHeight);
    }, [height, rowHeight]);

    useEffect(() => {
        if (!enableKeyboardShortcuts) return;

        const handleKeyDown = (event: KeyboardEvent) => {

            if (event.target instanceof HTMLInputElement || event.target instanceof HTMLTextAreaElement) {
                return;
            }

            if (paginatedRows.length === 0) return;

            if (matchesShortcut(event, SHORTCUTS.extendSelectDown)) {
                event.preventDefault();
                if (focusedRowIndex === null) {
                    moveFocus(0, true);
                } else {
                    moveFocus(Math.min(focusedRowIndex + 1, paginatedRows.length - 1), true);
                }
                return;
            }
            if (matchesShortcut(event, SHORTCUTS.extendSelectUp)) {
                event.preventDefault();
                if (focusedRowIndex === null) {
                    moveFocus(paginatedRows.length - 1, true);
                } else {
                    moveFocus(Math.max(focusedRowIndex - 1, 0), true);
                }
                return;
            }

            if (matchesShortcut(event, SHORTCUTS.exportData)) {
                event.preventDefault();
                openExport();
                return;
            }
            if (matchesShortcut(event, SHORTCUTS.mockData)) {
                if (isMockDataSupported) {
                    event.preventDefault();
                    setShowMockDataSheet(true);
                }
                return;
            }

            if (matchesShortcut(event, SHORTCUTS.importData)) {
                if (isImportSupported && allowImport) {
                    event.preventDefault();
                    setShowImport(true);
                }
                return;
            }
            if (matchesShortcut(event, SHORTCUTS.refresh)) {
                event.preventDefault();
                onRefresh?.();
                return;
            }
            if (matchesShortcut(event, SHORTCUTS.selectAll)) {
                event.preventDefault();
                setChecked(checked.length === paginatedRows.length ? [] : paginatedRows.map((_, index) => index));
                return;
            }
            if (matchesShortcut(event, SHORTCUTS.editRowAlt)) {
                if (focusedRowIndex !== null && canEditRows) {
                    event.preventDefault();
                    handleEdit(focusedRowIndex);
                }
                return;
            }
            if (matchesShortcut(event, SHORTCUTS.deleteRow) || matchesShortcut(event, SHORTCUTS.deleteRowAlt)) {
                if (focusedRowIndex !== null && canDeleteRows) {
                    event.preventDefault();
                    handleDeleteRow(focusedRowIndex);
                }
                return;
            }
            if (matchesShortcut(event, SHORTCUTS.nextPage)) {
                event.preventDefault();
                if (onPageChange && currentPage < totalPages) {
                    onPageChange(currentPage + 1);
                }
                return;
            }
            if (matchesShortcut(event, SHORTCUTS.prevPage)) {
                event.preventDefault();
                if (onPageChange && currentPage > 1) {
                    onPageChange(currentPage - 1);
                }
                return;
            }

            if (matchesShortcut(event, SHORTCUTS.moveDown)) {
                event.preventDefault();
                if (focusedRowIndex === null) {
                    moveFocus(0, false);
                } else {
                    moveFocus(Math.min(focusedRowIndex + 1, paginatedRows.length - 1), false);
                }
                return;
            }
            if (matchesShortcut(event, SHORTCUTS.moveUp)) {
                event.preventDefault();
                if (focusedRowIndex === null) {
                    moveFocus(paginatedRows.length - 1, false);
                } else {
                    moveFocus(Math.max(focusedRowIndex - 1, 0), false);
                }
                return;
            }
            if (matchesShortcut(event, SHORTCUTS.moveFirst)) {
                event.preventDefault();
                moveFocus(0, false);
                return;
            }
            if (matchesShortcut(event, SHORTCUTS.moveLast)) {
                event.preventDefault();
                moveFocus(paginatedRows.length - 1, false);
                return;
            }
            if (matchesShortcut(event, SHORTCUTS.pageDown)) {
                event.preventDefault();
                if (focusedRowIndex === null) {
                    moveFocus(Math.min(visibleRowCount - 1, paginatedRows.length - 1), false);
                } else {
                    moveFocus(Math.min(focusedRowIndex + visibleRowCount, paginatedRows.length - 1), false);
                }
                return;
            }
            if (matchesShortcut(event, SHORTCUTS.pageUp)) {
                event.preventDefault();
                if (focusedRowIndex === null) {
                    moveFocus(0, false);
                } else {
                    moveFocus(Math.max(focusedRowIndex - visibleRowCount, 0), false);
                }
                return;
            }
            if (matchesShortcut(event, SHORTCUTS.toggleSelect)) {
                if (focusedRowIndex !== null) {
                    event.preventDefault();
                    handleSelectRow(focusedRowIndex);
                }
                return;
            }
            if (matchesShortcut(event, SHORTCUTS.editRow)) {
                if (focusedRowIndex !== null && canEditRows) {
                    event.preventDefault();
                    handleEdit(focusedRowIndex);
                }
                return;
            }
            if (matchesShortcut(event, SHORTCUTS.closeDialogs)) {
                event.preventDefault();
                setFocusedRowIndex(null);
                return;
            }
        };

        window.addEventListener('keydown', handleKeyDown);
        return () => window.removeEventListener('keydown', handleKeyDown);
    }, [enableKeyboardShortcuts, onRefresh, checked, paginatedRows, handleDeleteRow, handleEdit, focusedRowIndex, moveFocus, visibleRowCount, handleSelectRow, canEditRows, canDeleteRows, onPageChange, currentPage, totalPages, openExport]);

    useEffect(() => {
        if (tableRef.current) {
            setContainerWidth(tableRef.current.offsetWidth);
        }
    }, [tableRef]);

    useEffect(() => {
        if (!searchRef) return;

        let lastHighlightedCell: HTMLElement | null = null;

        searchRef.current = (search: string) => {

            document.querySelectorAll('.table-search-highlight').forEach(el => {
                el.classList.remove('bg-yellow-200', 'table-search-highlight', 'bg-muted');
            });

            if (lastHighlightedCell) {
                lastHighlightedCell.classList.remove('bg-muted', 'table-search-highlight', 'bg-yellow-200');
                lastHighlightedCell = null;
            }

            if (!search || !rows || !columns) {
                lastSearchState.current = { search: '', matchIdx: 0 };
                return;
            }

            const matches: { rowIdx: number; colIdx: number }[] = [];
            rows.forEach((row, rowIdx) => {
                row.forEach((cellValue, colIdx) => {
                    if (cellValue !== undefined && cellValue !== null) {
                        const searchValue = String(cellValue);
                        if (searchValue.toLowerCase().includes(search.toLowerCase())) {
                            matches.push({ rowIdx, colIdx });
                        }
                    }
                });
            });

            if (matches.length > 0) {

                let matchIdx = 0;
                if (lastSearchState.current.search === search) {

                    matchIdx = (lastSearchState.current.matchIdx + 1) % matches.length;
                }

                lastSearchState.current = { search, matchIdx };

                const { rowIdx, colIdx } = matches[matchIdx];

                const selector = `[data-row-idx="${rowIdx}"] [data-col-idx="${colIdx}"]`;
                const cell = document.querySelector(selector) as HTMLElement | null;
                if (cell) {

                    document.querySelectorAll('.table-search-highlight').forEach(el => {
                        el.classList.remove('bg-muted', 'table-search-highlight', 'bg-yellow-200');
                    });
                    cell.classList.add('bg-muted', 'table-search-highlight');
                    lastHighlightedCell = cell;
                    cell.scrollIntoView({ behavior: 'smooth', block: 'center', inline: 'center' });
                    setTimeout(() => {
                        if (cell === lastHighlightedCell) {
                            cell.classList.remove('bg-muted', 'table-search-highlight');
                            lastHighlightedCell = null;
                        }
                    }, 3000);
                }
            } else {

                lastSearchState.current = { search, matchIdx: 0 };
            }
        };

        return () => {
            if (lastHighlightedCell) {
                lastHighlightedCell.classList.remove('bg-muted', 'table-search-highlight', 'bg-yellow-200');
                lastHighlightedCell = null;
            }
        };
    }, [searchRef, rows, columns]);

    const actualTableHeight = useMemo(() => {
        if (enforceMinHeight) {

            return height;
        }

        if (paginatedRows.length === 0) return Math.min(500, height);
        const contentHeight = paginatedRows.length * rowHeight;
        return Math.min(contentHeight + 1, height);
    }, [paginatedRows.length, rowHeight, height, enforceMinHeight]);

    const contextMenu = useCallback((index: number, style: React.CSSProperties) => {
        const isFocused = focusedRowIndex === index;
        const isSelected = checked.includes(index);

        const tableRow = (
            <TableRow
                data-row-idx={index}
                role="row"
                aria-rowindex={index + 1}
                aria-selected={isSelected}
                data-focused={isFocused || undefined}
                tabIndex={isFocused ? 0 : -1}
                className={cn(
                    "group relative cursor-pointer",

                    isFocused && "bg-primary/5",

                    isSelected && "bg-muted"
                )}
                style={style}
                onClick={() => setFocusedRowIndex(index)}
                onFocus={() => setFocusedRowIndex(index)}
            >
                <TableCell
                    role="gridcell"
                    className={cn("min-w-[40px] w-[40px]", {
                        "hidden": disableEdit,
                    })}
                >
                    <Checkbox
                        checked={isSelected}
                        onCheckedChange={() => setChecked(isSelected ? checked.filter(i => i !== index) : [...checked, index])}
                        aria-label={isSelected ? t('deselectRow') : t('selectRow')}
                    />
                    <Button variant="secondary" className="opacity-0 group-hover:opacity-100 absolute right-2 w-0 top-1.5" onClick={(e) => {
                        e.preventDefault();
                        e.stopPropagation();

                        const event = new MouseEvent("contextmenu", {
                            bubbles: true,
                            clientX: e.clientX,
                            clientY: e.clientY,
                        });
                        e.currentTarget.dispatchEvent(event);
                        }} data-testid="icon-button" aria-label={t('moreActions')}>
                        <EllipsisVerticalIcon className="w-4 h-4" />
                    </Button>
                </TableCell>
                {paginatedRows[index]?.map((cell, cellIdx) => (
                    <TableCell
                        key={cellIdx}
                        role="gridcell"
                        className="cursor-pointer"
                        title={t('cellInteractionHint')}
                        onClick={(e) => {
                            e.stopPropagation();
                            setFocusedRowIndex(index);
                            handleCellClick(index, cellIdx);
                        }}
                        onDoubleClick={() => handleCellDoubleClick(index)}
                        onContextMenu={() => !limitContextMenu && setContextMenuCellIdx(cellIdx)}
                        data-col-idx={cellIdx}
                    >
                        {cell}
                    </TableCell>
                ))}
            </TableRow>
        );

        return <ContextMenu key={index}>
            <ContextMenuTrigger className="contents">
                {tableRow}
            </ContextMenuTrigger>
            <ContextMenuContent
                className="w-52 max-h-[calc(100vh-2rem)] overflow-y-auto"
                collisionPadding={{ top: 20, right: 20, bottom: 20, left: 20 }}
            >
                <ContextMenuItem
                    onSelect={() => {
                        if (contextMenuCellIdx == null) return;
                        const cell = paginatedRows[index]?.[contextMenuCellIdx];
                        if (cell !== undefined && cell !== null) {
                            copyToClipboard(String(cell)).then(success => {
                                if (success) toast.success(t('copiedCellToClipboard'));
                            });
                        }
                    }}
                    disabled={contextMenuCellIdx == null}
                >
                    <DocumentDuplicateIcon className="w-4 h-4" />
                    {t('copyCell')}
                    <ContextMenuShortcut><CursorArrowRaysIcon className="w-4 h-4" /></ContextMenuShortcut>
                </ContextMenuItem>
                {onEntitySearch && contextMenuCellIdx !== null && columnIsForeignKey?.[contextMenuCellIdx] && !columnIsPrimary?.[contextMenuCellIdx] && (
                    <ContextMenuItem
                        onSelect={() => {
                            if (contextMenuCellIdx == null) return;
                            const cell = paginatedRows[index]?.[contextMenuCellIdx];
                            const columnName = columns[contextMenuCellIdx];
                            if (cell !== undefined && cell !== null && columnName) {
                                onEntitySearch(columnName, String(cell));
                            }
                        }}
                    >
                        <MagnifyingGlassIcon className="w-4 h-4" />
                        {t('searchForEntity')}
                    </ContextMenuItem>
                )}
                <ContextMenuItem
                    onSelect={() => {
                        const row = paginatedRows[index];
                        if (row && Array.isArray(row)) {
                            const rowString = row.map(cell => cell ?? "").join("\t");
                            copyToClipboard(rowString).then(success => {
                                if (success) toast.success(t('rowCopiedToClipboard'));
                            });
                        }
                    }}
                    className="[&>[data-slot='context-menu-shortcut']]:flex"
                >
                    <DocumentTextIcon className="w-4 h-4" />
                    {t('copyRow')}
                    <ContextMenuShortcut><CursorArrowRaysIcon className="w-4 h-4" /><CursorArrowRaysIcon className="w-4 h-4" /></ContextMenuShortcut>
                </ContextMenuItem>
                {!limitContextMenu && (
                    <ContextMenuItem onSelect={() => handleSelectRow(index)}>
                        <CheckCircleIcon className="w-4 h-4 text-primary" />
                        {checked.includes(index) ? t('deselectRow') : t('selectRow')}
                        <ContextMenuShortcut>Space</ContextMenuShortcut>
                    </ContextMenuItem>
                )}
                {!limitContextMenu && canEditRows && (
                    <ContextMenuItem onSelect={() => handleEdit(index)} disabled={checked.length > 1} data-testid="context-menu-edit-row">
                        <PencilSquareIcon className="w-4 h-4" />
                        {t('editRow')}
                        <ContextMenuShortcut>Enter</ContextMenuShortcut>
                    </ContextMenuItem>
                )}
                <ContextMenuSub>
                    <ContextMenuSubTrigger>
                        <ArrowDownTrayIcon className="w-4 h-4 mr-2" />
                        {t('export')}
                    </ContextMenuSubTrigger>
                    <ContextMenuSubContent
                        collisionPadding={{ top: 20, right: 20, bottom: 20, left: 20 }}
                    >
                        <ContextMenuItem
                            onSelect={() => openExport('csv', true)}
                        >
                            <DocumentIcon className="w-4 h-4" />
                            {t('exportAllAsCsv')}
                            <ContextMenuShortcut>{formatShortcut(SHORTCUTS.exportData.displayKeys)}</ContextMenuShortcut>
                        </ContextMenuItem>
                        <ContextMenuItem
                            onSelect={() => openExport('excel', true)}
                        >
                            <DocumentIcon className="w-4 h-4" />
                            {t('exportAllAsExcel')}
                        </ContextMenuItem>
                        {!disableEdit && (
                            <>
                                <ContextMenuSeparator />
                                <ContextMenuItem
                                    onSelect={() => openExport('csv')}
                                    disabled={checked.length === 0}
                                >
                                    <DocumentIcon className="w-4 h-4" />
                                    {t('exportSelectedAsCsv')}
                                </ContextMenuItem>
                                <ContextMenuItem
                                    onSelect={() => openExport('excel')}
                                    disabled={checked.length === 0}
                                >
                                    <DocumentIcon className="w-4 h-4" />
                                    {t('exportSelectedAsExcel')}
                                </ContextMenuItem>
                            </>
                        )}
                    </ContextMenuSubContent>
                </ContextMenuSub>
                {!limitContextMenu && isMockDataSupported && (
                    <ContextMenuItem
                        onSelect={() => setShowMockDataSheet(true)}
                    >
                        <DocumentDuplicateIcon className="w-4 h-4" />
                        {t('mockData')}
                        <ContextMenuShortcut>{formatShortcut(SHORTCUTS.mockData.displayKeys)}</ContextMenuShortcut>
                    </ContextMenuItem>
                )}
                {!limitContextMenu && canDeleteRows && (
                    <ContextMenuItem
                        variant="destructive"
                        disabled={deleting}
                        onSelect={async () => {
                            await handleDeleteRow(index);
                        }}
                        data-testid="context-menu-delete-row"
                    >
                        <TrashIcon className="w-4 h-4 text-destructive" />
                        {t('deleteRow')}
                        <ContextMenuShortcut>{formatShortcut(SHORTCUTS.deleteRow.displayKeys)}</ContextMenuShortcut>
                    </ContextMenuItem>
                )}
            </ContextMenuContent>
        </ContextMenu>
    }, [checked, handleCellClick, handleEdit, handleSelectRow, handleDeleteRow, paginatedRows, disableEdit, limitContextMenu, onRefresh, t, contextMenuCellIdx, columns, columnIsForeignKey, columnIsPrimary, onEntitySearch, deleting, focusedRowIndex, isMockDataSupported, openExport, canEditRows, canDeleteRows]);

    return (
        <div ref={tableRef} className="flex min-w-0 w-full">
            <div className="flex flex-col space-y-4 min-w-0 w-full" data-testid="table-container">
                <div className="overflow-x-auto" style={{
                    width: `${containerWidth}px`,
                }}>
                    <TableComponent
                        role="grid"
                        aria-label={storageUnit ? `${storageUnit} data table` : 'Data table'}
                        aria-rowcount={paginatedRows.length}
                        aria-multiselectable={true}
                    >
                    <TableHeader data-column-count={columns.length}>
                        <ContextMenu>
                            <ContextMenuTrigger asChild>
                                    <TableHeadRow role="row" aria-rowindex={0} className="group relative cursor-context-menu hover:bg-muted/50 transition-colors" title={t('rightClickForOptions')}>
                                        <TableHead className={cn("min-w-[40px] w-[40px] relative", {
                                            "hidden": disableEdit,
                                        })}>
                                            <Checkbox
                                                checked={checked.length === paginatedRows.length && paginatedRows.length > 0}
                                                onCheckedChange={() => {
                                                    setChecked(checked.length === paginatedRows.length ? [] : paginatedRows.map((_, index) => index));
                                                }}
                                                aria-label={checked.length === paginatedRows.length ? t('deselectAll') : t('selectAll')}
                                            />
                                            <Button variant="secondary" className="opacity-0 group-hover:opacity-100 absolute right-2 top-1.5 w-0" onClick={(e) => {
                                                e.preventDefault();
                                                e.stopPropagation();

                                                const event = new MouseEvent("contextmenu", {
                                                    bubbles: true,
                                                    clientX: e.clientX,
                                                    clientY: e.clientY,
                                                });
                                                e.currentTarget.dispatchEvent(event);
                                                }} data-testid="icon-button" aria-label={t('moreActions')}>
                                                <EllipsisVerticalIcon className="w-4 h-4" aria-hidden="true" />
                                            </Button>
                                        </TableHead>
                                        {columns.map((col, idx) => (
                                            <TableHead
                                                key={col + idx}
                                                icon={columnIsPrimary?.[idx] ? <KeyIcon className="w-4 h-4" aria-label="Primary key" /> : columnIsForeignKey?.[idx] ? <ShareIcon className="w-4 h-4" aria-label="Foreign key" /> : columnIcons?.[idx]}
                                                className={cn({
                                                    "cursor-pointer select-none": onColumnSort,
                                                })}
                                                tabIndex={onColumnSort ? 0 : undefined}
                                                onClick={() => onColumnSort?.(col)}
                                                onKeyDown={(e) => {
                                                    if (onColumnSort && (e.key === 'Enter' || e.key === ' ')) {
                                                        e.preventDefault();
                                                        onColumnSort(col);
                                                    }
                                                }}
                                                onFocus={() => { focusedColumnRef.current = col; }}
                                                data-testid={`column-header-${col}`}
                                                data-column-name={col}
                                                data-sort-direction={sortedColumns?.get(col) || undefined}
                                            >
                                                <Tip>
                                                    <p className={cn("flex items-center gap-xs", {
                                                        "font-bold": columnIsPrimary?.[idx],
                                                        "italic": columnIsForeignKey?.[idx] && !columnIsPrimary?.[idx],
                                                    })}>
                                                        {col}
                                                        {onColumnSort && sortedColumns?.has(col) && (
                                                            sortedColumns.get(col) === 'asc'
                                                                ? <ChevronUpIcon className="w-4 h-4" data-testid="sort-indicator" />
                                                                : <ChevronDownIcon className="w-4 h-4" data-testid="sort-indicator" />
                                                        )}
                                                    </p>
                                                    <p className="text-xs">{columnTypes?.[idx]?.toLowerCase()}</p>
                                                </Tip>
                                            </TableHead>
                                        ))}
                                    </TableHeadRow>
                                </ContextMenuTrigger>
                                <ContextMenuContent
                    className="w-64 max-h-[calc(100vh-2rem)] overflow-y-auto"
                    collisionPadding={{ top: 16, right: 16, bottom: 16, left: 16 }}
                >
                                    {!limitContextMenu && isMockDataSupported && (
                                        <ContextMenuItem onSelect={() => setShowMockDataSheet(true)} data-testid="context-menu-mock-data">
                                            <CalculatorIcon className="w-4 h-4" />
                                            {t('mockData')}
                                            <ContextMenuShortcut>{formatShortcut(SHORTCUTS.mockData.displayKeys)}</ContextMenuShortcut>
                                        </ContextMenuItem>
                                    )}
                                    {!limitContextMenu && isMockDataSupported && <ContextMenuSeparator />}
                                    <ContextMenuSub>
                                        <ContextMenuSubTrigger>
                                            <ArrowDownCircleIcon className="w-4 h-4 mr-2" />
                                            {t('exportData')}
                                        </ContextMenuSubTrigger>
                                        <ContextMenuSubContent
                                            collisionPadding={{ top: 20, right: 20, bottom: 20, left: 20 }}
                                        >
                                            <ContextMenuItem
                                                onSelect={() => openExport('csv', true)}
                                            >
                                                <DocumentIcon className="w-4 h-4" />
                                                {t('exportAllAsCsv')}
                                                <ContextMenuShortcut>{formatShortcut(SHORTCUTS.exportData.displayKeys)}</ContextMenuShortcut>
                                            </ContextMenuItem>
                                            <ContextMenuItem
                                                onSelect={() => openExport('excel', true)}
                                            >
                                                <DocumentIcon className="w-4 h-4" />
                                                {t('exportAllAsExcel')}
                                            </ContextMenuItem>
                                            {!disableEdit && (
                                                <>
                                                    <ContextMenuSeparator />
                                                    <ContextMenuItem
                                                        onSelect={() => openExport('csv')}
                                                        disabled={checked.length === 0}
                                                    >
                                                        <DocumentIcon className="w-4 h-4" />
                                                        {t('exportSelectedAsCsv')}
                                                    </ContextMenuItem>
                                                    <ContextMenuItem
                                                        onSelect={() => openExport('excel')}
                                                        disabled={checked.length === 0}
                                                    >
                                                        <DocumentIcon className="w-4 h-4" />
                                                        {t('exportSelectedAsExcel')}
                                                    </ContextMenuItem>
                                                </>
                                            )}
                                        </ContextMenuSubContent>
                                    </ContextMenuSub>
                                    <ContextMenuSeparator />
                                    {!limitContextMenu && (
                                        <ContextMenuItem onSelect={() => onRefresh?.()}>
                                            <CircleStackIcon className="w-4 h-4" />
                                            {t('refreshData')}
                                            <ContextMenuShortcut>{formatShortcut(SHORTCUTS.refresh.displayKeys)}</ContextMenuShortcut>
                                        </ContextMenuItem>
                                    )}
                                    {!limitContextMenu && (
                                        <ContextMenuItem
                                            onSelect={() => {
                                                setChecked(checked.length === paginatedRows.length ? [] : paginatedRows.map((_, index) => index));
                                            }}
                                        >
                                            <CheckCircleIcon className="w-4 h-4" />
                                            {checked.length === paginatedRows.length ? t('deselectAll') : t('selectAll')}
                                            <ContextMenuShortcut>{formatShortcut(SHORTCUTS.selectAll.displayKeys)}</ContextMenuShortcut>
                                        </ContextMenuItem>
                                    )}
                                </ContextMenuContent>
                        </ContextMenu>
                    </TableHeader>
                    {paginatedRows.length > 0 && (
                        <VirtualizedTableBody
                            rowCount={paginatedRows.length}
                            rowHeight={rowHeight}
                            height={actualTableHeight}
                            overscan={10}
                            style={{
                                overflowY: 'scroll',
                            }}
                        >
                            {(rowIdx: number, rowStyle: React.CSSProperties) => contextMenu(rowIdx, rowStyle)}
                        </VirtualizedTableBody>
                    )}
                </TableComponent>
                {paginatedRows.length === 0 && (
                    <ContextMenu>
                        <ContextMenuTrigger asChild>
                            <div className="flex items-center justify-center cursor-pointer border rounded-lg h-[200px]">
                                <EmptyState className="table-empty-state" title={t('noDataAvailable')} description={t('noDataAvailable')} icon={<DocumentTextIcon className="w-4 h-4" />} />
                            </div>
                        </ContextMenuTrigger>
                        <ContextMenuContent
                className="w-52 max-h-[calc(100vh-2rem)] overflow-y-auto"
                collisionPadding={{ top: 20, right: 20, bottom: 20, left: 20 }}
            >
                            <ContextMenuItem onSelect={() => setShowMockDataSheet(true)} className={cn({
                                "hidden": disableEdit || !isMockDataSupported,
                            })}>
                                <CalculatorIcon className="w-4 h-4" />
                                {t('mockData')}
                                <ContextMenuShortcut>{formatShortcut(SHORTCUTS.mockData.displayKeys)}</ContextMenuShortcut>
                            </ContextMenuItem>
                            <ContextMenuSub>
                                <ContextMenuSubTrigger>
                                    <ArrowDownCircleIcon className="w-4 h-4 mr-2" />
                                    {t('export')}
                                </ContextMenuSubTrigger>
                                <ContextMenuSubContent
                                    collisionPadding={{ top: 20, right: 20, bottom: 20, left: 20 }}
                                >
                                    <ContextMenuItem
                                        onSelect={() => openExport('csv', true)}
                                    >
                                        <DocumentIcon className="w-4 h-4" />
                                        {t('exportAllAsCsv')}
                                        <ContextMenuShortcut>{formatShortcut(SHORTCUTS.exportData.displayKeys)}</ContextMenuShortcut>
                                    </ContextMenuItem>
                                    <ContextMenuItem
                                        onSelect={() => openExport('excel', true)}
                                    >
                                        <DocumentIcon className="w-4 h-4" />
                                        {t('exportAllAsExcel')}
                                    </ContextMenuItem>
                                </ContextMenuSubContent>
                            </ContextMenuSub>
                        </ContextMenuContent>
                    </ContextMenu>
                )}
                </div>
                <div className={cn("flex justify-between items-center", {
                    "justify-end": children == null,
                    "mt-4": children != null,
                })}>
                    {children}
                    <Pagination
                        className={cn("flex justify-end", {
                            "hidden": !showPagination,
                        })}
                        aria-label={t('tablePagination')}
                    >
                        <PaginationContent>
                            <PaginationItem>
                                <PaginationPrevious
                                    href="#"
                                    onClick={e => {
                                        e.preventDefault();
                                        if (currentPage > 1) handlePageChange(currentPage - 1);
                                    }}
                                    aria-disabled={currentPage === 1}
                                    aria-label={t('previousPage')}
                                    size="sm"
                                    className={cn({
                                        "opacity-50 pointer-events-none": currentPage === 1,
                                    })}
                                />
                            </PaginationItem>
                            {renderPaginationLinks()}
                            <PaginationItem>
                                <PaginationNext
                                    href="#"
                                    onClick={e => {
                                        e.preventDefault();
                                        if (currentPage < totalPages) handlePageChange(currentPage + 1);
                                    }}
                                    aria-disabled={currentPage === totalPages}
                                    aria-label={t('nextPage')}
                                    size="sm"
                                    className={cn({
                                        "opacity-50 pointer-events-none": currentPage === totalPages,
                                    })}
                                />
                            </PaginationItem>
                        </PaginationContent>
                    </Pagination>
                </div>
                <div className="flex justify-end items-center mb-2 gap-4">
                    {totalCount != null && totalCount > 0 && (
                        <div className="text-sm" data-testid="total-count-bottom">
                            <span className="font-semibold">{t('totalCount')}</span> {formatNumber(totalCount, language)}
                        </div>
                    )}
                    {isImportSupported && allowImport && (
                        <Button
                            variant="secondary"
                            onClick={() => setShowImport(true)}
                            className="flex gap-sm"
                            data-testid="import-button"
                        >
                            <ArrowUpCircleIcon className="w-4 h-4" />
                            {t('importAction')}
                        </Button>
                    )}
                    <Button
                        variant="secondary"
                        onClick={() => openExport()}
                        className="flex gap-sm"
                        data-testid="export-all-button"
                    >
                        <ArrowDownCircleIcon className="w-4 h-4" />
                        {hasSelectedRows ? t('exportSelected', { count: checked.length }) : t('exportAll')}
                    </Button>
                </div>
                <Sheet open={editIndex !== null} onOpenChange={open => {
                    if (!open) {
                        setEditIndex(null);
                        setEditRow(null);
                        setEditRowInitialLengths([]);
                    }
                }}>
                    <SheetContent side="right" className="w-[400px] max-w-full p-8 flex flex-col" data-testid="edit-row-dialog">
                        <SheetTitle>{t('editRow')}</SheetTitle>
                        <div className="flex-1 overflow-y-auto mt-4">
                            <div className="flex flex-col gap-lg pr-2">
                                {editRow &&
                                    columns.map((col, idx) => (
                                        <div key={col} className="flex flex-col gap-2">
                                            <div className="flex flex-col gap-0.5">
                                                <Label>{col}</Label>
                                                {columnTypes?.[idx] && (
                                                    <span className="text-xs text-muted-foreground">
                                                        {t('typeHint', { type: columnTypes[idx] })}
                                                    </span>
                                                )}
                                            </div>
                                            {
                                                editRowInitialLengths[idx] < 50 ?
                                                    <Input
                                                        key={`input-${idx}`}
                                                        value={editRow[idx] ?? ""}
                                                        onChange={e => handleInputChange(e.target.value, idx)}
                                                        data-testid={`editable-field-${idx}`}
                                                        {...getInputPropsForColumnType(columnTypes?.[idx] || '')}
                                                    />
                                                    : <TextArea
                                                        key={`textarea-${idx}`}
                                                        value={editRow[idx] ?? ""}
                                                        onChange={e => handleInputChange(e.target.value, idx)}
                                                        rows={5}
                                                        className="min-h-[100px]"
                                                        data-testid={`editable-field-${idx}`}
                                                    />
                                            }
                                        </div>
                                    ))}
                            </div>
                        </div>
                        <SheetFooter className="flex gap-sm px-0 mt-4">
                            <Button
                                className="flex-1"
                                variant="secondary"
                                onClick={() => {
                                    setEditIndex(null);
                                    setEditRow(null);
                                    setEditRowInitialLengths([]);
                                }}
                                data-testid="cancel-edit-row"
                            >
                                {t('cancel')}
                            </Button>
                            <Button className="flex-1" onClick={handleUpdate} disabled={!editRow} data-testid="update-button">
                                {t('update')}
                            </Button>
                        </SheetFooter>
                    </SheetContent>
                </Sheet>
            </div>
            <Sheet open={showMockDataSheet} onOpenChange={(open) => {
                setShowMockDataSheet(open);
                    if (!open) {
                        setShowMockDataConfirmation(false);
                    }
                }}>
                <SheetContent side="right" className="p-8" data-testid="mock-data-sheet">
                    <div className="flex flex-col gap-lg h-full">
                        <SheetTitle className="flex items-center gap-2"><CalculatorIcon className="w-4 h-4" /> {t('mockData')}</SheetTitle>
                        {!showMockDataConfirmation ? (
                            <div className="space-y-4">
                                <Label>{t('numberOfRows', { max: maxRowCount })}</Label>
                                <Input
                                    value={mockDataRowCount}
                                    onChange={e => handleMockDataRowCountChange(e.target.value)}
                                    type="text"
                                    inputMode="numeric"
                                    pattern="[0-9]*"
                                    max={maxRowCount.toString()}
                                    placeholder={t('enterNumberOfRows', { max: maxRowCount })}
                                    data-testid="mock-data-rows-input"
                                />
                                <Label>{t('method')}</Label>
                                <Select value={mockDataMethod} onValueChange={setMockDataMethod}>
                                    <SelectTrigger className="w-full" data-testid="mock-data-method-select">
                                        <SelectValue />
                                    </SelectTrigger>
                                    <SelectContent>
                                        <SelectItem value="Normal" data-value="Normal">{t('methodNormal')}</SelectItem>
                                    </SelectContent>
                                </Select>
                                <Label>{t('dataHandling')}</Label>
                                <Select value={mockDataOverwriteExisting} onValueChange={setMockDataOverwriteExisting}>
                                    <SelectTrigger className="w-full" data-testid="mock-data-handling-select">
                                        <SelectValue />
                                    </SelectTrigger>
                                    <SelectContent>
                                        <SelectItem value="append" data-value="append">{t('appendToExisting')}</SelectItem>
                                        <SelectItem value="overwrite" data-value="overwrite">{t('overwriteExisting')}</SelectItem>
                                    </SelectContent>
                                </Select>
                                {supportsMockDataRelations && (
                                    <>
                                        <div>
                                            <Label>{t('fkVariety')}</Label>
                                            <p className="text-sm text-muted-foreground mb-2">{t('fkVarietyDescription')}</p>
                                        </div>
                                        <Select value={mockDataFkDensityRatio} onValueChange={setMockDataFkDensityRatio}>
                                            <SelectTrigger className="w-full" data-testid="mock-data-fk-variety-select">
                                                <SelectValue />
                                            </SelectTrigger>
                                            <SelectContent>
                                                <SelectItem value="5" data-value="5">{t('fkVarietyHigh')}</SelectItem>
                                                <SelectItem value="10" data-value="10">{t('fkVarietyMedium')}</SelectItem>
                                                <SelectItem value="20" data-value="20">{t('fkVarietyNormal')}</SelectItem>
                                                <SelectItem value="50" data-value="50">{t('fkVarietyLow')}</SelectItem>
                                            </SelectContent>
                                        </Select>
                                    </>
                                )}
                                { }
                                {adjustedDepAnalysis?.Error && (
                                    <Alert variant="destructive" className="mt-4">
                                        <AlertTitle>{t('dependencyError')}</AlertTitle>
                                        <AlertDescription>
                                            {adjustedDepAnalysis.Error}
                                        </AlertDescription>
                                    </Alert>
                                )}
                                {adjustedDepAnalysis && !adjustedDepAnalysis.Error && adjustedDepAnalysis.Tables && adjustedDepAnalysis.Tables.length > 1 && (
                                    <div className="mt-4 p-3 border rounded-md bg-muted/50">
                                        <p className="text-sm font-medium mb-2">{t('tablesToPopulate')}</p>
                                        <ul className="text-sm space-y-1">
                                            {adjustedDepAnalysis.Tables.map((tbl) => (
                                                <li key={tbl.Table} className="flex items-center gap-2">
                                                    <span className="font-mono">{tbl.Table}</span>
                                                    <span className="text-muted-foreground">
                                                        ({tbl.RowsToGenerate} {t('rows')})
                                                    </span>
                                                    {tbl.UsesExistingData && (
                                                        <span className="text-xs px-1.5 py-0.5 rounded bg-secondary text-secondary-foreground">
                                                            {t('usingExisting')}
                                                        </span>
                                                    )}
                                                </li>
                                            ))}
                                        </ul>
                                        <p className="text-sm text-muted-foreground mt-2">
                                            {t('totalRows', { count: adjustedDepAnalysis.TotalRows })}
                                        </p>
                                    </div>
                                )}
                                {analyzingDeps && (
                                    <div className="mt-4 flex justify-center">
                                        <Spinner />
                                    </div>
                                )}
                                {generatingMockData && (
                                    <div className="mt-8 flex justify-center">
                                        <Spinner />
                                    </div>
                                )}
                            </div>
                        ) : (
                            <div className="space-y-4">
                                <div className="flex items-center justify-center mb-4">
                                    <div className="w-16 h-16 bg-yellow-100 dark:bg-yellow-900 rounded-full flex items-center justify-center">
                                        <XMarkIcon className="w-8 h-8 text-yellow-600 dark:text-yellow-400" />
                                    </div>
                                </div>
                                <p className="text-center text-gray-700 dark:text-gray-300">
                                    {t('overwriteConfirmation', { storageUnit })}
                                </p>
                            </div>
                        )}
                    </div>
                    <SheetFooter className="flex gap-sm px-0">
                        <Alert variant={supportsMockDataRelations ? "info" : "default"} className="mb-4">
                            <AlertTitle>{t('mockDataNote')}</AlertTitle>
                            <AlertDescription>
                                {supportsMockDataRelations ? t('mockDataWarning') : t('mockDataWarningClickHouse')}
                            </AlertDescription>
                        </Alert>
                        <Button
                            className="flex-1"
                            variant="secondary"
                            onClick={() => setShowMockDataSheet(false)}
                            data-testid="cancel-mock-data"
                        >
                            {t('cancel')}
                        </Button>
                        {!showMockDataConfirmation ? (
                            <Button className="flex-1" onClick={handleMockDataGenerate} disabled={generatingMockData || !mockDataRowCount || parseInt(mockDataRowCount) < 1} data-testid="mock-data-generate-button">
                                {t('generate')}
                            </Button>
                        ) : (
                            <Button className="flex-1" onClick={handleMockDataGenerate} disabled={generatingMockData || !mockDataRowCount || parseInt(mockDataRowCount) < 1} variant="destructive" data-testid="mock-data-overwrite-button">
                                {t('yesOverwrite')}
                            </Button>
                        )}
                    </SheetFooter>
                </SheetContent>
            </Sheet>
            <Suspense fallback={<Spinner />}>
                <DynamicExport
                    open={showExportConfirm}
                    onOpenChange={setShowExportConfirm}
                    storageUnit={rawQuery ? 'query_export' : (storageUnit || '')}
                    objectRef={objectRef}
                    hasSelectedRows={hasSelectedRows}
                    selectedRowsData={selectedRowsData}
                    checkedRowsCount={checked.length}
                    databaseType={databaseType}
                    rawQuery={rawQuery}
                    preselectedFormat={preselectedFormat}
                    forceExportAll={forceExportAll}
                />
            </Suspense>
            {isImportSupported && allowImport && (
                <ImportData
                    open={showImport}
                    onOpenChange={setShowImport}
                    objectRef={objectRef}
                    onImportSuccess={onRefresh}
                />
            )}
            <AlertDialog open={pendingDeleteIndexes != null} onOpenChange={(open) => { if (!open) handleCancelDelete(); }}>
                <AlertDialogContent>
                    <AlertDialogHeader>
                        <AlertDialogTitle>{t('deleteRowConfirmTitle', { count: pendingDeleteIndexes?.length ?? 1 })}</AlertDialogTitle>
                        <AlertDialogDescription>
                            {t('deleteRowConfirmDescription', { count: pendingDeleteIndexes?.length ?? 1 })}
                        </AlertDialogDescription>
                    </AlertDialogHeader>
                    <AlertDialogFooter>
                        <AlertDialogCancel onClick={handleCancelDelete}>{t('cancel')}</AlertDialogCancel>
                        <AlertDialogAction asChild>
                            <Button variant="destructive" onClick={handleConfirmDelete}>
                                {t('deleteRow', { count: pendingDeleteIndexes?.length ?? 1 })}
                            </Button>
                        </AlertDialogAction>
                    </AlertDialogFooter>
                </AlertDialogContent>
            </AlertDialog>
        </div>
    );
};