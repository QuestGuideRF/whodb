import { type ReactNode, useEffect, useState } from 'react';
import { useAppSelector } from '@/store/hooks';
import { loadTranslationsSync, getTranslation, getPluralCategory } from '@/utils/i18n';

export const useTranslation = (componentPath: string) => {
    const language = useAppSelector(state => state.settings.language);
    const [translations, setTranslations] = useState<Record<string, string>>(() =>
        loadTranslationsSync(componentPath, language)
    );
    useEffect(() => {
        setTranslations(loadTranslationsSync(componentPath, language));
    }, [componentPath, language]);

    const t: {
        (key: string): string;
        (key: string, params: Record<string, string | number>): string;
        (key: string, params: Record<string, ReactNode>): ReactNode;
    } = (key: string, params?: Record<string, any>): any => {
        if (typeof params !== 'object' || params === null) {
            return getTranslation(translations, key, language);
        }

        const hasJsx = Object.values(params).some(
            v => v !== null && v !== undefined && typeof v === 'object'
        );

        if (!hasJsx) {
            return getTranslation(translations, key, language, params);
        }

        let resolvedKey = key;
        if (typeof params.count === 'number') {
            const category = getPluralCategory(language, params.count);
            const pluralKey = `${key}_${category}`;
            if (translations[pluralKey]) {
                resolvedKey = pluralKey;
            }
        }

        const template = translations[resolvedKey] || resolvedKey;
        if (!template.includes('{')) {
            return template;
        }

        const parts: ReactNode[] = [];
        let lastIndex = 0;
        const regex = /\{(\w+)\}/g;
        let match;

        while ((match = regex.exec(template)) !== null) {
            if (match.index > lastIndex) {
                parts.push(template.slice(lastIndex, match.index));
            }
            parts.push(params[match[1]] ?? match[0]);
            lastIndex = match.index + match[0].length;
        }

        if (lastIndex < template.length) {
            parts.push(template.slice(lastIndex));
        }

        return parts;
    };

    return { t, language };
};