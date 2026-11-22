import {ApolloClient, InMemoryCache} from '@apollo/client';
import {setContext} from '@apollo/client/link/context';
import {onError} from '@apollo/client/link/error';
import {HttpLink} from '@apollo/client/link/http';
import {toast} from '@clidey/ux';
import {print} from 'graphql';
import {
    LoginSourceDocument,
    LoginSourceMutationVariables,
    LoginWithSourceProfileDocument,
    LoginWithSourceProfileMutationVariables
} from '@graphql';
import {LocalLoginProfile} from '../store/auth';
import {reduxStore} from '../store';
import {addAuthHeader} from '../utils/auth-headers';
import {withBasePath} from '../utils/base-path';
import {isAwsHostname} from '../utils/cloud-connection-prefill';
import {getTranslation, loadTranslationsSync} from '../utils/i18n';
import {type SupportedLanguage, DEFAULT_LANGUAGE} from '../utils/languages';
import {clearSourceSessionMetadata} from '../utils/source-session-metadata-cache';

const uri = withBasePath("/api/query");
const loginWithSourceProfileQuery = print(LoginWithSourceProfileDocument);
const loginMutationQuery = print(LoginSourceDocument);

let onUnauthorizedHandler: (() => Promise<boolean>) | null = null;

export const registerOnUnauthorized = (fn: () => Promise<boolean>): void => {
    onUnauthorizedHandler = fn;
};

type GraphQLClientTranslationKey = 'sessionExpired' | 'autoLoginSuccess' | 'autoLoginFailed';
type TranslatorFn = (key: GraphQLClientTranslationKey) => string;

const getTranslator = (): TranslatorFn => {
    const language = (reduxStore.getState().settings.language ?? DEFAULT_LANGUAGE) as SupportedLanguage;
    const translations = loadTranslationsSync('config/graphql-client', language);
    return (key: GraphQLClientTranslationKey) => getTranslation(translations, key, language);
};

const redirectToLoginWithMessage = (
    key: GraphQLClientTranslationKey,
    translator?: TranslatorFn
) => {
    const t = translator ?? getTranslator();
    toast.error(t(key));
    window.location.href = withBasePath('/login');
};

const httpLink = new HttpLink({
  uri,
  credentials: "include",
});

const authLink = setContext((_, prevContext) => {
    return {
        headers: addAuthHeader(prevContext.headers),
    };
});

const errorLink = onError(({networkError}) => {
    if (networkError && 'statusCode' in networkError && networkError.statusCode === 401) {
        if (onUnauthorizedHandler) {
            void onUnauthorizedHandler().then(handled => {
                if (!handled) {
                    fallbackAutoLogin();
                }
            });
            return;
        }
        fallbackAutoLogin();
    } else if (networkError) {
        console.error('Network error:', networkError);
    }
});

function fallbackAutoLogin() {

    const authState = reduxStore.getState().auth;
    const currentProfile = authState.current;

    if (currentProfile) {
        void handleAutoLogin(currentProfile);
    } else {
        if (!window.location.pathname.startsWith(withBasePath('/login'))) {
            redirectToLoginWithMessage('sessionExpired');
        }
    }
}

async function handleAutoLogin(currentProfile: LocalLoginProfile) {
    const t = getTranslator();
    try {

        const cloudProvidersEnabled = reduxStore.getState().settings.cloudProvidersEnabled;
        if (isAwsHostname(currentProfile.Hostname) && !cloudProvidersEnabled) {
            return;
        }

        let response, result;
        if (currentProfile.Saved) {

            const variables: LoginWithSourceProfileMutationVariables = {
                profile: {
                    Id: currentProfile.Id,
                    Values: currentProfile.Database ? [{ Key: "Database", Value: currentProfile.Database }] : [],
                },
            };
            response = await fetch(uri, {
                method: 'POST',
                headers: addAuthHeader({
                    'Content-Type': 'application/json',
                }),
                credentials: 'include',
                body: JSON.stringify({
                    operationName: 'LoginWithSourceProfile',
                    query: loginWithSourceProfileQuery,
                    variables,
                }),
            });
            result = await response.json();
            if (result.data?.LoginWithSourceProfile?.Status) {
                toast.success(t('autoLoginSuccess'));
                window.location.reload();
                return;
            } else {
                redirectToLoginWithMessage('autoLoginFailed', t);
                return;
            }
        } else {

            const variables: LoginSourceMutationVariables = {
                credentials: {
                    Id: currentProfile.Id,
                    SourceType: currentProfile.SourceType,
                    Values: currentProfile.Values,
                    AccessToken: currentProfile.AccessToken,
                },
            };
            response = await fetch(uri, {
                method: 'POST',
                headers: addAuthHeader({
                    'Content-Type': 'application/json',
                }),
                credentials: 'include',
                body: JSON.stringify({
                    operationName: 'LoginSource',
                    query: loginMutationQuery,
                    variables,
                }),
            });
            result = await response.json();
            if (result.data?.LoginSource?.Status) {
                toast.success(t('autoLoginSuccess'));
                window.location.reload();
                return;
            } else {
                redirectToLoginWithMessage('autoLoginFailed', t);
                return;
            }
        }
    } catch (error) {
        console.error('Auto-login error:', error);
        redirectToLoginWithMessage('autoLoginFailed', t);
    }
}

export const graphqlClient = new ApolloClient({
    link: errorLink.concat(authLink.concat(httpLink)),
  cache: new InMemoryCache(),
  defaultOptions: {
      watchQuery: {
        fetchPolicy: "cache-first",
        refetchWritePolicy: "overwrite",
      },
      query: {
        fetchPolicy: "no-cache",
      },
      mutate: {
        fetchPolicy: "no-cache",
      },
  }
});

export async function clearGraphqlStore(): Promise<void> {
    await graphqlClient.clearStore();
    clearSourceSessionMetadata();
}