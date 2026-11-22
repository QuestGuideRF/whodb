const SQL_SAFE_KEYWORDS = ['SELECT', 'WITH', 'EXPLAIN', 'DESCRIBE', 'SHOW', 'USE'];

const SQL_DESTRUCTIVE_KEYWORDS = [
    'INSERT', 'UPDATE', 'DELETE', 'CREATE', 'DROP', 'ALTER',
    'TRUNCATE', 'REPLACE', 'MERGE', 'CALL', 'EXEC', 'EXECUTE',
];

const SQL_ALL_KEYWORDS = [
    ...SQL_SAFE_KEYWORDS,
    'INSERT', 'UPDATE', 'DELETE', 'CREATE', 'DROP', 'ALTER', 'SET',
];

export const isValidSQLQuery = (text: string): boolean => {
    const trimmed = text.trim();
    if (!trimmed) return false;
    const upperText = trimmed.toUpperCase();
    return SQL_ALL_KEYWORDS.some(keyword => upperText.startsWith(keyword));
};

export const isDestructiveQuery = (text: string): boolean => {
    const trimmed = text.trim();
    if (!trimmed) return false;
    const upperText = trimmed.toUpperCase();
    if (SQL_SAFE_KEYWORDS.some(keyword => upperText.startsWith(keyword))) return false;
    if (SQL_DESTRUCTIVE_KEYWORDS.some(keyword => upperText.startsWith(keyword))) return true;

    return true;
};