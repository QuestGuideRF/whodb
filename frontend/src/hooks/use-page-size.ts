import {useCallback, useState} from "react";
import {toast} from "@clidey/ux";
import {useTranslation} from "./use-translation";

export const PRESET_PAGE_SIZES = ["10", "25", "50", "100", "250", "500", "1000"];

interface UsePageSizeOptions {
    onPageSizeChange?: (size: number) => void;
    maxPageSize?: number;
}

export function usePageSize(initialSize: number, options?: UsePageSizeOptions) {
    const { t } = useTranslation('hooks/use-page-size');
    const [pageSize, setPageSize] = useState(String(initialSize));
    const [isCustom, setIsCustom] = useState(!PRESET_PAGE_SIZES.includes(String(initialSize)));
    const [customInput, setCustomInput] = useState(String(initialSize));

    const handleSelectChange = useCallback((value: string) => {
        if (value === "custom") {
            setIsCustom(true);
            setCustomInput(pageSize);
        } else {
            setIsCustom(false);
            setPageSize(value);
            options?.onPageSizeChange?.(Number.parseInt(value, 10));
        }
    }, [pageSize, options]);

    const handleCustomApply = useCallback(() => {
        const parsed = Number.parseInt(customInput, 10);
        if (Number.isNaN(parsed) || parsed <= 0) {
            toast.error(t('enterValidNumber'));
            setCustomInput(pageSize);
            return false;
        }
        const limit = options?.maxPageSize ?? 10000;
        if (parsed > limit) {
            toast.error(t('exceedsMaxPageSize', { max: limit }));
            setCustomInput(pageSize);
            return false;
        }
        setPageSize(String(parsed));
        options?.onPageSizeChange?.(parsed);
        return true;
    }, [customInput, pageSize, options, t]);

    return {
        pageSize: Number.parseInt(pageSize, 10),
        pageSizeString: pageSize,
        isCustom,
        customInput,
        setCustomInput,
        handleSelectChange,
        handleCustomApply,
    };
}