import { ReactElement } from "react";

export type RegisteredSidebarItem = {

    name: string;

    title: string;

    path: string;

    icon: ReactElement;

    disabled?: boolean;
};

const items: RegisteredSidebarItem[] = [];

export function registerSidebarItem(item: RegisteredSidebarItem): void {
    items.push(item);
}

export function getRegisteredSidebarItems(): RegisteredSidebarItem[] {
    return items;
}