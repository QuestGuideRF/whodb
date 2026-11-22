import {type SupportedLanguage, DEFAULT_LANGUAGE} from '@/utils/languages';

type TranslationMap = Record<string, string>;
type TranslationCache = Record<string, TranslationMap>;
type ParsedYaml = Record<string, TranslationMap>;

const translationCache: TranslationCache = {};

const ceModules = import.meta.glob<ParsedYaml>('/src/locales/**/*.yaml', { import: 'default', eager: true });

const buildIndex = (modules: Record<string, ParsedYaml>): Map<string, ParsedYaml> => {
    const index = new Map<string, ParsedYaml>();
    for (const [key, mod] of Object.entries(modules)) {
        const match = key.match(/\/locales\/(.+)\.yaml$/);
        if (match) index.set(match[1], mod);
    }
    return index;
};

const ceIndex = buildIndex(ceModules);

let extensionIndex = new Map<string, ParsedYaml>();

export const registerLocaleModules = (modules: Record<string, ParsedYaml>) => {
    extensionIndex = buildIndex(modules);
};

const pluralRulesCache = new Map<string, Intl.PluralRules>();

export const getPluralCategory = (locale: string, count: number): Intl.LDMLPluralRule => {
    let rules = pluralRulesCache.get(locale);
    if (!rules) {
        rules = new Intl.PluralRules(locale.replace('_', '-'));
        pluralRulesCache.set(locale, rules);
    }
    return rules.select(count);
};

export const loadTranslationsSync = (
    componentPath: string,
    language: SupportedLanguage
): TranslationMap => {
    const cacheKey = `${componentPath}-${language}`;

    if (translationCache[cacheKey]) {
        return translationCache[cacheKey];
    }

    try {

        const commonParsed = ceIndex.get('common');
        let translations: TranslationMap = commonParsed
            ? { ...(commonParsed[language] || commonParsed[DEFAULT_LANGUAGE]) }
            : {};

        const ceParsed = ceIndex.get(componentPath);
        if (ceParsed) {
            translations = { ...translations, ...(ceParsed[language] || ceParsed[DEFAULT_LANGUAGE]) };
        }

        const extParsed = extensionIndex.get(componentPath);
        if (extParsed) {
            const extTranslations = extParsed[language] || extParsed[DEFAULT_LANGUAGE];
            if (extTranslations) {
                translations = { ...translations, ...extTranslations };
            }
        }

        if (!ceParsed && !extParsed) {
            console.error(`Translation file not found for ${componentPath}`, {
                availableCE: [...ceIndex.keys()],
                language
            });
            return translations;
        }

        translationCache[cacheKey] = translations;
        return translations;
    } catch (error) {
        console.error(`Failed to load translations for ${componentPath}:`, error);
        return {};
    }
};

export const getTranslation = (
    translations: TranslationMap,
    key: string,
    language: SupportedLanguage,
    params?: Record<string, any>
): string => {
    let resolvedKey = key;

    if (params != null && typeof params.count === 'number') {
        const category = getPluralCategory(language, params.count);
        const pluralKey = `${key}_${category}`;
        if (translations[pluralKey]) {
            resolvedKey = pluralKey;
        }
    }

    const template = translations[resolvedKey] || resolvedKey;

    if (params != null && template.includes('{')) {
        return template.replace(/\{(\w+)\}/g, (match, paramKey) => {
            return String(params[paramKey] ?? match);
        });
    }

    return template;
};