import React from 'react';

const registry = new Map<string, React.LazyExoticComponent<any>>();

export const registerComponent = (name: string, loader: () => Promise<{ default: any }>) => {
    registry.set(name, React.lazy(loader));
};

export const getComponent = (name: string): React.LazyExoticComponent<any> | undefined => {
    return registry.get(name);
};

export const hasComponent = (name: string): boolean => {
    return registry.has(name);
};