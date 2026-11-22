import { ComponentType, lazy, LazyExoticComponent } from "react";

type RouteFactory = () => Promise<{ default: ComponentType<any> }>;

export type RegisteredRoute = {
    name: string;
    path: string;

    lazyComponent: LazyExoticComponent<ComponentType<any>>;
};

const registrations: RegisteredRoute[] = [];
const publicRegistrations: RegisteredRoute[] = [];

export function registerRoute(name: string, path: string, factory: RouteFactory): void {
    registrations.push({ name, path, lazyComponent: lazy(factory) });
}

export function registerPublicRoute(name: string, path: string, factory: RouteFactory): void {
    publicRegistrations.push({ name, path, lazyComponent: lazy(factory) });
}

export function getRegisteredRoutes(): RegisteredRoute[] {
    return registrations;
}

export function getRegisteredPublicRoutes(): RegisteredRoute[] {
    return publicRegistrations;
}