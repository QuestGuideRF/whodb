import {useLazyQuery, useMutation} from "@apollo/client/react";
import {
    Alert,
    AlertDescription,
    AlertTitle,
    Button,
    Card,
    cn,
    Dialog,
    DialogContent,
    DialogDescription,
    DialogFooter,
    DialogHeader,
    DialogTitle,
    DropdownMenu,
    DropdownMenuContent,
    DropdownMenuItem,
    DropdownMenuTrigger,
    EmptyState,
    Input,
    Select,
    SelectContent,
    SelectItem,
    SelectTrigger,
    SelectValue,
    toast
} from "@clidey/ux";
import {ChatHistorySidebar} from "./chat-history-sidebar";
import {ExecuteConfirmedSqlDocument, GenerateChatTitleDocument, GetAiChatQuery, GetDatabaseQuerySuggestionsDocument} from '@graphql';
import {
    ArrowUpCircleIcon,
    CheckCircleIcon,
    CircleStackIcon,
    CodeBracketIcon,
    CommandLineIcon,
    DocumentDuplicateIcon,
    EllipsisHorizontalIcon,
    PresentationChartLineIcon,
    SparklesIcon,
    TableCellsIcon
} from "../../components/heroicons";
import classNames from "classnames";
import {cloneElement, FC, KeyboardEventHandler, useCallback, useEffect, useMemo, useRef, useState} from "react";
import ReactMarkdown from 'react-markdown';
import logoImage from "../../../public/images/logo.svg";
import {AIProvider, useAI} from "../../components/ai";
import {Tip} from "../../components/tip";
import {CodeEditor} from "../../components/editor";
import {ErrorState} from "../../components/error-state";
import {Loading} from "../../components/loading";
import {InternalPage} from "../../components/page";
import {StorageUnitTable} from "../../components/table";
import {copyToClipboard} from "../../services/clipboard";
import {extensions} from "../../config/features";
import {InternalRoutes} from "../../config/routes";
import {HoudiniActions} from "../../store/chat";
import {useAppDispatch, useAppSelector} from "../../store/hooks";
import {ScratchpadActions} from "../../store/scratchpad";
import {featureFlags} from "../../config/features";
import {chooseRandomItems} from "../../utils/functions";
import {getComponent} from "../../config/component-registry";
import {useSourceContract} from "../../hooks/useSourceContract";
import {useNavigate} from "react-router-dom";
import {useChatExamples} from "./examples";
import {useTranslation} from '@/hooks/use-translation';
import {addAuthHeader} from "../../utils/auth-headers";
import {withBasePath} from "../../utils/base-path";
import {matchesShortcut, SHORTCUTS} from "../../utils/shortcuts";
import {useContainerWidth} from "../../hooks/use-container-width";
import {buildSourceScopeRef} from "../../utils/source-refs";

const LineChart = getComponent('line-chart');
const PieChart = getComponent('pie-chart');

const THINKING_PHRASES_COUNT = 25;

const CodeBlock: FC<{ children: string }> = ({ children }) => {
    const [copied, setCopied] = useState(false);

    const handleCopy = useCallback(() => {
        copyToClipboard(children).then(success => {
            if (success) {
                setCopied(true);
                setTimeout(() => setCopied(false), 2000);
            }
        });
    }, [children]);

    return (
        <div className="relative group/code-block my-2">
            <code className="block bg-neutral-100 dark:bg-neutral-800 p-2 pr-16 rounded text-sm overflow-x-auto">
                {children}
            </code>
            <button
                onClick={handleCopy}
                className="absolute top-1.5 right-1.5 flex items-center gap-1 px-1.5 py-0.5 rounded text-xs text-neutral-500 dark:text-neutral-400 hover:text-neutral-800 dark:hover:text-neutral-100 bg-neutral-200 dark:bg-neutral-700 hover:bg-neutral-300 dark:hover:bg-neutral-600 transition-colors"
            >
                {copied
                    ? <><CheckCircleIcon className="w-3.5 h-3.5 text-green-500" /> <span className="text-green-500">Copied!</span></>
                    : <><DocumentDuplicateIcon className="w-3.5 h-3.5" /> <span>Copy</span></>
                }
            </button>
        </div>
    );
};

type TableData = GetAiChatQuery["AIChat"][0]["Result"];

const TablePreview: FC<{ type: string, data: TableData, text: string, containerWidth?: number }> = ({ type, data, text, containerWidth }) => {
    const { t } = useTranslation('pages/chat');
    const dispatch = useAppDispatch();
    const [showSQL, setShowSQL] = useState(false);
    const [showScratchpadDialog, setShowScratchpadDialog] = useState(false);
    const [selectedPage, setSelectedPage] = useState<string>("new");
    const [newPageName, setNewPageName] = useState<string>("");
    const [dropdownOpen, setDropdownOpen] = useState(false);
    const navigate = useNavigate();
    const current = useAppSelector(state => state.auth.current);
    const { pages, activePageId } = useAppSelector(state => state.scratchpad);
    const { supportsScratchpad } = useSourceContract(current?.Type);

    const handleCodeToggle = useCallback(() => {
        setShowSQL(status => !status);
    }, []);

    const pageOptions = useMemo(() => {
        return [
            ...pages.map(page => ({ value: page.id, label: page.name })),
            { value: "new", label: t('createNewPage') }
        ];
    }, [pages, activePageId, t]);

    const handleMoveToScratchpad = useCallback(() => {
        if (!supportsScratchpad) {
            toast.error(t('scratchpadNotSupported'));
            return;
        }

        if (pages.length === 0) {
            dispatch(ScratchpadActions.ensurePagesHaveCells());
        }
        setShowScratchpadDialog(true);
    }, [dispatch, pages.length, supportsScratchpad, t]);

    const handleScratchpadConfirm = useCallback(() => {
        if (selectedPage === "new") {

            const pageName = newPageName.trim() || `Page ${pages.length + 1}`;
            dispatch(ScratchpadActions.addPage({ name: pageName, initialQuery: text }));

            navigate(InternalRoutes.RawExecute.path, {
                state: {
                    targetPage: "new"
                }
            });
        } else {

            dispatch(ScratchpadActions.addCellToPageAndActivate({
                pageId: selectedPage,
                initialQuery: text
            }));

            navigate(InternalRoutes.RawExecute.path, {
                state: {
                    targetPage: selectedPage
                }
            });
        }
        setShowScratchpadDialog(false);
        setSelectedPage("new");
        setNewPageName("");
        toast.success(t('queryMoved'));
    }, [navigate, text, selectedPage, newPageName, pages.length, dispatch, t]);

    const previewResult = useMemo(() => {
        if (data == null || data.Rows.length === 0) {
            return t('noDataReturned');
        }
        return type.toUpperCase().split(":")?.[1];
    }, [data, type, t]);

    const canMoveToScratchpad = useMemo(() => {
        return supportsScratchpad && type.startsWith("sql:");
    }, [supportsScratchpad, type]);

    return <div className="flex gap-2 w-[calc(100%-50px)] max-w-full min-w-0 group/table-preview">
        <div className={cn("transition-all shrink-0 pt-1", {
            "opacity-0 group-hover/table-preview:opacity-100 focus-within:opacity-100": !dropdownOpen,
            "opacity-100": dropdownOpen,
        })}>
            <DropdownMenu open={dropdownOpen} onOpenChange={setDropdownOpen}>
                <DropdownMenuTrigger>
                    <Button variant="outline" size="sm" data-testid="icon-button" aria-label={t('actions')}>
                        <EllipsisHorizontalIcon className="w-5 h-5" aria-hidden="true" />
                    </Button>
                </DropdownMenuTrigger>
                <DropdownMenuContent align="start">
                    <DropdownMenuItem onClick={handleCodeToggle} data-testid="toggle-view-option">
                        {showSQL ? (
                            <>
                                <TableCellsIcon className="w-4 h-4 mr-2" />
                                {t('showTable')}
                            </>
                        ) : (
                            <>
                                <CodeBracketIcon className="w-4 h-4 mr-2" />
                                {t('showCode')}
                            </>
                        )}
                    </DropdownMenuItem>
                    {canMoveToScratchpad && (
                        <DropdownMenuItem onClick={handleMoveToScratchpad} data-testid="move-to-scratchpad-option">
                            <CommandLineIcon className="w-4 h-4 mr-2" />
                            {t('moveToScratchpad')}
                        </DropdownMenuItem>
                    )}
                </DropdownMenuContent>
            </DropdownMenu>
        </div>
        <div className="flex flex-col gap-lg overflow-hidden break-all leading-6 shrink-0 w-full max-w-full min-w-0">
            {
                showSQL
                ? <div className="h-[300px] w-full">
                    <CodeEditor value={text} language="sql" />
                </div>
                :  (data != null && data.Rows.length > 0) || type === "sql:get"
                    ? <div className="w-full">
                        <StorageUnitTable
                            key={containerWidth}
                            columns={data?.Columns?.map(c => c.Name) ?? []}
                            columnTypes={data?.Columns?.map(c => c.Type) ?? []}
                            rows={data?.Rows ?? []}
                            disableEdit={true}
                            limitContextMenu={true}
                            databaseType={current?.Type}
                            rawQuery={text}
                            height={200}
                            enforceMinHeight={true}
                            totalCount={data?.Rows?.length ?? 0}
                        />
                    </div>
                    : (type.startsWith("sql:") && (type === "sql:insert" || type === "sql:update" || type === "sql:delete" || type === "sql:create" || type === "sql:alter" || type === "sql:drop"))
                    ? <Alert title={t('actionExecuted')} className="w-fit">
                        <CheckCircleIcon className="w-4 h-4" />
                        <AlertTitle>{t('actionExecuted')}</AlertTitle>
                        <AlertDescription>
                            {previewResult}
                        </AlertDescription>
                    </Alert>
                    : null
            }
        </div>

        <Dialog open={showScratchpadDialog} onOpenChange={setShowScratchpadDialog}>
            <DialogContent>
                <DialogHeader>
                    <DialogTitle>{t('moveToScratchpad')}</DialogTitle>
                    <DialogDescription>
                        {t('dialogDescription')}
                    </DialogDescription>
                </DialogHeader>
                <div className="py-4 space-y-4">
                    <div>
                        <label className="text-sm font-medium mb-2 block">{t('selectPageLabel')}</label>
                        <Select value={selectedPage} onValueChange={setSelectedPage}>
                            <SelectTrigger className="w-full">
                                <SelectValue placeholder={t('selectPagePlaceholder')} />
                            </SelectTrigger>
                            <SelectContent>
                                {pageOptions.map((option) => (
                                    <SelectItem key={option.value} value={option.value}>
                                        {option.label}
                                    </SelectItem>
                                ))}
                            </SelectContent>
                        </Select>
                    </div>
                    {selectedPage === "new" && (
                        <div>
                            <label className="text-sm font-medium mb-2 block">{t('newPageLabel')}</label>
                            <Input
                                value={newPageName}
                                onChange={(e) => setNewPageName(e.target.value)}
                                placeholder={t('newPagePlaceholder')}
                            />
                        </div>
                    )}
                </div>
                <DialogFooter>
                    <Button variant="outline" onClick={() => {
                        setShowScratchpadDialog(false);
                        setSelectedPage("new");
                        setNewPageName("");
                    }}>
                        {t('cancel')}
                    </Button>
                    <Button onClick={handleScratchpadConfirm}>
                        {t('moveToScratchpad')}
                    </Button>
                </DialogFooter>
            </DialogContent>
        </Dialog>
    </div>
}

export const ChatPage: FC = () => {
    const { t } = useTranslation('pages/chat');
    const [query, setQuery] = useState("");
    const { sessions, activeSessionId } = useAppSelector(state => state.houdini);
    const chats = useMemo(() => {
        if (sessions.length > 0 && activeSessionId) {
            const activeSession = sessions.find(s => s.id === activeSessionId);
            return activeSession?.messages || [];
        }
        return [];
    }, [sessions, activeSessionId]);
    const [executeConfirmedSql] = useMutation(ExecuteConfirmedSqlDocument);
    const [generateChatTitleMutation] = useMutation(GenerateChatTitleDocument);
    const scrollContainerRef = useRef<HTMLDivElement>(null);
    const containerWidth = useContainerWidth(scrollContainerRef);
    const schemaFromState = useAppSelector(state => state.database.schema);
    const authProfile = useAppSelector(state => state.auth.current);
    const { item } = useSourceContract(authProfile?.Type);
    const [executingConfirmedId, setExecutingConfirmedId] = useState<number | null>(null);
    const [showQueryForId, setShowQueryForId] = useState<number | null>(null);
    const [copiedSqlId, setCopiedSqlId] = useState<number | null>(null);
    const messageIdCounter = useRef(0);
    const sourceScopeRef = useMemo(() => buildSourceScopeRef(item, authProfile, schemaFromState), [authProfile, item, schemaFromState]);
    const [currentSearchIndex, setCurrentSearchIndex] = useState<number>();

    const dispatch = useAppDispatch();

    const getUniqueMessageId = useCallback(() => {
        messageIdCounter.current += 1;
        return Date.now() * 1000 + messageIdCounter.current;
    }, []);

    const aiState = useAI();
    const { modelType, currentModel, modelAvailable, models } = aiState;

    const chatExamples = useChatExamples();

    const thinkingPhrases = useMemo(() => {
        return Array.from({ length: THINKING_PHRASES_COUNT }, (_, i) => t(`thinking${i}`));
    }, [t]);

    const [loading, setLoading] = useState(false);
    const loadingPhraseRef = useRef<string>("");

    const [getDatabaseSuggestions, { loading: suggestionsLoading }] = useLazyQuery(GetDatabaseQuerySuggestionsDocument, {
        fetchPolicy: 'network-only',
        errorPolicy: 'all',
    });
    const [databaseSuggestions, setDatabaseSuggestions] = useState<Array<{ description: string; category: string }>>([]);
    const [useDatabaseSuggestions, setUseDatabaseSuggestions] = useState(false);
    const hasFetchedSuggestionsRef = useRef(false);

    const exampleIndicesRef = useRef<number[] | null>(null);

    useEffect(() => {
        if (exampleIndicesRef.current === null && chatExamples.length > 0) {
            const indices: number[] = [];
            const available = [...Array(chatExamples.length).keys()];
            const count = Math.min(3, chatExamples.length);
            for (let i = 0; i < count; i++) {
                const randomIndex = Math.floor(Math.random() * available.length);
                indices.push(available[randomIndex]);
                available.splice(randomIndex, 1);
            }
            exampleIndicesRef.current = indices;
        }
    }, [chatExamples.length]);

    const getCategoryIcon = useCallback((category: string) => {
        switch (category) {
            case 'SELECT':
                return <CircleStackIcon className="w-4 h-4" />;
            case 'AGGREGATE':
                return <PresentationChartLineIcon className="w-4 h-4" />;
            default:
                return <SparklesIcon className="w-4 h-4" />;
        }
    }, []);

    const examples = useMemo(() => {
        if (useDatabaseSuggestions && databaseSuggestions.length > 0) {
            return databaseSuggestions.map(s => ({
                icon: getCategoryIcon(s.category),
                description: s.description,
            }));
        }

        if (exampleIndicesRef.current === null) {
            return chatExamples.slice(0, 3);
        }
        return exampleIndicesRef.current.map(i => chatExamples[i]);
    }, [useDatabaseSuggestions, databaseSuggestions, chatExamples, getCategoryIcon]);

    const handleSubmitQuery = useCallback(async () => {
        const sanitizedQuery = query.trim();
        if (modelType == null || sanitizedQuery.length === 0) {
            return;
        }

        const activeSession = sessions.find(s => s.id === activeSessionId);
        const hasDefaultName = activeSession?.name?.match(/^Chat \d+$/);
        const shouldTryTitle = hasDefaultName;

        setLoading(true);
        loadingPhraseRef.current = chooseRandomItems(thinkingPhrases)[0];
        dispatch(HoudiniActions.addChatMessage({ Type: "message", Text: sanitizedQuery, isUserInput: true, RequiresConfirmation: false }));
        setQuery("");

        const streamingMessageId = getUniqueMessageId();
        dispatch(HoudiniActions.addChatMessage({
            Type: "message",
            Text: "",
            isStreaming: true,
            id: streamingMessageId,
            RequiresConfirmation: false
        }));

        setTimeout(() => {
            if (scrollContainerRef.current != null) {
                scrollContainerRef.current.scroll({
                    top: scrollContainerRef.current.scrollHeight,
                    behavior: "smooth",
                });
            }
        }, 250);

        try {
            const response = await fetch(withBasePath('/api/ai-chat/stream'), {
                method: 'POST',
                credentials: 'include',
                headers: addAuthHeader({
                    'Content-Type': 'application/json',
                }),
                body: JSON.stringify({
                    ref: sourceScopeRef,
                    modelType: modelType.modelType,
                    providerId: modelType.id || '',
                    token: modelType.token || '',
                    model: currentModel ?? '',
                    input: {
                        Query: sanitizedQuery,
                        PreviousConversation: chats.map(chat =>
                            `${chat.isUserInput ? "<User>" : "<System>"}${chat.Text}${chat.isUserInput ? "</User>" : "</System>"}`
                        ).join("\n"),
                    },
                }),
            });

            if (!response.ok) {
                setLoading(false);
                return;
            }

            if (!response.body) {
                setLoading(false);
                return;
            }

            const reader = response.body.getReader();
            const decoder = new TextDecoder();
            let streamingText = '';
            let currentEventType = '';
            let addedSqlMessages = new Set<string>();
            let streamDone = false;

            while (true) {
                const { done, value } = await reader.read();
                if (done) {
                    break;
                }

                const chunk = decoder.decode(value, { stream: true });
                const lines = chunk.split('\n');

                for (const line of lines) {
                    if (line.startsWith('event: ')) {
                        currentEventType = line.slice(7).trim();
                    } else if (line.startsWith('data: ')) {
                        const data = line.slice(6);
                        if (!data.trim()) continue;

                        try {
                            const parsed = JSON.parse(data);

                            if (currentEventType === 'chunk') {
                                const text = parsed.text || '';
                                const chunkType = parsed.type || '';

                                if (chunkType !== 'sql' && chunkType !== 'error' && text && text.length > streamingText.length) {
                                    streamingText = text;
                                    dispatch(HoudiniActions.updateChatMessage({
                                        id: streamingMessageId,
                                        Text: streamingText,
                                    }));
                                }

                                if (scrollContainerRef.current != null) {
                                    scrollContainerRef.current.scroll({
                                        top: scrollContainerRef.current.scrollHeight,
                                        behavior: "smooth",
                                    });
                                }
                            } else if (currentEventType === 'message') {

                                if (parsed.Type?.startsWith("sql") || parsed.Type === "error") {

                                    const messageKey = `${parsed.Type}:${parsed.Text}`;

                                    if (!addedSqlMessages.has(messageKey)) {
                                        addedSqlMessages.add(messageKey);

                                        const messageId = getUniqueMessageId();
                                        dispatch(HoudiniActions.addChatMessage({
                                            Type: parsed.Type,
                                            Text: parsed.Text,
                                            Result: parsed.Result,
                                            RequiresConfirmation: parsed.RequiresConfirmation || false,
                                            id: messageId,
                                        }));

                                        setTimeout(() => {
                                            if (scrollContainerRef.current != null) {
                                                scrollContainerRef.current.scroll({
                                                    top: scrollContainerRef.current.scrollHeight,
                                                    behavior: "smooth",
                                                });
                                            }
                                        }, 100);
                                    }
                                }
                            } else if (currentEventType === 'done') {

                                if (streamingText === '' || streamingText.trim() === '') {

                                    dispatch(HoudiniActions.removeChatMessage(streamingMessageId));
                                } else {

                                    dispatch(HoudiniActions.completeStreamingMessage({
                                        id: streamingMessageId,
                                        message: { Type: "message", Text: streamingText },
                                    }));
                                }
                                setLoading(false);

                                if (shouldTryTitle) {
                                    generateChatTitle(sanitizedQuery);
                                }
                                streamDone = true;
                            } else if (currentEventType === 'error') {
                                dispatch(HoudiniActions.removeChatMessage(streamingMessageId));
                                const errorMessage = typeof parsed.error === 'string'
                                    ? parsed.error
                                    : parsed.error?.message || parsed.message || 'Unknown error';
                                toast.error(t('unableToQuery') + " " + errorMessage);
                                setLoading(false);
                                streamDone = true;
                            }
                        } catch (e) {
                            console.error('Failed to parse SSE data:', e);
                        }
                    }
                }

                if (streamDone) {
                    reader.cancel();
                    break;
                }
            }
        } catch (error) {
            dispatch(HoudiniActions.removeChatMessage(streamingMessageId));
            const errorMessage = error instanceof Error
                ? error.message
                : typeof error === 'string'
                ? error
                : 'Unknown error';
            toast.error(t('unableToQuery') + " " + errorMessage);
            setLoading(false);
        }

    }, [chats, currentModel, modelType, query, sourceScopeRef, dispatch, t, scrollContainerRef, getUniqueMessageId, activeSessionId, sessions]);

    const generateChatTitle = useCallback(async (userQuery: string) => {
        if (!modelType || !activeSessionId || !currentModel) {
            return;
        }

        try {
            const result = await generateChatTitleMutation({
                variables: {
                    input: {
                        Query: userQuery,
                        ModelType: modelType.modelType,
                        ProviderId: modelType.id || undefined,
                        Token: modelType.token || undefined,
                        Model: currentModel,
                        Endpoint: undefined,
                    }
                }
            });

            const title = result.data?.GenerateChatTitle?.Title;
            if (title && title.trim() !== '') {
                dispatch(HoudiniActions.updateSessionName({
                    sessionId: activeSessionId,
                    name: title,
                }));
            }
        } catch (error) {
            console.error('[Chat Title] Failed to generate chat title:', error);

        }
    }, [modelType, currentModel, activeSessionId, dispatch, generateChatTitleMutation]);

    const disableChat = useMemo(() => {
        return loading || models.length === 0 || (!modelAvailable && !currentModel) || query.trim().length === 0;
    }, [loading, modelAvailable, models.length, currentModel, query]);

    const handleKeyDown: KeyboardEventHandler<HTMLInputElement> = useCallback((e) => {
        if (matchesShortcut(e, SHORTCUTS.clearEditor)) {
            e.preventDefault();
            setQuery('');
            setCurrentSearchIndex(undefined);
            return;
        }
    }, []);

    const handleKeyUp: KeyboardEventHandler<HTMLInputElement> = useCallback((e) => {
        if (e.key === "Enter") {
            if (query.trim().length > 0 && !disableChat) {
                handleSubmitQuery();
            }
            return;
        }
        if (e.key === "ArrowUp") {
          const foundSearchIndex = currentSearchIndex != null ? currentSearchIndex - 1 : chats.length - 1;
          let searchIndex = foundSearchIndex;

          while (searchIndex >= 0) {
            if (chats[searchIndex].isUserInput) {
              setCurrentSearchIndex(searchIndex);
              setQuery(chats[searchIndex].Text);
              return;
            }
            searchIndex--;
          }

          if (searchIndex < 0) {
            setCurrentSearchIndex(undefined);
            setQuery('');
            return;
          }

          if (currentSearchIndex !== chats.length - 1) {
            searchIndex = chats.length - 1;
            while (searchIndex > foundSearchIndex) {
              if (chats[searchIndex].isUserInput) {
                setCurrentSearchIndex(searchIndex);
                setQuery(chats[searchIndex].Text);
                return;
              }
              searchIndex--;
            }
          }
        }
        if (e.key === "ArrowDown") {
          if (currentSearchIndex == null) return;

          let searchIndex = currentSearchIndex + 1;
          while (searchIndex < chats.length) {
            if (chats[searchIndex].isUserInput) {
              setCurrentSearchIndex(searchIndex);
              setQuery(chats[searchIndex].Text);
              return;
            }
            searchIndex++;
          }

          setCurrentSearchIndex(undefined);
          setQuery('');
        }
    }, [chats, currentSearchIndex, query, handleSubmitQuery, disableChat]);

    const handleSelectExample = useCallback((example: string) => {
        setQuery(example);
    }, []);

    const handleClear = useCallback(() => {
        dispatch(HoudiniActions.clear());
        setQuery("");
        setCurrentSearchIndex(undefined);

        hasFetchedSuggestionsRef.current = false;
        setDatabaseSuggestions([]);
        setUseDatabaseSuggestions(false);
    }, [dispatch]);

    const handleConfirmSQL = useCallback(async (messageId: number, sql: string, operationType: string) => {
        setExecutingConfirmedId(messageId);
        try {
            const { data, error } = await executeConfirmedSql({
                variables: {
                    query: sql,
                    operationType: operationType,
                },
            });

            if (error || !data) {
                toast.error(t('unableToQuery') + " " + (error?.message || t('failedToExecuteSQL')));
                setLoading(false);
                return;
            }

            const result = data.ExecuteConfirmedSQL;

            dispatch(HoudiniActions.completeStreamingMessage({
                id: messageId,
                message: {
                    Type: result.Type,
                    Text: result.Text,
                    Result: result.Result,
                    RequiresConfirmation: false,
                },
            }));

            setTimeout(() => {
                if (scrollContainerRef.current != null) {
                    scrollContainerRef.current.scroll({
                        top: scrollContainerRef.current.scrollHeight,
                        behavior: "smooth",
                    });
                }
            }, 100);

        } catch (error) {
            const errorMessage = error instanceof Error
                ? error.message
                : typeof error === 'string'
                ? error
                : 'Unknown error';
            toast.error(t('unableToQuery') + " " + errorMessage);
        } finally {
            setExecutingConfirmedId(null);
        }
    }, [executeConfirmedSql, dispatch, t, scrollContainerRef]);

    const handleCancelSQL = useCallback((messageId: number) => {
        dispatch(HoudiniActions.removeChatMessage(messageId));
        toast.info(t('queryCancelled') || 'Query cancelled');
    }, [dispatch, t]);

    const disableAll = useMemo(() => {
        return models.length === 0 || (!modelAvailable && !currentModel);
    }, [modelAvailable, models.length, currentModel]);

    const hasInitialized = useRef(false);
    useEffect(() => {
        if (!hasInitialized.current) {
            hasInitialized.current = true;
            dispatch(HoudiniActions.initializeChatSessions());
        }
    }, [dispatch]);

    useEffect(() => {
        if (scrollContainerRef.current != null && chats.length > 0) {
            scrollContainerRef.current.scrollTop = scrollContainerRef.current.scrollHeight;
        }
    }, [chats.length]);

    useEffect(() => {

        const shouldFetch =
            modelAvailable &&
            currentModel &&
            chats.length === 0 &&
            !hasFetchedSuggestionsRef.current;

        if (shouldFetch) {
            hasFetchedSuggestionsRef.current = true;
            getDatabaseSuggestions({
                variables: {
                    ref: sourceScopeRef,
                },
            }).then(({ data }) => {
                if (data?.DatabaseQuerySuggestions && data.DatabaseQuerySuggestions.length > 0) {
                    setDatabaseSuggestions(data.DatabaseQuerySuggestions.map(s => ({
                        description: s.description,
                        category: s.category,
                    })));
                    setUseDatabaseSuggestions(true);
                } else {

                    setUseDatabaseSuggestions(false);
                }
            }).catch((error) => {
                console.error('[Database Suggestions] Error fetching suggestions:', error);

                setUseDatabaseSuggestions(false);
            });
        }
    }, [chats.length, currentModel, getDatabaseSuggestions, modelAvailable, sourceScopeRef]);

    return (
        <InternalPage routes={[InternalRoutes.Chat]} className="h-full" sidebar={<ChatHistorySidebar />}>
            <div className="flex flex-col w-full h-full gap-2">
                <AIProvider
                    {...aiState}
                    onClear={handleClear}
                />
                <div className={classNames("flex grow w-full rounded-xl overflow-hidden", {
                    "hidden": disableAll,
                })}>
                    {
                        chats.length === 0
                        ? <div className="flex flex-col justify-center items-center w-full gap-8" data-testid="chat-empty-state-container">
                            { }
                            <EmptyState title={t('emptyStateTitle')} description="" icon={<SparklesIcon className="w-16 h-16" data-testid="empty-state-sparkles-icon" />} />
                            {suggestionsLoading ? (
                                <Loading loadingText={t('loadingSuggestions')} size="sm" />
                            ) : (
                                <div className="flex flex-col gap-2 items-center">
                                    {!useDatabaseSuggestions && examples.length > 0 && (
                                        <p className="text-xs text-muted-foreground">{t('genericExamplesLabel')}</p>
                                    )}
                                    {useDatabaseSuggestions && databaseSuggestions.length > 0 && (
                                        <p className="text-xs text-muted-foreground">{t('databaseSpecificSuggestionsLabel')}</p>
                                    )}
                                    <div className="flex flex-wrap justify-center items-center gap-4" data-testid="chat-examples-list">
                                        {
                                            examples.map((example, i) => (
                                                <Card key={`chat-${i}`} className="flex flex-col gap-sm w-[250px] h-[120px] p-4 text-sm cursor-pointer hover:opacity-80 transition-all"
                                                    onClick={() => handleSelectExample(example.description)}>
                                                    {example.icon}
                                                    {example.description}
                                                </Card>
                                            ))
                                        }
                                    </div>
                                </div>
                            )}
                        </div>
                        : <div className="h-full w-full py-8 max-h-[calc(75vh-25px)] overflow-y-auto" ref={scrollContainerRef}>
                            <div className="flex justify-center w-full h-full max-w-full">
                                <div className="flex w-full flex-col gap-2 max-w-full min-w-0">
                                    {
                                        chats.map((chat, i) => {
                                            if (chat.Type === "message" || chat.Type === "text") {
                                                return <div key={`chat-${i}`} className={classNames("flex gap-lg overflow-hidden break-words leading-6 shrink-0 relative", {
                                                    "self-end ml-3": chat.isUserInput,
                                                    "self-start": !chat.isUserInput,
                                                })} data-testid={chat.isUserInput ? "user-message" : "system-message"}>
                                                    {!chat.isUserInput && chats[i-1]?.isUserInput
                                                        ? extensions.MetaIcon ?? <img src={logoImage} alt="clidey logo" className="w-auto h-8" />
                                                        : <div className="pl-4" />}
                                                    {chat.isUserInput ? (
                                                        <p className={classNames("py-2 rounded-xl whitespace-pre-wrap bg-neutral-600/5 dark:bg-[#2C2F33] px-4", {
                                                            "animate-fade-in": chat.isStreaming,
                                                        })} data-input-message="user">
                                                            {chat.Text}
                                                            {chat.isStreaming && <span className="inline-block w-2 h-4 ml-1 bg-current animate-pulse" />}
                                                        </p>
                                                    ) : (
                                                        <div className={classNames("py-2 rounded-xl markdown-content", {
                                                            "animate-fade-in": chat.isStreaming,
                                                        })} data-input-message="system">
                                                            <ReactMarkdown
                                                                components={{
                                                                    p: ({node, ...props}) => <p className="mb-2 last:mb-0" {...props} />,
                                                                    strong: ({node, ...props}) => <strong className="font-semibold" {...props} />,
                                                                    ul: ({node, ...props}) => <ul className="list-disc list-inside mb-2 space-y-1" {...props} />,
                                                                    ol: ({node, ...props}) => <ol className="list-decimal list-inside mb-2 space-y-1" {...props} />,
                                                                    li: ({node, ...props}) => <li className="ml-2" {...props} />,
                                                                    h1: ({node, ...props}) => <h1 className="text-xl font-bold mb-2 mt-4 first:mt-0" {...props} />,
                                                                    h2: ({node, ...props}) => <h2 className="text-lg font-semibold mb-2 mt-3 first:mt-0" {...props} />,
                                                                    h3: ({node, ...props}) => <h3 className="text-md font-semibold mb-1 mt-2 first:mt-0" {...props} />,
                                                                    code: ({node, children, ...props}) => {
                                                                        const isInline = !String(props.className || '').includes('language-');
                                                                        return isInline
                                                                            ? <code className="bg-neutral-100 dark:bg-neutral-800 px-1 py-0.5 rounded text-sm" {...props}>{children}</code>
                                                                            : <CodeBlock>{String(children)}</CodeBlock>;
                                                                    },
                                                                    blockquote: ({node, ...props}) => <blockquote className="border-l-4 border-neutral-300 dark:border-neutral-700 pl-4 my-2 italic" {...props} />,
                                                                }}
                                                            >
                                                                {chat.Text}
                                                            </ReactMarkdown>
                                                            {chat.isStreaming && <span className="inline-block w-2 h-4 ml-1 bg-current animate-pulse" />}
                                                        </div>
                                                    )}
                                                </div>
                                            } else if (chat.Type === "error") {
                                                return (
                                                    <div key={`chat-${i}`} className="flex overflow-hidden break-words leading-6 shrink-0 pt-6 relative self-start" data-testid="error-message">
                                                        {!chat.isUserInput && chats[i-1]?.isUserInput
                                                            ? extensions.MetaIcon ?? <img src={logoImage} alt="clidey logo" className="w-auto h-8" />
                                                            : null}
                                                        <ErrorState error={chat.Text.replace(/^ERROR:\s*/i, "")} />
                                                    </div>
                                                );
                                            } else if (featureFlags.dataVisualization && (chat.Type === "sql:pie-chart" || chat.Type === "sql:line-chart")) {
                                                return <div key={`chat-${i}`} className="flex gap-lg w-full max-w-full min-w-0 pt-4 relative" data-testid="visual-message">
                                                    {!chat.isUserInput && chats[i-1]?.isUserInput && (extensions.MetaIcon ?? <img src={logoImage} alt="clidey logo" className="w-auto h-8" />)}
                                                    { }
                                                    {chat.Type === "sql:pie-chart" && PieChart && <PieChart columns={chat.Result?.Columns?.map(col => col.Name) ?? []} data={chat.Result?.Rows ?? []} text={chat.Text} />}
                                                    { }
                                                    {chat.Type === "sql:line-chart" && LineChart && <LineChart columns={chat.Result?.Columns?.map(col => col.Name) ?? []} data={chat.Result?.Rows ?? []} text={chat.Text} />}
                                                </div>
                                            } else if (chat.RequiresConfirmation) {
                                                // Show confirmation UI inline
                                                const isExecuting = executingConfirmedId === chat.id;
                                                const showQuery = showQueryForId === chat.id;

                                                return <div key={`chat-${i}`} className="flex gap-lg w-full max-w-full min-w-0 pt-4 relative" data-testid="confirmation-message">
                                                    {!chat.isUserInput && chats[i-1]?.isUserInput
                                                        ? (extensions.MetaIcon ?? <img src={logoImage} alt="clidey logo" className="w-auto h-8" />)
                                                        : <div className="pl-4" />}
                                                    <div className="flex flex-col gap-3 w-[calc(100%-50px)] max-w-full min-w-0">
                                                        <Alert className="w-full">
                                                            <SparklesIcon className="w-4 h-4" />
                                                            <AlertTitle>{t('confirmExecutionTitle') || 'Confirm Execution'}</AlertTitle>
                                                            <AlertDescription>
                                                                {t('confirmExecutionDescription') || 'This operation will modify your database. Review and confirm to proceed.'}
                                                            </AlertDescription>
                                                        </Alert>

                                                        {/* SQL Query Toggle */}
                                                        <Button
                                                            variant="ghost"
                                                            size="sm"
                                                            onClick={() => setShowQueryForId(showQuery ? null : (chat.id ?? null))}
                                                            className="w-fit"
                                                        >
                                                            <CodeBracketIcon className="w-4 h-4 mr-2" />
                                                            {showQuery ? t('hideQuery') || 'Hide Query' : t('showQuery') || 'Show Query'}
                                                        </Button>

                                                        {/* SQL Query Display */}
                                                        {showQuery && (
                                                            <div className="relative w-full rounded-lg overflow-hidden">
                                                                <div className="h-[300px]">
                                                                    <CodeEditor value={chat.Text} language="sql" />
                                                                </div>
                                                                <button
                                                                    onClick={() => {
                                                                        copyToClipboard(chat.Text).then(success => {
                                                                            if (success) {
                                                                                setCopiedSqlId(chat.id ?? null);
                                                                                setTimeout(() => setCopiedSqlId(null), 2000);
                                                                            }
                                                                        });
                                                                    }}
                                                                    className="absolute top-2 right-2 flex items-center gap-1 px-1.5 py-0.5 rounded text-xs text-neutral-500 dark:text-neutral-400 hover:text-neutral-800 dark:hover:text-neutral-100 bg-neutral-200 dark:bg-neutral-700 hover:bg-neutral-300 dark:hover:bg-neutral-600 transition-colors z-10"
                                                                >
                                                                    {copiedSqlId === chat.id
                                                                        ? <><CheckCircleIcon className="w-3.5 h-3.5 text-green-500" /> <span className="text-green-500">Copied!</span></>
                                                                        : <><DocumentDuplicateIcon className="w-3.5 h-3.5" /> <span>Copy</span></>
                                                                    }
                                                                </button>
                                                            </div>
                                                        )}

                                                        { }
                                                        <div className="flex gap-2">
                                                            <Button
                                                                variant="outline"
                                                                onClick={() => chat.id && handleCancelSQL(chat.id)}
                                                                disabled={isExecuting}
                                                                size="sm"
                                                            >
                                                                {t('no') || 'No'}
                                                            </Button>
                                                            <Button
                                                                onClick={() => chat.id && handleConfirmSQL(chat.id, chat.Text, chat.Type)}
                                                                disabled={isExecuting}
                                                                size="sm"
                                                            >
                                                                {isExecuting ? (t('executing') || 'Executing...') : (t('yes') || 'Yes')}
                                                            </Button>
                                                        </div>
                                                    </div>
                                                </div>
                                            }
                                            return <div key={`chat-${i}`} className="flex gap-lg w-full max-w-full min-w-0 pt-4 relative" data-testid="table-message">
                                                {!chat.isUserInput && chats[i-1]?.isUserInput && (extensions.MetaIcon ?? <img src={logoImage} alt="clidey logo" className="w-auto h-8" />)}
                                                <TablePreview type={chat.Type} text={chat.Text} data={chat.Result} containerWidth={containerWidth} />
                                            </div>
                                        })
                                    }
                                    { loading &&  <div className="flex w-full mt-4">
                                        <Loading loadingText={loadingPhraseRef.current} size="sm" />
                                    </div> }
                                </div>
                            </div>
                        </div>
                    }
                </div>
                {
                    (models.length === 0 || (!modelAvailable && !currentModel)) &&
                    <EmptyState title={t('noModelTitle')} description={t('noModelDescription')} icon={<SparklesIcon className="w-16 h-16" data-testid="empty-state-sparkles-icon" />} />
                }
                <div className={classNames("flex justify-between items-center gap-2", {
                    "opacity-80": disableChat,
                    "opacity-10": disableAll,
                })}>
                    <Input
                        value={query}
                        onChange={e => setQuery(e.target.value)}
                        placeholder={t('placeholder')}
                        onSubmit={handleSubmitQuery}
                        disabled={disableAll}
                        onKeyDown={handleKeyDown}
                        onKeyUp={handleKeyUp}
                        autoComplete="off"
                        data-testid="chat-input"
                    />
                    <Tip className="w-fit">
                        <Button tabIndex={0} onClick={loading ? undefined : handleSubmitQuery} className={cn("rounded-full", {
                            "opacity-50": loading,
                        })} disabled={disableChat} variant={disableChat ? "secondary" : undefined} data-testid="icon-button" aria-label={t('sendMessage')}>
                            <ArrowUpCircleIcon className="w-8 h-8" aria-hidden="true" />
                        </Button>
                        <p>{t('sendMessage')}</p>
                    </Tip>
                </div>
            </div>
        </InternalPage>
    )
}