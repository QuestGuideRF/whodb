import {Alert, AlertDescription, AlertTitle, Button, toast} from "@clidey/ux";
import {BellAlertIcon, CheckCircleIcon, ClipboardDocumentIcon} from "./heroicons";
import {useState} from "react";
import {useTranslation} from "@/hooks/use-translation";
import {copyToClipboard} from "@/services/clipboard";

interface ErrorStateProps {
    error?: { message?: string } | string | null;
}

export const ErrorState = ({ error }: ErrorStateProps) => {
    const { t } = useTranslation('components/error-state');
    const [copied, setCopied] = useState(false);
    const message = typeof error === "string" ? error : error?.message ?? t('unknownError');

    const handleCopy = () => {
        copyToClipboard(message).then(success => {
            if (success) {
                setCopied(true);
                toast.success(t('copiedToClipboard'));
            }
        });
    };

    return (
        <Alert variant="destructive" title={t('title')} description={message} className="group relative" role="alert" data-testid="error-state">
            <BellAlertIcon className="w-4 h-4" aria-hidden="true" />
            <AlertTitle>{t('title')}</AlertTitle>
            <AlertDescription>{message}</AlertDescription>
            <Button
                variant="ghost"
                size="icon"
                className="absolute top-0 right-0 opacity-0 group-hover:opacity-100 focus:opacity-100 transition-all"
                onClick={handleCopy}
                aria-label={t('copyLabel')}
            >
                {copied ? <CheckCircleIcon className="w-4 h-4" aria-hidden="true" /> : <ClipboardDocumentIcon className="w-4 h-4" aria-hidden="true" />}
            </Button>
        </Alert>
    );
};