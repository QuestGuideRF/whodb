import { ReactNode, useMemo } from "react";
import { CircleStackIcon, UsersIcon, UserGroupIcon, InformationCircleIcon, PresentationChartLineIcon, ShoppingBagIcon, BuildingStorefrontIcon, BanknotesIcon, PresentationChartBarIcon } from "../../components/heroicons";
import { useTranslation } from "@/hooks/use-translation";

export interface ChatExample {
    icon: ReactNode;
    description: string;
}

const exampleIcons: ReactNode[] = [
    <CircleStackIcon className="w-4 h-4" />,
    <UsersIcon className="w-4 h-4" />,
    <UserGroupIcon className="w-4 h-4" />,
    <InformationCircleIcon className="w-4 h-4" />,
    <PresentationChartLineIcon className="w-4 h-4" />,
    <PresentationChartLineIcon className="w-4 h-4" />,
    <UserGroupIcon className="w-4 h-4" />,
    <ShoppingBagIcon className="w-4 h-4" />,
    <ShoppingBagIcon className="w-4 h-4" />,
    <BuildingStorefrontIcon className="w-4 h-4" />,
    <UserGroupIcon className="w-4 h-4" />,
    <UserGroupIcon className="w-4 h-4" />,
    <UserGroupIcon className="w-4 h-4" />,
    <BanknotesIcon className="w-4 h-4" />,
    <BanknotesIcon className="w-4 h-4" />,
    <BanknotesIcon className="w-4 h-4" />,
    <PresentationChartBarIcon className="w-4 h-4" />,
    <PresentationChartBarIcon className="w-4 h-4" />,
    <PresentationChartBarIcon className="w-4 h-4" />,
    <PresentationChartBarIcon className="w-4 h-4" />,
    <PresentationChartBarIcon className="w-4 h-4" />,
    <PresentationChartBarIcon className="w-4 h-4" />,
    <PresentationChartBarIcon className="w-4 h-4" />,
    <PresentationChartBarIcon className="w-4 h-4" />,
    <PresentationChartBarIcon className="w-4 h-4" />,
    <PresentationChartBarIcon className="w-4 h-4" />,
    <PresentationChartBarIcon className="w-4 h-4" />,
    <PresentationChartBarIcon className="w-4 h-4" />,
];

export const useChatExamples = (): ChatExample[] => {
    const { t } = useTranslation("pages/chat");

    return useMemo(() => {
        return exampleIcons.map((icon, index) => ({
            icon,
            description: t(`example${index}`),
        }));
    }, [t]);
};

export const chatExamples: ChatExample[] = exampleIcons.map((icon, index) => ({
    icon,
    description: `example${index}`,
}));