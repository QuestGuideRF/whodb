import { FC, ReactElement, ReactNode } from "react";
import { cn, Tooltip, TooltipContent, TooltipTrigger } from "@clidey/ux";
import { CloudIcon, ShieldCheckIcon } from "../heroicons";
import { useTranslation } from "@/hooks/use-translation";

interface SSLStatus {
    IsEnabled: boolean;
    Mode: string;
}

interface DatabaseIconWithBadgeProps {

    icon: ReactElement | null;

    showCloudBadge?: boolean;

    sslStatus?: SSLStatus | null;

    className?: string;

    size?: "sm" | "md" | "lg";
}

export const DatabaseIconWithBadge: FC<DatabaseIconWithBadgeProps> = ({
    icon,
    showCloudBadge = false,
    sslStatus,
    className,
    size = "md",
}) => {
    const { t } = useTranslation("components/aws-providers-section");
    const { t: tSidebar } = useTranslation("components/sidebar");

    const sizeClasses = {
        sm: "w-4 h-4",
        md: "w-6 h-6",
        lg: "w-8 h-8",
    };

    const badgeSizeClasses = {
        sm: "w-2 h-2 -right-0.5 -bottom-0.5",
        md: "w-3 h-3 -right-1 -bottom-1",
        lg: "w-4 h-4 -right-1 -bottom-1",
    };

    const sslBadgeSizeClasses = {
        sm: "w-2 h-2 -right-0.5 -top-0.5",
        md: "w-3 h-3 -right-1 -top-1",
        lg: "w-4 h-4 -right-1 -top-1",
    };

    if (!icon) return null;

    const showSslBadge = sslStatus?.IsEnabled;

    return (
        <div className={cn("relative inline-flex", sizeClasses[size], className)}>
            {icon}
            {showCloudBadge && (
                <div
                    className={cn(
                        "absolute rounded-full bg-background border border-border flex items-center justify-center",
                        badgeSizeClasses[size]
                    )}
                    title={t('cloudConnection')}
                >
                    <CloudIcon className={cn(
                        "text-brand",
                        size === "sm" ? "w-1.5 h-1.5" : size === "md" ? "w-2 h-2" : "w-2.5 h-2.5"
                    )} />
                </div>
            )}
            {showSslBadge && (
                <Tooltip>
                    <TooltipTrigger asChild>
                        <div
                            data-testid="ssl-badge"
                            className={cn(
                                "absolute rounded-full bg-background border border-border flex items-center justify-center",
                                sslBadgeSizeClasses[size]
                            )}
                        >
                            <ShieldCheckIcon className={cn(
                                "text-green-500",
                                size === "sm" ? "w-1.5 h-1.5" : size === "md" ? "w-2 h-2" : "w-2.5 h-2.5"
                            )} />
                        </div>
                    </TooltipTrigger>
                    <TooltipContent>
                        {tSidebar('sslSecured', { mode: sslStatus?.Mode })}
                    </TooltipContent>
                </Tooltip>
            )}
        </div>
    );
};

/**
 * Helper to determine if a profile ID indicates an AWS connection.
 * AWS connections have IDs prefixed with "aws-".
 */
export function isAwsConnection(profileId: string | undefined): boolean {
    return profileId?.startsWith("aws-") ?? false;
}