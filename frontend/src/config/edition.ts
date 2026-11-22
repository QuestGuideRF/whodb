let currentEdition: string = 'ce';

export const getEdition = (): string => currentEdition;

export const setEdition = (edition: string) => {
    currentEdition = edition;
};

export const isEnterpriseEdition = (): boolean => currentEdition !== 'ce';