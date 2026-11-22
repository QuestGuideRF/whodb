import {withBasePath} from "../utils/base-path";

const DEFAULT_TITLE = "Clidey WhoDB";
const DEFAULT_ICON_PATH = "/images/logo.svg";

export const updateDocumentMeta = (extensions: Record<string, any>) => {
    const title = extensions.MetaTitle || DEFAULT_TITLE;
    const icon = extensions.MetaIcon || withBasePath(DEFAULT_ICON_PATH);

    document.title = title;

    const faviconLink = document.querySelector("link[rel='icon']") as HTMLLinkElement;
    if (faviconLink) {
        faviconLink.href = icon;
    }

    const appleTouchIcon = document.querySelector("link[rel='apple-touch-icon']") as HTMLLinkElement;
    if (appleTouchIcon) {
        appleTouchIcon.href = icon;
    }
};