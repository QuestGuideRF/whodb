import {useCallback, useEffect} from "react";
import {useNavigate} from "react-router-dom";
import {InternalRoutes} from "../config/routes";
import {useSourceContract} from "./useSourceContract";
import {useAppSelector} from "../store/hooks";
import {matchesShortcut, resolveShortcut, SHORTCUTS} from "../utils/shortcuts";

export const useSidebarShortcuts = () => {
    const navigate = useNavigate();
    const current = useAppSelector(state => state.auth.current);
    const isLoggedIn = useAppSelector(state => state.auth.status === "logged-in");
    const { supportsChat, supportsGraph, supportsScratchpad } = useSourceContract(current?.Type);

    const handleKeyDown = useCallback((event: KeyboardEvent) => {

        if (!isLoggedIn || !current) return;

        if (
            event.target instanceof HTMLInputElement ||
            event.target instanceof HTMLTextAreaElement ||
            (event.target as HTMLElement)?.isContentEditable
        ) {
            return;
        }

        const routes: string[] = [];

        if (supportsChat) {
            routes.push(InternalRoutes.Chat.path);
        }

        routes.push(InternalRoutes.Dashboard.StorageUnit.path);

        if (supportsGraph) {
            routes.push(InternalRoutes.Graph.path);
        }

        if (supportsScratchpad) {
            routes.push(InternalRoutes.RawExecute.path);
        }

        const navDefs = [SHORTCUTS.navFirst, SHORTCUTS.navSecond, SHORTCUTS.navThird, SHORTCUTS.navFourth];
        for (let i = 0; i < navDefs.length; i++) {
            if (matchesShortcut(event, resolveShortcut(navDefs[i]))) {
                if (routes[i]) {
                    event.preventDefault();
                    navigate(routes[i]);
                }
                return;
            }
        }

        if (matchesShortcut(event, SHORTCUTS.toggleSidebar)) {
            event.preventDefault();
            window.dispatchEvent(new CustomEvent('menu:toggle-sidebar'));
        }
    }, [current, isLoggedIn, navigate, supportsChat, supportsGraph, supportsScratchpad]);

    useEffect(() => {
        window.addEventListener('keydown', handleKeyDown);
        return () => window.removeEventListener('keydown', handleKeyDown);
    }, [handleKeyDown]);
};