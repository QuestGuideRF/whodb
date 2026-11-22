import {useCallback, useEffect} from 'react';
import {useLocation, useNavigate} from 'react-router-dom';
import * as desktopService from '../services/desktop';
import {isDesktopApp} from '../utils/external-links';
import {useAppSelector} from '../store/hooks';
import {InternalRoutes} from '../config/routes';

export const useDesktopFile = () => {
  const isDesktop = isDesktopApp();

  const saveFile = useCallback(async (data: string, defaultName: string) => {
    return await desktopService.saveFile(data, defaultName);
  }, []);

  const saveBinaryFile = useCallback(async (data: Uint8Array, defaultName: string) => {
    return await desktopService.saveBinaryFile(data, defaultName);
  }, []);

  const selectDirectory = useCallback(async () => {
    return await desktopService.selectDirectory();
  }, []);

  const selectDatabaseFile = useCallback(async (dbType: string) => {
    return await desktopService.selectDatabaseFile(dbType);
  }, []);

  return {
    isDesktop,
    saveFile,
    saveBinaryFile,
    selectDirectory,
    selectDatabaseFile,
  };
};

export const useDesktopWindow = () => {
  const isDesktop = isDesktopApp();

  const minimizeWindow = useCallback(async () => {
    await desktopService.minimizeWindow();
  }, []);

  const maximizeWindow = useCallback(async () => {
    await desktopService.maximizeWindow();
  }, []);

  return {
    isDesktop,
    minimizeWindow,
    maximizeWindow,
  };
};

export const useDesktopDialog = () => {
  const isDesktop = isDesktopApp();

  const showMessage = useCallback(async (title: string, message: string, type: 'info' | 'warning' | 'error' = 'info') => {
    return await desktopService.showMessageDialog(title, message, type);
  }, []);

  const showConfirm = useCallback(async (title: string, message: string) => {
    return await desktopService.showConfirmDialog(title, message);
  }, []);

  return {
    isDesktop,
    showMessage,
    showConfirm,
  };
};

export const useDesktopMenu = () => {
  const navigate = useNavigate();
  const location = useLocation();
  const isDesktop = isDesktopApp();
  const { showConfirm } = useDesktopDialog();
  const currentAuth = useAppSelector(state => state.auth.current);

  useEffect(() => {
    if (!isDesktop) return;

    const handleKeydown = (e: KeyboardEvent) => {

      if ((e.metaKey || e.ctrlKey) && (e.key === '+' || e.key === '-' || e.key === '=' || e.key === '0')) {
        e.preventDefault();
        return false;
      }
    };

    const handleWheel = (e: WheelEvent) => {
      if (e.ctrlKey || e.metaKey) {
        e.preventDefault();
        return false;
      }
    };

    document.addEventListener('keydown', handleKeydown, true);
    document.addEventListener('wheel', handleWheel, { passive: false });

    const safeHandler = (handler: () => void | Promise<void>) => {
      return async () => {
        try {
          await handler();
        } catch (error) {
          console.error('Menu handler error:', error);
        }
      };
    };

    const handlers = {
      'menu:toggle-sidebar-new-connection': safeHandler(() => {
        if (currentAuth) {

          window.dispatchEvent(new CustomEvent('menu:open-add-profile'));
        } else {

          navigate('/login');
        }
      }),
      'menu:export-data': safeHandler(() => {

        window.dispatchEvent(new CustomEvent('menu:trigger-export'));
      }),
      'menu:refresh': safeHandler(() => {

        window.dispatchEvent(new CustomEvent('app:refresh-data'));
      }),
      'menu:toggle-sidebar': safeHandler(() => {

        const sidebarTrigger = document.querySelector('[data-sidebar-trigger]') as HTMLButtonElement;
        if (sidebarTrigger) {
          sidebarTrigger.click();
        } else {

          window.dispatchEvent(new CustomEvent('menu:toggle-sidebar'));
        }
      }),
      'menu:execute-query': safeHandler(() => {

        window.dispatchEvent(new CustomEvent('menu:trigger-execute-query'));
      }),
      'menu:disconnect': safeHandler(async () => {
        const confirm = await showConfirm('Disconnect', 'Are you sure you want to disconnect from the current database?');
        if (confirm) {
          navigate(InternalRoutes.Logout.path);
        }
      }),
      'menu:new-scratchpad-page': safeHandler(() => {

        if (location.pathname === InternalRoutes.RawExecute.path) {

          window.dispatchEvent(new CustomEvent('menu:new-scratchpad-page'));
        }

      }),
    };

    Object.entries(handlers).forEach(([event, handler]) => {
      desktopService.onEvent(event, handler);
    });

    return () => {
      document.removeEventListener('keydown', handleKeydown, true);
      document.removeEventListener('wheel', handleWheel);
      Object.keys(handlers).forEach(event => {
        desktopService.offEvent(event);
      });
    };
  }, [isDesktop]);
};