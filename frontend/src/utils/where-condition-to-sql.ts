import { WhereCondition, WhereConditionType } from '@graphql';

function escapeSqlString(value: string): string {
    return value.replace(/'/g, "''");
}

export function whereConditionToSql(condition: WhereCondition | undefined): string {
    if (!condition) {
        return '';
    }

    switch (condition.Type) {
        case WhereConditionType.Atomic:
            if (!condition.Atomic) return '';
            const { Key, Operator, Value } = condition.Atomic;

            if (Value.toUpperCase() === 'NULL') {
                if (Operator === '=') return `${Key} IS NULL`;
                if (Operator === '!=') return `${Key} IS NOT NULL`;
            }

            const isNumeric = !isNaN(Number(Value)) && Value.trim() !== '';
            const formattedValue = isNumeric ? Value : `'${escapeSqlString(Value)}'`;

            return `${Key} ${Operator} ${formattedValue}`;

        case WhereConditionType.And:
            if (!condition.And?.Children || condition.And.Children.length === 0) return '';
            const andClauses = condition.And.Children
                .map(child => whereConditionToSql(child))
                .filter(clause => clause !== '');
            if (andClauses.length === 0) return '';
            if (andClauses.length === 1) return andClauses[0];
            return `(${andClauses.join(' AND ')})`;

        case WhereConditionType.Or:
            if (!condition.Or?.Children || condition.Or.Children.length === 0) return '';
            const orClauses = condition.Or.Children
                .map(child => whereConditionToSql(child))
                .filter(clause => clause !== '');
            if (orClauses.length === 0) return '';
            if (orClauses.length === 1) return orClauses[0];
            return `(${orClauses.join(' OR ')})`;

        default:
            return '';
    }
}