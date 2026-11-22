import sampleSize from "lodash/sampleSize";

export function formatNumber(value: number, language: string): string {
    return new Intl.NumberFormat(language.replace('_', '-')).format(value);
}

export function isNumeric(str: string) {
    return !isNaN(Number(str));
}

export function chooseRandomItems<T>(array: T[], n: number = 3): T[] {
    if (n > array.length) {
        throw new Error("n cannot be greater than the array length");
    }
    return sampleSize(array, n);
}