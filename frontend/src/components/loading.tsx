import {FC} from "react";
import {Container} from "./page";

import {Spinner} from "@clidey/ux";
import {useTranslation} from "@/hooks/use-translation";

type ILoadingProps = {
  className?: string;
  size?: "sm" | "md" | "lg";
  hideText?: boolean;
  loadingText?: string;
}

export const Loading: FC<ILoadingProps> = ({className, size = "md", hideText = true, loadingText}) => {
  const { t } = useTranslation('components/loading');
  let textSize = "text-base";
  if (size === "sm") {
      textSize = "text-xs";
  } else if (size === "md") {
      textSize = "text-sm";
  } else if (size === "lg") {
      textSize = "text-base";
  }

    return (
        <div
            className="flex justify-center items-center w-fit h-fit gap-sm"
            data-testid="loading-spinner"
            role="status"
            aria-busy="true"
            aria-label={loadingText || t('loading')}
        >
            <Spinner className={className} size={size} aria-hidden="true" />
            {!hideText && <p className={textSize}>{loadingText || t('loading')}</p>}
        </div>
    );
};

export const LoadingPage: FC = () => {
  return <Container className="flex justify-center items-center h-full w-full">
    <Loading size="lg" />
  </Container>
}