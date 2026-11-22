import { getEffectiveIsMac } from "./platform";

type Modifier = "mod" | "shift" | "alt" | "ctrl";

export interface ShortcutDef {

    key: string;

    modifiers: Modifier[];

    displayKeys: string[];
}

export interface PlatformShortcutDef {
    mac: ShortcutDef;
    win: ShortcutDef;
}

export function resolveShortcut(def: PlatformShortcutDef): ShortcutDef {
    return getEffectiveIsMac() ? def.mac : def.win;
}

function platformNav(num: string): PlatformShortcutDef {
    return {
        mac: { key: num, modifiers: ["ctrl"], displayKeys: ["Ctrl", num] },
        win: { key: num, modifiers: ["alt"], displayKeys: ["Alt", num] },
    };
}

function mod(key: string, display?: string): ShortcutDef {
    return { key, modifiers: ["mod"], displayKeys: ["Mod", display ?? key.toUpperCase()] };
}

function modShift(key: string, display?: string): ShortcutDef {
    return { key, modifiers: ["mod", "shift"], displayKeys: ["Mod", "Shift", display ?? key.toUpperCase()] };
}

function plain(key: string, display?: string): ShortcutDef {
    return { key, modifiers: [], displayKeys: [display ?? key] };
}

function shift(key: string, display?: string): ShortcutDef {
    return { key, modifiers: ["shift"], displayKeys: ["Shift", display ?? key] };
}

export const SHORTCUTS = {

    showShortcuts:   { key: "?", modifiers: [], displayKeys: ["Shift", "?"] } as ShortcutDef,
    commandPalette:  mod("k"),
    closeDialogs:    plain("Escape"),
    toggleSidebar:   mod("b"),

    navFirst:  platformNav("1"),
    navSecond: platformNav("2"),
    navThird:  platformNav("3"),
    navFourth: platformNav("4"),

    moveDown:   plain("ArrowDown"),
    moveUp:     plain("ArrowUp"),
    moveFirst:  plain("Home"),
    moveLast:   plain("End"),
    pageDown:   plain("PageDown"),
    pageUp:     plain("PageUp"),
    nextPage:   mod("ArrowRight"),
    prevPage:   mod("ArrowLeft"),

    toggleSelect:     plain(" ", "Space"),
    extendSelectDown: shift("ArrowDown"),
    extendSelectUp:   shift("ArrowUp"),
    selectAll:        mod("a"),

    editRow:       plain("Enter"),
    deleteRow:     mod("Delete", "Delete"),
    deleteRowAlt:  mod("Backspace"),
    editRowAlt:    mod("e"),
    mockData:      modShift("g"),
    refresh:       mod("r"),
    exportData:    modShift("e"),
    importData:    mod("i"),

    executeQuery: mod("Enter"),
    clearEditor:  mod("u"),
} as const;

export function matchesShortcut(
    event: KeyboardEvent | React.KeyboardEvent,
    def: ShortcutDef,
): boolean {
    const isMac = getEffectiveIsMac();

    const keyMatches =
        event.key.toLowerCase() === def.key.toLowerCase() ||
        (def.key === "?" && event.key === "/" && event.shiftKey);
    if (!keyMatches) return false;

    const needsMod   = def.modifiers.includes("mod");
    const needsShift = def.modifiers.includes("shift");
    const needsAlt   = def.modifiers.includes("alt");
    const needsCtrl  = def.modifiers.includes("ctrl");

    const modPressed = isMac ? event.metaKey : event.ctrlKey;

    if (needsMod && !modPressed) return false;
    if (needsShift && !event.shiftKey) return false;
    if (needsAlt && !event.altKey) return false;
    if (needsCtrl && !event.ctrlKey) return false;

    if (!needsMod && modPressed) return false;
    if (!needsShift && event.shiftKey) {

        if (def.key !== "?") return false;
    }
    if (!needsAlt && event.altKey) return false;

    if (!needsCtrl && event.ctrlKey && !needsMod) return false;

    return true;
}