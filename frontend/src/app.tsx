import {useMutation} from "@apollo/client/react";
import {Toaster} from "@clidey/ux";
import {UpdateSettingsDocument} from '@graphql';
import {Suspense, useCallback, useEffect} from "react";
import {Route, Routes} from "react-router-dom";
import {getStoredConsentState, optInUser, optOutUser, resetAnalyticsIdentity} from "./config/posthog";
import {getRoutes, PrivateRoute, PublicRoutes} from './config/routes';
import {getRegisteredPublicRoutes} from './config/route-registry';
import {NavigateToDefault} from "./pages/chat/default-chat-route";
import {useAppDispatch, useAppSelector} from "./store/hooks";
import {SettingsActions} from "./store/settings";
import {useThemeCustomization} from "./hooks/use-theme-customization";
import {useDesktopMenu} from "./hooks/useDesktop";
import {useSidebarShortcuts} from "./hooks/useSidebarShortcuts";
import {TourProvider} from "./components/tour/tour-provider";
import {useKeyboardShortcutsHelp} from "./components/keyboard-shortcuts-help";
import {useCommandPalette} from "./components/command-palette";
import {healthCheckService} from "./services/health-check";
import {ServerDownOverlay, DatabaseDownOverlay} from "./components/health/health-overlays";
import {HealthActions} from "./store/health";

export const App = () => {
    const [updateSettings] = useMutation(UpdateSettingsDocument);
    const dispatch = useAppDispatch();
  const metricsEnabled = useAppSelector(state => state.settings.metricsEnabled);

  useThemeCustomization();

  useDesktopMenu();

  const { KeyboardShortcutsHelpModal } = useKeyboardShortcutsHelp();

  const { CommandPaletteModal } = useCommandPalette();

  useSidebarShortcuts();

  useEffect(() => {
      const consent = getStoredConsentState();

      if (consent === 'denied' && metricsEnabled) {
          dispatch(SettingsActions.setMetricsEnabled(false));
          return;
      }

      if (consent === 'granted' && !metricsEnabled) {
          dispatch(SettingsActions.setMetricsEnabled(true));
          return;
      }

      if (consent === 'unknown') {
          return;
      }

      if (!metricsEnabled) {
          optOutUser();
          return;
      }

      optInUser();
  }, [metricsEnabled, dispatch]);

    useEffect(() => {
        const consent = getStoredConsentState();
        if (consent !== 'granted') {
            if (!metricsEnabled || consent === 'denied') {
                resetAnalyticsIdentity().catch(() => undefined);
            }
            return;
        }

        if (!metricsEnabled) {
            resetAnalyticsIdentity().catch(() => undefined);
        }
    }, [metricsEnabled]);

  const updateBackendWithSettings = useCallback(() => {
    updateSettings({
      variables: {
        newSettings: {
          MetricsEnabled: String(metricsEnabled)
        }
      }
    });
  }, [updateSettings, metricsEnabled])

  useEffect(() => {
    updateBackendWithSettings();
  }, [updateBackendWithSettings]);

  const authStatus = useAppSelector(state => state.auth.status);

  useEffect(() => {
    if (authStatus === 'logged-in') {
      healthCheckService.start();
    } else {
      healthCheckService.stop();

      dispatch(HealthActions.resetHealth());
    }

    return () => {
      healthCheckService.stop();
    };
  }, [authStatus, dispatch]);

  return (
    <TourProvider>
      <div className="h-[100vh] w-[100vw]" id="whodb-app-container">
        <Toaster />
        {KeyboardShortcutsHelpModal}
        {CommandPaletteModal}
        <ServerDownOverlay />
        <DatabaseDownOverlay />
        <Routes>
          <Route path="/" element={<PrivateRoute />}>
            {getRoutes().map(route => (
              <Route key={route.path} path={route.path} element={route.component} />
            ))}
            <Route path="/" element={<NavigateToDefault />} />
          </Route>
          <Route path={PublicRoutes.Login.path} element={PublicRoutes.Login.component} />
          {getRegisteredPublicRoutes().map(route => (
            <Route key={route.path} path={route.path} element={<Suspense fallback={null}>{<route.lazyComponent />}</Suspense>} />
          ))}
        </Routes>
      </div>
    </TourProvider>
  );
}