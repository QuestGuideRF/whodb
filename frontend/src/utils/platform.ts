interface NavigatorUAData {
    platform: string;
}

declare global {
    interface Navigator {
        userAgentData?: NavigatorUAData;
    }
}

import { reduxStore } from '../store';

function detectMacPlatform(): boolean {
    if (typeof navigator === 'undefined') return false;

    if (navigator.userAgentData) {
        return navigator.userAgentData.platform === 'macOS';
    }

    return /Macintosh|Mac OS X|iPhone|iPad|iPod/i.test(navigator.userAgent);
}

export const isMacPlatform = detectMacPlatform();

export function getEffectiveIsMac(): boolean {
    const os = reduxStore.getState().settings.os;
    if (os === undefined) return isMacPlatform;
    return os === 'macos';
}

export function isModKeyPressed(event: KeyboardEvent | React.KeyboardEvent): boolean {
    return getEffectiveIsMac() ? event.metaKey : event.ctrlKey;
}

const keyDisplayMap: Record<string, { mac: string; win: string }> = {
    Mod: { mac: "⌘", win: "Ctrl" },
    Alt: { mac: "⌥", win: "Alt" },
    Shift: { mac: "⇧", win: "⇧" },
    Delete: { mac: "⌫", win: "Del" },
    Backspace: { mac: "⌫", win: "Backspace" },
    Enter: { mac: "↵", win: "Enter" },
    Space: { mac: "Space", win: "Space" },
    Escape: { mac: "Esc", win: "Esc" },
    ArrowUp: { mac: "↑", win: "↑" },
    ArrowDown: { mac: "↓", win: "↓" },
    ArrowLeft: { mac: "←", win: "←" },
    ArrowRight: { mac: "→", win: "→" },
    Home: { mac: "Home", win: "Home" },
    End: { mac: "End", win: "End" },
    PageUp: { mac: "PgUp", win: "PgUp" },
    PageDown: { mac: "PgDn", win: "PgDn" },
};

export function getKeyDisplay(key: string): string {
    const mapping = keyDisplayMap[key];
    if (mapping) {
        return getEffectiveIsMac() ? mapping.mac : mapping.win;
    }
    return key;
}

export function formatShortcut(keys: string[]): string {
    const displayKeys = keys.map(getKeyDisplay);
    return getEffectiveIsMac() ? displayKeys.join("") : displayKeys.join("+");
}