import {reduxStore} from '../store';
import {getAnalyticsDistinctId} from '../config/posthog';

let authHeaderProvider: (() => string | null) | null = null;

type SourceCredentialValueLike = {
    Key: string;
    Value: string;
};

function mapSourceCredentialValues(values: readonly SourceCredentialValueLike[] | undefined): Record<string, string> {
    return (values ?? []).reduce<Record<string, string>>((acc, value) => {
        acc[value.Key] = value.Value;
        return acc;
    }, {});
}

export const registerAuthHeaderProvider = (fn: () => string | null): void => {
    authHeaderProvider = fn;
};

const analyticsHeaderName = 'X-WhoDB-Analytics-Id';

export function isDesktopScheme(): boolean {

    if (typeof window !== 'undefined') {
        const wailsGo = (window as any).go;
        if (wailsGo?.main?.App || wailsGo?.common?.App) {
            return true;
        }
    }

  return typeof window !== 'undefined' && !['http:', 'https:'].includes(window.location.protocol);
}

export function getAuthorizationHeader(): string | null {
  if (authHeaderProvider) {
    return authHeaderProvider();
  }
  try {

    const authState = reduxStore.getState().auth;
    if (authState?.status !== 'logged-in') {
      return null;
    }

    const current = authState?.current;
    if (!current) {
      return null;
    }

    const values = mapSourceCredentialValues(current.Values);

    const tokenPayload = current.Saved ? {
      Id: current.Id,
      Values: current.Database ? { Database: current.Database } : {},
    } : {
      Id: current.Id,
      SourceType: current.SourceType,
      Values: values,
      AccessToken: current.AccessToken,
      IsProfile: false,
    };

    const jsonString = JSON.stringify(tokenPayload);

    const utf8Bytes = encodeURIComponent(jsonString).replace(/%([0-9A-F]{2})/g,
      (_, p1) => String.fromCharCode(parseInt(p1, 16)));
    const bearer = btoa(utf8Bytes);
    return `Bearer ${bearer}`;
  } catch (error) {
    console.error('Failed to get authorization header:', error);
    return null;
  }
}

export function addAuthHeader(headers: HeadersInit = {}): HeadersInit {
    const authHeader = getAuthorizationHeader();
    const id = getAnalyticsDistinctId()
    headers = {
        ...headers,
        [analyticsHeaderName]: id != null ? id : ""
    }
    if (authHeader) {
        return {
            ...headers,
            Authorization: authHeader,
        };
    }
    return headers;
}