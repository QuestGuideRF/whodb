import React, {useEffect, useState} from 'react';
import ReactDOM from 'react-dom/client';
import './index.css';
import {ApolloProvider} from "@apollo/client/react";
import {graphqlClient} from './config/graphql-client';
import {Provider} from "react-redux";
import {reduxStore, reduxStorePersistor} from './store';
import {App} from './app';
import {BrowserRouter, HashRouter} from "react-router-dom";
import {PersistGate} from 'redux-persist/integration/react';
import {PostHogProvider} from 'posthog-js/react';
import type {PostHog} from 'posthog-js';
import {initPosthog} from "./config/posthog";
import {ThemeProvider} from '@clidey/ux'
import {isDesktopApp} from './utils/external-links';
import {PosthogConsentBanner} from './components/analytics/posthog-consent-banner';
import {ErrorBoundary} from './components/error-boundary';
import {getBasePath} from './utils/base-path';

try {
    if (typeof navigator !== 'undefined' && typeof document !== 'undefined') {
        if (isDesktopApp() && /Linux/i.test(navigator.userAgent || '')) {
            document.documentElement.classList.add('linux');
        }
    }
} catch (e) {

}

const root = ReactDOM.createRoot(
  document.getElementById('root') as HTMLElement
);

const AppWithProviders = () => {
    const [posthogClient, setPosthogClient] = useState<PostHog | null>(null);
    const [initialized, setInitialized] = useState(false);

    useEffect(() => {
        if (initialized) {
            return;
        }
        setInitialized(true);
        initPosthog()
            .then(client => setPosthogClient(client))
            .catch(() => setPosthogClient(null));
    }, [initialized]);

    const app = (
        <ThemeProvider>
            <App />
            <PosthogConsentBanner/>
        </ThemeProvider>
    );

    if (posthogClient) {
        return <PostHogProvider client={posthogClient}>{app}</PostHogProvider>;
    }

    return app;
};

const desktopApp = isDesktopApp();
const browserBasePath = desktopApp ? undefined : (getBasePath() || undefined);

root.render(
  <React.StrictMode>
    {desktopApp ? (
      <HashRouter>
        <ApolloProvider client={graphqlClient}>
          <Provider store={reduxStore}>
            <PersistGate loading={null} persistor={reduxStorePersistor}>
              <ErrorBoundary>
                <AppWithProviders />
              </ErrorBoundary>
            </PersistGate>
          </Provider>
        </ApolloProvider>
      </HashRouter>
    ) : (
      <BrowserRouter basename={browserBasePath}>
        <ApolloProvider client={graphqlClient}>
          <Provider store={reduxStore}>
            <PersistGate loading={null} persistor={reduxStorePersistor}>
              <ErrorBoundary>
                <AppWithProviders />
              </ErrorBoundary>
            </PersistGate>
          </Provider>
        </ApolloProvider>
      </BrowserRouter>
    )}
  </React.StrictMode>
);