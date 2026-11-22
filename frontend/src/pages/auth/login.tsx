import { useLazyQuery, useMutation, useQuery } from "@apollo/client/react";
import {Badge, Button, Card, cn, Input, Label, ModeToggle, Separator, toast, useTheme} from '@clidey/ux';
import {SearchSelect} from '../../components/ux';
import {
    SettingsConfigDocument,
    SourceFieldOptionsDocument,
    SourceProfilesDocument,
    LoginSourceDocument,
    LoginWithSourceProfileDocument,
    SourceConnectionFieldKind,
    SourceConnectionFieldSection,
    SourceConnectionTransport,
    SourceHostInputMode,
    SourceHostInputUrlParser,
} from '@graphql';
import camelCase from "lodash/camelCase";
import classNames from "classnames";
import {FC, ReactElement, Suspense, useCallback, useEffect, useMemo, useRef, useState} from "react";
import {getComponent} from "../../config/component-registry";
import {useNavigate, useSearchParams} from "react-router-dom";
import logoImage from "../../../public/images/logo.svg";
import {
    AdjustmentsHorizontalIcon,
    ChatBubbleLeftRightIcon,
    CheckCircleIcon,
    ChevronDownIcon,
    CircleStackIcon,
    CodeBracketIcon,
    ShareIcon,
    SparklesIcon,
    TableCellsIcon
} from '../../components/heroicons';
import {Icons} from "../../components/icons";
import {Loading} from "../../components/loading";
import {Container} from "../../components/page";
import {updateProfileLastAccessed} from "../../components/profile-info-tooltip";
import {SourceTypeItem} from "../../config/source-types";
import {extensions, featureFlags, getAppName, sources} from '../../config/features';
import {InternalRoutes} from "../../config/routes";
import {useSourceTypeItems} from "../../hooks/useSourceCatalog";
import {useDesktopFile} from '../../hooks/useDesktop';
import {useTranslation} from '@/hooks/use-translation';
import {AuthActions} from "../../store/auth";
import {DatabaseActions} from "../../store/database";
import {TourActions} from "../../store/tour";
import {SettingsActions} from "../../store/settings";
import {isSupportedLanguage} from "@/utils/languages";
import {HealthActions} from "../../store/health";
import {useAppDispatch, useAppSelector} from "../../store/hooks";
import {isDesktopApp} from '../../utils/external-links';
import {v4 as uuidv4} from 'uuid';
import {hasCompletedOnboarding, markOnboardingComplete} from '../../utils/onboarding';
import {
    AwsConnectionPicker,
    DatabaseIconWithBadge,
    isAwsConnection
} from '../../components/aws';
import {AzureConnectionPicker, isAzureConnection} from '../../components/azure';
import {
    GcpConnectionPicker,
    isGcpConnection,
} from '../../components/gcp';
import {ConnectionPrefillData, isAwsHostname, isAzureHostname, isGcpHostname} from '../../utils/cloud-connection-prefill';
import {SSL_KEYS, SSLConfig} from '../../components/ssl-config';
import { clearGraphqlStore } from '@/config/graphql-client';
import {
    buildRecordInputs,
    buildSourceValues,
    createProfilePayload,
    createProfilePayloadFromSourceProfile,
    getValue,
} from '../../utils/source-credentials';

const LOGIN_RESERVED_PARAMS = new Set([
    "type", "host", "username", "password", "database",
    "port", "region", "search_path",
    "login", "resource", "credentials",
]);

const LOGIN_UI_PARAMS = new Set(["locale", "mode", "theme", "os"]);

const EMPTY_DATABASE_TYPE: SourceTypeItem = {
    id: "",
    label: "",
    connector: "",
    icon: <span className="w-6 h-6" />,
    extra: {},
    fields: {},
    requiredFields: {},
};

type SourceConnectionFieldItem = NonNullable<SourceTypeItem["connectionFields"]>[number];

function usesFileTransport(databaseType: SourceTypeItem): boolean {
    return databaseType.traits?.connection.transport === SourceConnectionTransport.File;
}

function findConnectionFieldByKey(
    databaseType: SourceTypeItem,
    key: string
): SourceConnectionFieldItem | undefined {
    return databaseType.connectionFields?.find(field =>
        field.Key.toLowerCase() === key.toLowerCase()
    );
}

function supportsDatabaseFieldOptions(databaseType: SourceTypeItem): boolean {
    return databaseType.connectionFields?.some(field =>
        field.Key.toLowerCase() === "database" && field.SupportsOptions
    ) ?? false;
}

function canSubmitDatabaseCredentials(
    databaseType: SourceTypeItem,
    hostName: string,
    username: string,
    password: string,
    database: string
): boolean {
    const requiredFields = databaseType.requiredFields ?? {};

    const hostnameOk = !requiredFields.hostname || hostName.length > 0;
    const usernameOk = !requiredFields.username || username.length > 0;
    const passwordOk = !requiredFields.password || password.length > 0;
    const databaseOk = !requiredFields.database || database.length > 0;

    return hostnameOk && usernameOk && passwordOk && databaseOk;
}

function generateCredentialId(type: string, hostname: string, username: string, database: string): string | undefined {

    if (!isDesktopApp()) {
        return uuidv4();
    }

    const parts = [
        'whodb',
        type || 'unknown',
        hostname || 'localhost',
        username || 'default',
        database || 'default'
    ];

    const combined = parts.join('::');
    try {
        const encoded = btoa(combined).replace(/[+/=]/g, '');
        return encoded.substring(0, 16).toLowerCase();
    } catch {
        return uuidv4();
    }
}

export interface LoginFormProps {

    onLoginSuccess?: () => void;

    hideHeader?: boolean;

    compact?: boolean;

    className?: string;
    advancedDirection?: "horizontal" | "vertical";
}

export const LoginForm: FC<LoginFormProps> = ({
    onLoginSuccess,
    hideHeader = false,
    className = "",
    advancedDirection = "horizontal",
}) => {
    const { t } = useTranslation('pages/login');
    const appName = getAppName();
    const dispatch = useAppDispatch();
    const navigate = useNavigate();
    const currentProfile = useAppSelector(state => state.auth.current);
    const shouldUpdateLastAccessed = useRef(false);
    const usernameInputRef = useRef<HTMLInputElement>(null);
    const handleSubmitRef = useRef<() => void>(() => {});
    const handleLoginWithSourceProfileSubmitRef = useRef<(overrideProfileId?: string) => void>(() => {});
    const [pendingAutoLogin, setPendingAutoLogin] = useState(false);
    const { setTheme } = useTheme();

    const FIRST_LOGIN_KEY = 'whodb_has_logged_in';
    const [isFirstLogin, setIsFirstLogin] = useState(() => {
        return !localStorage.getItem(FIRST_LOGIN_KEY);
    });

    const [login, { loading: loginLoading }] = useMutation(LoginSourceDocument);
    const [loginWithSourceProfile, { loading: loginWithSourceProfileLoading }] = useMutation(LoginWithSourceProfileDocument);
    const [getDatabases, { loading: databasesLoading, data: foundDatabases }] = useLazyQuery(SourceFieldOptionsDocument);
    const { loading: profilesLoading, data: profiles } = useQuery(SourceProfilesDocument);
    const { data: settingsData } = useQuery(SettingsConfigDocument);
    const cloudProvidersEnabled = settingsData?.SettingsConfig?.CloudProvidersEnabled ?? false;
    const disableCredentialForm = settingsData?.SettingsConfig?.DisableCredentialForm ?? false;
    const maxPageSize = settingsData?.SettingsConfig?.MaxPageSize ?? 10000;
    const {
        items: databaseTypeItems,
        loading: databaseTypesLoading,
        error: databaseTypesError,
    } = useSourceTypeItems({ cloudProvidersEnabled });
    const [searchParams, setSearchParams] = useSearchParams();

    const databaseTypesLoaded = !databaseTypesLoading;
    const [databaseType, setDatabaseType] = useState<SourceTypeItem>(EMPTY_DATABASE_TYPE);
    const [hostName, setHostName] = useState("");
    const [database, setDatabase] = useState("");
    const [username, setUsername] = useState("");
    const [password, setPassword] = useState("");
    const [error, setError] = useState<string>();
    const [missingDriver, setMissingDriver] = useState<string | null>(null);
    const [advancedForm, setAdvancedForm] = useState<Record<string, string>>({});
    const [showAdvanced, setShowAdvanced] = useState(false);
    const [formResetKey, setFormResetKey] = useState(0);
    const [selectedAvailableProfile, setSelectedAvailableProfile] = useState<string>();
    const [isAutoLoggingIn, setIsAutoLoggingIn] = useState(() => {

        return searchParams.has("resource") || searchParams.has("login");
    });
    const [isEmbedded] = useState(() => {
        return searchParams.has("credentials") || searchParams.has("resource") || searchParams.has("login");
    });
    const { isDesktop, selectDatabaseFile } = useDesktopFile();

    useEffect(() => {
        dispatch(SettingsActions.setCloudProvidersEnabled(cloudProvidersEnabled));
        dispatch(SettingsActions.setMaxPageSize(maxPageSize));
    }, [cloudProvidersEnabled, maxPageSize, dispatch]);

    useEffect(() => {
        if (databaseTypeItems.length === 0) {
            return;
        }

        const nextType = databaseTypeItems.find(item => item.id === databaseType.id) ?? databaseTypeItems[0];
        if (databaseType.id !== nextType.id) {
            setDatabaseType(nextType);
            setAdvancedForm(nextType.extra ?? {});
        }
    }, [databaseType.id, databaseTypeItems]);

    useEffect(() => {
        if (databaseTypesError) {
            console.error('Failed to load source catalog:', databaseTypesError);
        }
    }, [databaseTypesError]);

    const loading = useMemo(() => {
        return loginLoading || loginWithSourceProfileLoading || isAutoLoggingIn;
    }, [loginLoading, loginWithSourceProfileLoading, isAutoLoggingIn]);

    const markFirstLoginComplete = useCallback(() => {
        if (isFirstLogin) {
            localStorage.setItem(FIRST_LOGIN_KEY, 'true');
            setIsFirstLogin(false);
            markOnboardingComplete();
        }
    }, [isFirstLogin, FIRST_LOGIN_KEY]);

    const handleLoginError = useCallback((loginError: unknown, allowDriverInstallPrompt = false) => {
        setIsAutoLoggingIn(false);

        const errorMessage = loginError instanceof Error ? loginError.message : String(loginError);

        if (allowDriverInstallPrompt) {
            const driverMatch = errorMessage.match(/driver_not_installed:(\w+)/);
            if (driverMatch) {
                setMissingDriver(driverMatch[1]);
                return;
            }
        }

        const lowerCaseMessage = errorMessage.toLowerCase();
        const isNetworkError = lowerCaseMessage.includes('network') ||
            lowerCaseMessage.includes('fetch') ||
            lowerCaseMessage.includes('econnrefused') ||
            (loginError instanceof Error && 'statusCode' in loginError);

        if (isNetworkError) {
            dispatch(HealthActions.setHealthStatus({
                server: 'error',
                database: 'unavailable',
            }));
        }

        toast.error(t('loginFailedWithError', { error: errorMessage }));
    }, [dispatch, t]);

    const handleSubmit = useCallback(() => {
        if (databaseType.id === "" || !canSubmitDatabaseCredentials(databaseType, hostName, username, password, database)) {
            setIsAutoLoggingIn(false);
            return setError(t('allFieldsRequired'));
        }
        setError(undefined);

        const credentialId = generateCredentialId(databaseType.id, hostName, username, database);

        const values = buildSourceValues(hostName, database, username, password, advancedForm);
        const credentials = {
            Id: credentialId,
            SourceType: databaseType.id,
            Values: buildRecordInputs(hostName, database, username, password, advancedForm),
        };

        void (async () => {
            try {
                const { data } = await login({
                    variables: {
                        credentials,
                    },
                });

                if (!data?.LoginSource.Status) {
                    setIsAutoLoggingIn(false);
                    toast.error(t('loginFailed'));
                    return;
                }

                const sslMode = advancedForm[SSL_KEYS.MODE];
                const profileData = createProfilePayload(credentialId, databaseType.id, values, {
                    SSLConfigured: sslMode != null && sslMode !== 'disabled' && sslMode !== '',
                });

                await clearGraphqlStore();
                shouldUpdateLastAccessed.current = true;
                dispatch(AuthActions.login(profileData));
                markFirstLoginComplete();

                const hasLoginParams = [...searchParams.keys()].some(k => !LOGIN_UI_PARAMS.has(k));
                if (hasLoginParams) {
                    const newParams = new URLSearchParams();
                    LOGIN_UI_PARAMS.forEach(key => {
                        if (searchParams.has(key)) {
                            newParams.set(key, searchParams.get(key)!);
                        }
                    });
                    setSearchParams(newParams, { replace: true });
                }

                if (onLoginSuccess) {
                    onLoginSuccess();
                } else {
                    navigate(InternalRoutes.Dashboard.StorageUnit.path);
                }
                toast.success(t('loginSuccessful'));
            } catch (error) {
                handleLoginError(error, true);
            }
        })();
    }, [advancedForm, database, databaseType, dispatch, handleLoginError, hostName, login, markFirstLoginComplete, navigate, onLoginSuccess, password, searchParams, setSearchParams, t, username]);

    const handleLoginWithSourceProfileSubmit = useCallback((overrideProfileId?: string) => {
        const profileId = overrideProfileId ?? selectedAvailableProfile;
        if (profileId == null) {
            return setError(t('selectProfile'));
        }
        setError(undefined);

        const profile = profiles?.SourceProfiles.find(p => p.Id === profileId);

        void (async () => {
            try {
                const { data } = await loginWithSourceProfile({
                    variables: {
                        profile: {
                            Id:  profileId,
                        },
                    },
                });

                if (!data?.LoginWithSourceProfile.Status) {
                    setIsAutoLoggingIn(false);
                    toast.error(t('loginFailed'));
                    return;
                }

                updateProfileLastAccessed(profileId);
                await clearGraphqlStore();
                if (profile != null) {
                    dispatch(AuthActions.login(createProfilePayloadFromSourceProfile(profile)));
                }
                markFirstLoginComplete();

                if (searchParams.has("resource")) {
                    const newParams = new URLSearchParams(searchParams);
                    newParams.delete("resource");
                    setSearchParams(newParams, { replace: true });
                }

                if (onLoginSuccess) {
                    onLoginSuccess();
                } else {
                    navigate(InternalRoutes.Dashboard.StorageUnit.path);
                }
                toast.success(t('loginSuccessful'));
            } catch (error) {
                handleLoginError(error);
            }
        })();
    }, [dispatch, handleLoginError, loginWithSourceProfile, markFirstLoginComplete, navigate, onLoginSuccess, profiles?.SourceProfiles, searchParams, selectedAvailableProfile, setSearchParams, t]);

    handleSubmitRef.current = handleSubmit;
    handleLoginWithSourceProfileSubmitRef.current = handleLoginWithSourceProfileSubmit;

    const handleSampleDatabaseLogin = useCallback(() => {
        const sampleProfile = profiles?.SourceProfiles.find(p => p.Source === "builtin");
        if (!sampleProfile) {
            return toast.error(t('sampleDatabaseNotFound'));
        }

        setError(undefined);

        void (async () => {
            try {
                const { data } = await loginWithSourceProfile({
                    variables: {
                        profile: {
                            Id: sampleProfile.Id,
                        },
                    },
                });

                if (!data?.LoginWithSourceProfile.Status) {
                    setIsAutoLoggingIn(false);
                    toast.error(t('loginFailed'));
                    return;
                }

                updateProfileLastAccessed(sampleProfile.Id);
                await clearGraphqlStore();
                dispatch(AuthActions.login(createProfilePayloadFromSourceProfile(sampleProfile)));
                markFirstLoginComplete();
                if (featureFlags.autoStartTourOnLogin) {
                    dispatch(TourActions.scheduleTourOnLoad('sample-database-tour'));
                }
                if (onLoginSuccess) {
                    onLoginSuccess();
                } else {
                    navigate(InternalRoutes.Dashboard.StorageUnit.path);
                }
                toast.success(t('welcomeToWhodb', { appName }));
            } catch (error) {
                handleLoginError(error);
            }
        })();
    }, [appName, dispatch, handleLoginError, loginWithSourceProfile, markFirstLoginComplete, navigate, onLoginSuccess, profiles?.SourceProfiles, t]);

    const handleDatabaseTypeChange = useCallback((item: SourceTypeItem) => {
        setHostName("");
        setUsername("");
        setPassword("");
        setDatabase("");
        setDatabaseType(item);
        setAdvancedForm(item.extra ?? {});
        setFormResetKey(k => k + 1);
    }, []);

    const handleAdvancedToggle = useCallback(() => {
        setShowAdvanced(a => !a);
    }, []);

    useEffect(() => {
        if (supportsDatabaseFieldOptions(databaseType)) {
            getDatabases({ variables: { sourceType: databaseType.id } });
        }
    }, [databaseType, getDatabases, formResetKey]);

    const handleAdvancedForm = useCallback((key: string, value: string) => {
        setAdvancedForm(form => {
            const newForm = {...form};
            newForm[key] = value;
            return newForm;
        });
    }, []);

    const handleAvailableProfileChange = useCallback((itemId: string) => {
        setSelectedAvailableProfile(itemId);
    }, []);

    const handleCloudConnectionPrefill = useCallback((data: ConnectionPrefillData) => {

        const dbType = databaseTypeItems.find(item =>
            item.id.toLowerCase() === data.databaseType.toLowerCase()
        );

        if (dbType) {

            handleDatabaseTypeChange(dbType);

            setTimeout(() => {
                if (data.hostname) {
                    setHostName(data.hostname);
                }

                if (data.advanced && Object.keys(data.advanced).length > 0) {
                    setAdvancedForm(prev => ({
                        ...prev,
                        ...data.advanced,
                    }));
                    setShowAdvanced(true);
                }

                setTimeout(() => {
                    usernameInputRef.current?.focus();
                }, 50);
            }, 0);
        }
    }, [databaseTypeItems, handleDatabaseTypeChange]);

    const handleBrowseDatabaseFile = useCallback(async () => {
        try {
            const filePath = await selectDatabaseFile(databaseType.id);
            if (filePath) {
                setDatabase(filePath);
            }
        } catch (error) {
            console.error('Failed to select database file:', error);
            toast.error(t('failedToSelectDatabaseFile'));
        }
    }, [selectDatabaseFile, databaseType.id, t]);

    useEffect(() => {
        dispatch(DatabaseActions.setSchema(""));
    }, [dispatch]);

    useEffect(() => {
        const hasAutoLoginParams = searchParams.has("credentials") ||
                                   searchParams.has("resource") ||
                                   searchParams.has("login");
        if (hasAutoLoginParams) {
            dispatch(AuthActions.setEmbedded(true));
        }
    }, [searchParams, dispatch]);

    useEffect(() => {
        if (searchParams.has("locale")) {
            const locale = searchParams.get("locale");
            if (locale && isSupportedLanguage(locale)) {
                dispatch(SettingsActions.setLanguage(locale));
            }
        }
    }, [searchParams, dispatch]);

    useEffect(() => {
        if (searchParams.has("mode")) {
            const mode = searchParams.get("mode")?.toLowerCase();
            if (mode === 'light' || mode === 'dark' || mode === 'system') {
                setTheme(mode);
            }
        }
    }, [searchParams, setTheme]);

    useEffect(() => {
        if (searchParams.has("theme")) {
            const theme = searchParams.get("theme")?.toLowerCase();
            if (theme === 'default') {
                dispatch(SettingsActions.setAppTheme('default'));
            }
        }
    }, [searchParams, dispatch]);

    useEffect(() => {
        if (searchParams.has("os")) {
            const os = searchParams.get("os")?.toLowerCase();
            if (os === 'linux' || os === 'macos' || os === 'windows') {
                dispatch(SettingsActions.setOS(os));
            }
        }
    }, [searchParams, dispatch]);

    useEffect(() => {
        if (shouldUpdateLastAccessed.current && currentProfile?.Id) {
            updateProfileLastAccessed(currentProfile.Id);
            shouldUpdateLastAccessed.current = false;
        }
    }, [currentProfile]);

    const availableProfiles = useMemo(() => {
        return profiles?.SourceProfiles
            .filter(profile => profile.Source !== "builtin")

            .filter(profile => {
                if (cloudProvidersEnabled) {
                    return true;
                }

                const hostname = getValue(profile.Values, "Hostname");
                return !isAwsHostname(hostname) && !isAzureHostname(hostname) && !isGcpHostname(hostname);
            })
            .map(profile => ({
                value: profile.Id,
                label: profile.Alias ?? profile.Id,
                icon: (
                    <DatabaseIconWithBadge
                        icon={(Icons.Logos as Record<string, ReactElement>)[profile.Type]}
                        showCloudBadge={isAwsConnection(profile.Id) || isAzureConnection(profile.Id) || isGcpConnection(profile.Id)}
                        sslStatus={profile.SSLConfigured ? { IsEnabled: true, Mode: 'configured' } : undefined}
                        size="sm"
                    />
                ),
                rightIcon: sources[profile.Source],
            })) ?? [];
    }, [profiles?.SourceProfiles, cloudProvidersEnabled]);

    const hasAvailableProfiles = availableProfiles.length > 0;

    const sampleProfile = useMemo(() => {
        return profiles?.SourceProfiles.find(p => p.Source === "builtin");
    }, [profiles?.SourceProfiles]);

    useEffect(() => {
        if (searchParams.size === 0) {
            return;
        }

        if (!databaseTypesLoaded) {
            return;
        }

        if (searchParams.has("credentials")) {
            try {
                const credentialsBase64 = searchParams.get("credentials")!;
                const credentialsJson = atob(credentialsBase64);
                const credentials = JSON.parse(credentialsJson);

                if (credentials.type) {
                    const dbType = databaseTypeItems.find(item =>
                        item.id.toLowerCase() === credentials.type.toLowerCase()
                    );
                    if (dbType) {
                        handleDatabaseTypeChange(dbType);
                    }
                }
                if (credentials.host) setHostName(credentials.host);
                if (credentials.username) setUsername(credentials.username);
                if (credentials.password) setPassword(credentials.password);
                if (credentials.database) setDatabase(credentials.database);

                if (credentials.port) {
                    setAdvancedForm(prev => ({...prev, 'Port': String(credentials.port)}));
                    setShowAdvanced(true);
                }

                if (credentials.config && typeof credentials.config === 'object') {
                    const advancedFormData: Record<string, string> = {};
                    for (const [key, value] of Object.entries(credentials.config)) {
                        advancedFormData[key] = String(value);
                    }

                    if (credentials.port) {
                        advancedFormData['Port'] = String(credentials.port);
                    }
                    setAdvancedForm(advancedFormData);
                    if (Object.keys(advancedFormData).length > 0) {
                        setShowAdvanced(true);
                    }
                }
            } catch (error) {
                console.error('Failed to parse credentials:', error);
                toast.error(t('failedToParseCredentials'));
            }
        } else {

            if (searchParams.has("type")) {
                const typeParam = searchParams.get("type")!;
                const dbType = databaseTypeItems.find(item =>
                    item.id.toLowerCase() === typeParam.toLowerCase()
                );
                if (dbType) {
                    handleDatabaseTypeChange(dbType);
                }
            }

            if (searchParams.has("host")) setHostName(searchParams.get("host")!);
            if (searchParams.has("username")) setUsername(searchParams.get("username")!);
            if (searchParams.has("password")) setPassword(searchParams.get("password")!);
            if (searchParams.has("database")) setDatabase(searchParams.get("database")!);

            const hasPort = searchParams.has("port");
            const hasRegion = searchParams.has("region");
            const hasSearchPath = searchParams.has("search_path");
            if (hasPort || hasRegion || hasSearchPath) {
                setAdvancedForm(prev => ({
                    ...prev,
                    ...(hasPort ? {'Port': searchParams.get("port")!} : {}),
                    ...(hasRegion ? {'Region': searchParams.get("region")!} : {}),
                    ...(hasSearchPath ? {'Search Path': searchParams.get("search_path")!} : {}),
                }));
            }

            const advancedEntries: Record<string, string> = {};
            searchParams.forEach((value, key) => {
                if (!LOGIN_RESERVED_PARAMS.has(key) && !LOGIN_UI_PARAMS.has(key)) {
                    advancedEntries[key] = value;
                }
            });
            if (Object.keys(advancedEntries).length > 0) {
                setAdvancedForm(prev => ({...prev, ...advancedEntries}));
                setShowAdvanced(true);
            }
        }

        if (searchParams.has("resource")) {
            const selectedProfile = availableProfiles.find(profile => profile.value === searchParams.get("resource"));
            if (selectedProfile?.value) {
                setSelectedAvailableProfile(selectedProfile.value);
                handleLoginWithSourceProfileSubmitRef.current(selectedProfile.value);
            }
        } else if (searchParams.has("login")) {
            setPendingAutoLogin(true);
            const newParams = new URLSearchParams(searchParams);
            newParams.delete("login");
            setSearchParams(newParams, { replace: true });
        }
    }, [searchParams, databaseTypeItems, databaseTypesLoaded, profiles?.SourceProfiles, availableProfiles, handleDatabaseTypeChange]);

    useEffect(() => {
        if (!pendingAutoLogin) return;
        setPendingAutoLogin(false);
        handleSubmitRef.current();
    }, [pendingAutoLogin]);

    const handleHostNameChange = useCallback((newHostName: string) => {
        const urlParser = databaseType.traits?.connection.hostInputUrlParser ?? SourceHostInputUrlParser.None;

        if (urlParser === SourceHostInputUrlParser.MongoSrv && newHostName.startsWith("mongodb+srv://")) {
            const url = new URL(newHostName);
            setHostName(url.hostname);
            setUsername(url.username);
            setPassword(url.password);
            setDatabase(url.pathname.substring(1));
            const advancedForm = {
                "Port": "27017",
                "URL Params": `?${url.searchParams.toString()}`,
                "DNS Enabled": "false"
            };
            if (url.port.length === 0) {
                advancedForm["Port"] = "";
                advancedForm["DNS Enabled"] = "true";
            }
            setAdvancedForm(advancedForm);
            setShowAdvanced(true);
            return;
        }

        if (urlParser === SourceHostInputUrlParser.Postgres && (newHostName.startsWith("postgres://") || newHostName.startsWith("postgresql://"))) {
            try {
                const url = new URL(newHostName);
                const hostname = url.hostname;
                const username = url.username;
                const password = url.password;
                const database = url.pathname.substring(1);

                if (!hostname || !username || !password || !database) {
                    toast.warning(t('urlParseWarning'));
                }
                setHostName(hostname);
                setUsername(username);
                setPassword(password);
                setDatabase(database);

                if (url.port) {
                    const advancedForm = {
                        "Port": url.port,
                        "SSL Mode": "disable"
                    };
                    setAdvancedForm(advancedForm);
                    setShowAdvanced(true);
                }
            } catch (error) {
                toast.warning(t('urlParseWarning'));
            }
            return;
        }

        setHostName(newHostName);
    }, [databaseType.traits?.connection.hostInputUrlParser, t]);

    const fields = useMemo(() => {
        if (databaseType.customFormRenderer) {
            const CustomForm = databaseType.customFormRenderer;
            return <CustomForm
                key={formResetKey}
                hostName={hostName}
                setHostName={setHostName}
                username={username}
                setUsername={setUsername}
                password={password}
                setPassword={setPassword}
                advancedForm={advancedForm}
                setAdvancedForm={setAdvancedForm}
            />;
        }

        const hostnameField = findConnectionFieldByKey(databaseType, "Hostname");
        const usernameField = findConnectionFieldByKey(databaseType, "Username");
        const passwordField = findConnectionFieldByKey(databaseType, "Password");
        const databaseField = findConnectionFieldByKey(databaseType, "Database");
        const searchPathField = findConnectionFieldByKey(databaseType, "Search Path");

        if (usesFileTransport(databaseType)) {
            return <div className="flex flex-col gap-lg w-full">
                <div className="flex flex-col gap-xs w-full">
                    <Label htmlFor="sqlite-database">{t('database')}</Label>
                    {isDesktop ? (
                        <div className="flex flex-col gap-sm w-full">
                            <Input
                                id="sqlite-database"
                                value={database}
                                onChange={(e) => setDatabase(e.target.value)}
                                placeholder={databaseField?.PlaceholderKey ? t(databaseField.PlaceholderKey) : t('selectOrEnterDatabasePath')}
                                data-testid="database"
                                aria-required={databaseField?.Required ? "true" : undefined}
                                aria-invalid={error ? "true" : undefined}
                                aria-describedby={error ? "login-error" : undefined}
                            />
                            <Button
                                onClick={handleBrowseDatabaseFile}
                                variant="outline"
                                className="w-full"
                            >
                                {t('browseForSqliteFile')}
                            </Button>
                        </div>
                    ) : (
                        <SearchSelect
                            value={database}
                            onChange={setDatabase}
                            disabled={databasesLoading}
                            options={
                                databasesLoading
                                    ? []
                                    : foundDatabases?.SourceFieldOptions?.map(db => ({
                                    value: db,
                                    label: db,
                                    icon: <CircleStackIcon className="w-4 h-4"/>,
                                })) ?? []
                            }
                            placeholder={t('selectDatabase')}
                            buttonProps={{
                                "data-testid": "database",
                                "aria-required": databaseField?.Required ? "true" : undefined,
                                "aria-invalid": error ? "true" : undefined,
                                "aria-describedby": error ? "login-error" : undefined,
                            }}
                            contentClassName="w-[var(--radix-popover-trigger-width)] login-select-popover"
                            rightIcon={<ChevronDownIcon className="w-4 h-4"/>}
                        />
                    )}
                </div>
            </div>
        }
        return <div className="flex flex-col gap-lg w-full">
            { databaseType.fields?.hostname && (
                <div className="flex flex-col gap-sm w-full">
                    <Label htmlFor="login-hostname">{databaseType.traits?.connection.hostInputMode === SourceHostInputMode.HostnameOrUrl ? t('hostNameOrUrl') : t('hostName')}</Label>
                    <Input id="login-hostname" value={hostName} onChange={(e) => handleHostNameChange(e.target.value)} data-testid="hostname" placeholder={hostnameField?.PlaceholderKey ? t(hostnameField.PlaceholderKey) : t('enterHostName')} aria-required={hostnameField?.Required ? "true" : undefined} aria-invalid={error ? "true" : undefined} aria-describedby={error ? "login-error" : undefined} />
                </div>
            )}
            { databaseType.fields?.username && (
                <div className="flex flex-col gap-sm w-full">
                    <Label htmlFor="login-username">{t('username')}</Label>
                    <Input ref={usernameInputRef} id="login-username" value={username} onChange={(e) => setUsername(e.target.value)} data-testid="username" placeholder={usernameField?.PlaceholderKey ? t(usernameField.PlaceholderKey) : t('enterUsername')} aria-required={usernameField?.Required ? "true" : undefined} aria-invalid={error ? "true" : undefined} aria-describedby={error ? "login-error" : undefined} />
                </div>
            )}
            { databaseType.fields?.password && (
                <div className="flex flex-col gap-sm w-full">
                    <Label htmlFor="login-password">{t('password')}</Label>
                    <Input id="login-password" value={password} onChange={(e) => setPassword(e.target.value)} type="password" data-testid="password" placeholder={passwordField?.PlaceholderKey ? t(passwordField.PlaceholderKey) : t('enterPassword')} aria-required={passwordField?.Required ? "true" : undefined} aria-invalid={error ? "true" : undefined} aria-describedby={error ? "login-error" : undefined} showPasswordToggle={!isEmbedded} />
                </div>
            )}
            { databaseType.fields?.database && (
                <div className="flex flex-col gap-sm w-full">
                    <Label htmlFor="login-database">{t('database')}</Label>
                    <Input id="login-database" value={database} onChange={(e) => setDatabase(e.target.value)} data-testid="database" placeholder={databaseField?.PlaceholderKey ? t(databaseField.PlaceholderKey) : t('enterDatabase')} aria-required={databaseField?.Required ? "true" : undefined} aria-invalid={error ? "true" : undefined} aria-describedby={error ? "login-error" : undefined} />
                </div>
            )}
            { databaseType.fields?.searchPath && (
                <div className="flex flex-col gap-sm w-full">
                    <Label htmlFor="login-search-path">{searchPathField?.LabelKey ? t(searchPathField.LabelKey) : t(`advancedFields.${camelCase('Search Path')}`)}</Label>
                    <Input id="login-search-path" value={advancedForm['Search Path'] ?? ''} onChange={(e) => handleAdvancedForm('Search Path', e.target.value)} data-testid="search-path" placeholder={searchPathField?.PlaceholderKey ? t(searchPathField.PlaceholderKey) : t('enterSearchPath')} aria-required={searchPathField?.Required ? "true" : undefined} />
                </div>
            )}
        </div>
    }, [database, databaseType, databasesLoading, foundDatabases?.SourceFieldOptions, handleHostNameChange, hostName, password, username, isDesktop, handleBrowseDatabaseFile, advancedForm, formResetKey, t, error, isEmbedded]);

    const loginWithCredentialsEnabled = useMemo(() => {
        if (databaseType.customFormRenderer) {
            return hostName.length > 0 || Object.keys(advancedForm).length > 0;
        }
        return canSubmitDatabaseCredentials(databaseType, hostName, username, password, database);
    }, [databaseType, hostName, username, password, database, advancedForm]);

    const loginWithSourceProfileEnabled = useMemo(() => {
        return selectedAvailableProfile != null;
    }, [selectedAvailableProfile]);

    const sslAdvancedKeys = useMemo(() => new Set<string>(Object.values(SSL_KEYS)), []);

    const declaredAdvancedFields = useMemo(() => {
        return (databaseType.connectionFields ?? []).filter(field =>
            field.Section === SourceConnectionFieldSection.Advanced &&
            !sslAdvancedKeys.has(field.Key)
        );
    }, [databaseType.connectionFields, sslAdvancedKeys]);

    const declaredAdvancedFieldKeys = useMemo(() => {
        return new Set(declaredAdvancedFields.map(field => field.Key));
    }, [declaredAdvancedFields]);

    const excludedAdvancedKeys = useMemo(() => {
        const excluded = new Set<string>(sslAdvancedKeys);
        for (const field of databaseType.connectionFields ?? []) {
            if (field.Section !== SourceConnectionFieldSection.Advanced) {
                excluded.add(field.Key);
            }
        }
        return excluded;
    }, [databaseType.connectionFields, sslAdvancedKeys]);

    const fallbackAdvancedEntries = useMemo(() => {
        return Object.entries(advancedForm).filter(([key]) =>
            !excludedAdvancedKeys.has(key) && !declaredAdvancedFieldKeys.has(key)
        );
    }, [advancedForm, declaredAdvancedFieldKeys, excludedAdvancedKeys]);

    const hasAdvancedSection = declaredAdvancedFields.length > 0 || fallbackAdvancedEntries.length > 0 || (databaseType.sslModes?.length ?? 0) > 0;

    if (!databaseTypesLoaded || isAutoLoggingIn || loading || profilesLoading)  {
        return (
            <div className={classNames("flex flex-col justify-center items-center gap-lg w-full", className)}>
                <div>
                    <Loading size="lg" />
                </div>
                <h1 className="text-xl">
                    {t('loggingIn')}
                </h1>
            </div>
        );
    }

    const showSidePanel = sampleProfile && !hideHeader && featureFlags.sampleDatabaseTour && isFirstLogin && !hasCompletedOnboarding();

    return (
        <div className={classNames("w-fit h-fit", className, {
            "w-full h-full": advancedDirection === "vertical",
            "flex gap-8": showSidePanel && advancedDirection === "horizontal",
        })} data-testid="login-form-container">
            <div className={cn("fixed top-4 right-4 z-20", {
                "hidden": !showSidePanel,
            })} data-testid="mode-toggle-login">
                <ModeToggle />
            </div>
            <div className={classNames("flex flex-col grow gap-lg", {
                "justify-between": advancedDirection === "horizontal",
                "h-full": advancedDirection === "vertical" && availableProfiles.length === 0,
            })}>
                {!hideHeader && (
                    <header className="flex justify-between" data-testid="login-header">
                        <h1 className="flex items-center gap-xs text-xl">
                            {extensions.Logo ?? <img src={logoImage} alt="WhoDB" className="w-auto h-8 mr-1"/>}
                            <span className="text-brand-foreground" data-testid="app-name">{getAppName()}</span>
                        </h1>
                        <span className="text-xl">{t('title')}</span>
                        {
                            error &&
                            <Badge id="login-error" variant="destructive" className="self-end" role="alert">
                                {error}
                            </Badge>
                        }
                    </header>
                )}
                <div className={classNames("flex", {
                    "flex-row grow": advancedDirection === "horizontal",
                    "flex-col w-full gap-lg": advancedDirection === "vertical",
                })} data-testid="login-form">
                    <div className={classNames("flex flex-col gap-lg grow", advancedDirection === "vertical" ? "w-full" : "w-[350px]")}>
                        <div className={cn("flex flex-col grow gap-lg", {
                            "justify-center": advancedDirection === "horizontal" && !showSidePanel,
                        })}>
                            {disableCredentialForm && !hasAvailableProfiles ? (
                                <Card className="p-6 max-w-md">
                                    <h1 className="text-xl font-semibold">
                                        {t('noConnectionsTitle')}
                                    </h1>
                                    <p className="mt-2 text-sm text-muted-foreground">
                                        {t('noConnectionsDescription')}
                                    </p>
                                </Card>
                            ) : (
                                <>
                                    {!disableCredentialForm && (
                                        <div className="flex flex-col gap-sm w-full">
                                            <Label>{t('databaseType')}</Label>
                                            <SearchSelect
                                                value={databaseType?.id || ""}
                                                onChange={(value) => {
                                                    const selected = databaseTypeItems.find(item => item.id === value);
                                                    if (selected) {
                                                        handleDatabaseTypeChange(selected);
                                                    }
                                                }}
                                                options={databaseTypeItems.map(item => ({
                                                    value: item.id,
                                                    label: item.label,
                                                    icon: item.icon,
                                                }))}
                                                buttonProps={{
                                                    "data-testid": "database-type-select",
                                                }}
                                                contentClassName="w-[var(--radix-popover-trigger-width)] login-select-popover"
                                                rightIcon={<ChevronDownIcon className="w-4 h-4"/>}
                                            />
                                        </div>
                                    )}
                                    {!disableCredentialForm && fields}
                                </>
                            )}
                        </div>
                    </div>
                    {
                        (showAdvanced && hasAdvancedSection && !databaseType.customFormRenderer) &&
                        <div className={classNames("transition-all h-full overflow-hidden flex flex-col gap-lg", {
                            "w-[350px] ml-4": advancedDirection === "horizontal",
                            "w-full": advancedDirection === "vertical",
                        })}>
                            {declaredAdvancedFields.map(field => {
                                const value = advancedForm[field.Key] ?? field.DefaultValue ?? "";
                                return (
                                <div className="flex flex-col gap-sm" key={field.Key}>
                                    <Label htmlFor={`${field.Key}-input`}>{t(field.LabelKey)}</Label>
                                    <Input
                                        id={`${field.Key}-input`}
                                        value={value}
                                        onChange={e => handleAdvancedForm(field.Key, e.target.value)}
                                        data-testid={`${field.Key}-input`}
                                        type={field.Kind === SourceConnectionFieldKind.Password ? "password" : "text"}
                                        placeholder={field.PlaceholderKey ? t(field.PlaceholderKey) : undefined}
                                        aria-required={field.Required ? "true" : undefined}
                                        showPasswordToggle={field.Kind === SourceConnectionFieldKind.Password && !isEmbedded}
                                    />
                                </div>
                                );
                            })}
                            {fallbackAdvancedEntries.map(([key, value]) => (
                                <div className="flex flex-col gap-sm" key={key}>
                                    <Label htmlFor={`${key}-input`}>{t(`advancedFields.${camelCase(key)}`)}</Label>
                                    <Input
                                        id={`${key}-input`}
                                        value={value}
                                        onChange={e => handleAdvancedForm(key, e.target.value)}
                                        data-testid={`${key}-input`}
                                    />
                                </div>
                            ))}
                            <SSLConfig
                                supportsCustomCAContent={databaseType.traits?.connection.supportsCustomCAContent ?? true}
                                sslModes={databaseType.sslModes}
                                advancedForm={advancedForm}
                                onAdvancedFormChange={handleAdvancedForm}
                            />
                        </div>
                    }
                </div>
                <div className={classNames("flex login-action-buttons", {
                    "justify-end": advancedForm == null,
                    "justify-between": advancedForm != null,
                })}>
                    {!disableCredentialForm && <>
                    <Button className={classNames({
                        "hidden": !hasAdvancedSection || usesFileTransport(databaseType) || databaseType.customFormRenderer != null,
                    })} onClick={handleAdvancedToggle} data-testid="advanced-button" variant="secondary">
                        <AdjustmentsHorizontalIcon className="w-4 h-4" /> {showAdvanced ? t('lessAdvancedButton') : t('advancedButton')}
                    </Button>
                    {advancedDirection === "horizontal" && (
                        <Button onClick={handleSubmit} data-testid="login-button" variant={loginWithCredentialsEnabled ? "default" : "secondary"} disabled={!loginWithCredentialsEnabled}>
                            <CheckCircleIcon className="w-4 h-4" /> {t('title')}
                        </Button>
                    )}
                    </>}
                </div>
                {advancedDirection === "vertical" && (
                    <div className={cn("flex flex-col justify-end", {
                        "grow": availableProfiles.length === 0,
                    })}>
                        {!disableCredentialForm && <>
                        <Button onClick={handleSubmit} data-testid="login-button" variant={loginWithCredentialsEnabled ? "default" : "secondary"} disabled={!loginWithCredentialsEnabled}>
                            <CheckCircleIcon className="w-4 h-4" /> {t('title')}
                        </Button>
                        </>}
                    </div>
                )}
                {
                    availableProfiles.length > 0 &&
                    <>
                        {!disableCredentialForm && <Separator className="my-8" />}
                        <div className="flex flex-col gap-lg">
                            <Label>{t('availableProfiles')}</Label>
                            <SearchSelect
                                value={selectedAvailableProfile}
                                onChange={handleAvailableProfileChange}
                                placeholder={t('selectProfile')}
                                contentClassName="w-[var(--radix-popover-trigger-width)]"
                                options={availableProfiles}
                                buttonProps={{
                                    "data-testid": "available-profiles-select",
                                }}
                                rightIcon={<ChevronDownIcon className="w-4 h-4"/>}
                            />
                            <Button onClick={() => handleLoginWithSourceProfileSubmit()} data-testid="login-with-profile-button" variant={loginWithSourceProfileEnabled ? "default" : "secondary"} disabled={!loginWithSourceProfileEnabled}>
                                <CheckCircleIcon className="w-4 h-4" /> {t('title')}
                            </Button>
                        </div>
                    </>
                }
                {cloudProvidersEnabled && (
                    <>
                        <Separator className="my-8" />
                        <AwsConnectionPicker
                            onSelectConnection={handleCloudConnectionPrefill}
                            sourceTypes={databaseTypeItems}
                        />
                        <Separator className="my-8" />
                        <AzureConnectionPicker
                            onSelectConnection={handleCloudConnectionPrefill}
                            sourceTypes={databaseTypeItems}
                        />
                        <Separator className="my-8" />
                        <GcpConnectionPicker
                            onSelectConnection={handleCloudConnectionPrefill}
                            sourceTypes={databaseTypeItems}
                        />
                    </>
                )}
            </div>
            {
                showSidePanel && advancedDirection === "horizontal" && (
                    <Card className="flex flex-col gap-6 p-8 w-[380px] shadow-xl" data-testid="sample-database-panel" aria-labelledby="sample-db-heading">
                        <div className="flex flex-col gap-4">
                            <div className="flex items-center gap-3">
                                <div className="h-14 w-14 rounded-2xl flex justify-center items-center bg-gradient-to-br from-brand to-brand/80 shadow-lg" aria-hidden="true">
                                    <SparklesIcon className="w-7 h-7 text-brand-foreground" />
                                </div>
                                <div className="flex flex-col gap-1">
                                    <h2 id="sample-db-heading" className="text-2xl font-bold text-foreground">
                                        {t('tryWhodb', { appName })}
                                    </h2>
                                    <Badge variant="secondary" className="w-fit">
                                        {t('noSetupRequired')}
                                    </Badge>
                                </div>
                            </div>

                            <p className="text-base text-muted-foreground leading-relaxed">
                                {t('experienceDescription', { appName })}
                            </p>
                        </div>

                        <Separator />

                        <div className="flex flex-col gap-3" role="list" aria-label={t('whatsIncluded')}>
                            <h3 className="text-sm font-semibold text-foreground uppercase tracking-wide">
                                {t('whatsIncluded')}
                            </h3>
                            <div className="flex flex-col gap-3">
                                <div className="flex items-start gap-3" role="listitem">
                                    <div className="h-6 w-6 rounded-lg flex justify-center items-center bg-brand/10 mt-0.5" aria-hidden="true">
                                        <ChatBubbleLeftRightIcon className="w-3.5 h-3.5 stroke-brand" />
                                    </div>
                                    <div className="flex flex-col gap-1">
                                        <p className="text-sm font-medium text-foreground">{t('aiChatAssistant')}</p>
                                        <p className="text-xs text-muted-foreground">{t('aiChatDescription')}</p>
                                    </div>
                                </div>
                                <div className="flex items-start gap-3" role="listitem">
                                    <div className="h-6 w-6 rounded-lg flex justify-center items-center bg-brand/10 mt-0.5" aria-hidden="true">
                                        <ShareIcon className="w-3.5 h-3.5 stroke-brand" />
                                    </div>
                                    <div className="flex flex-col gap-1">
                                        <p className="text-sm font-medium text-foreground">{t('visualSchema')}</p>
                                        <p className="text-xs text-muted-foreground">{t('visualSchemaDescription')}</p>
                                    </div>
                                </div>
                                <div className="flex items-start gap-3" role="listitem">
                                    <div className="h-6 w-6 rounded-lg flex justify-center items-center bg-brand/10 mt-0.5" aria-hidden="true">
                                        <TableCellsIcon className="w-3.5 h-3.5 stroke-brand" />
                                    </div>
                                    <div className="flex flex-col gap-1">
                                        <p className="text-sm font-medium text-foreground">{t('dataGrid')}</p>
                                        <p className="text-xs text-muted-foreground">{t('dataGridDescription')}</p>
                                    </div>
                                </div>
                                <div className="flex items-start gap-3" role="listitem">
                                    <div className="h-6 w-6 rounded-lg flex justify-center items-center bg-brand/10 mt-0.5" aria-hidden="true">
                                        <CodeBracketIcon className="w-3.5 h-3.5 stroke-brand" />
                                    </div>
                                    <div className="flex flex-col gap-1">
                                        <p className="text-sm font-medium text-foreground">{t('sqlEditor')}</p>
                                        <p className="text-xs text-muted-foreground">{t('sqlEditorDescription')}</p>
                                    </div>
                                </div>
                            </div>
                        </div>

                        <Button
                            onClick={handleSampleDatabaseLogin}
                            data-testid="get-started-sample-db"
                            size="lg"
                            className="w-full mt-2"
                        >
                            <SparklesIcon className="w-4 h-4" aria-hidden="true" />
                            {t('getStarted')}
                        </Button>

                        <p className="text-xs text-center text-muted-foreground">
                            {t('quickStartFooter')}
                        </p>
                    </Card>
                )
            }
        {(() => {
            const DriverInstallDialog = getComponent('driver-install-dialog') as React.LazyExoticComponent<FC<{driverName: string; onInstalled: () => void; onCancel: () => void}>> | undefined;
            if (!DriverInstallDialog || !missingDriver) return null;
            return (
                <Suspense fallback={null}>
                    <DriverInstallDialog
                        driverName={missingDriver}
                        onInstalled={() => {
                            setMissingDriver(null);
                            handleSubmit();
                        }}
                        onCancel={() => setMissingDriver(null)}
                    />
                </Suspense>
            );
        })()}
        </div>
    );
};

export const LoginPage: FC = () => {
    const { t } = useTranslation('pages/login');

    return (
        <Container className="justify-center items-center">
            <LoginForm />
            <div className="fixed bottom-4 left-1/2 -translate-x-1/2 text-xs text-foreground/60" data-testid="login-page-version">
                {t('version')}: {__APP_VERSION__}
            </div>
        </Container>
    );
};