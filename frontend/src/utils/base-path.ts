function isDesktopContext(): boolean {
    if (typeof window === "undefined") {
        return false;
    }

    const wailsGo = (window as any).go;
    return !!(wailsGo?.main?.App || wailsGo?.common?.App);
}

function normalizeBasePath(pathname: string): string {
    if (pathname === "/" || pathname === "") {
        return "";
    }

    return pathname.replace(/\/+$/, "");
}

export function getBasePath(): string {
    if (typeof window === "undefined" || typeof document === "undefined" || isDesktopContext()) {
        return "";
    }

    return normalizeBasePath(new URL(document.baseURI || window.location.href).pathname);
}

export function withBasePath(path: string): string {
    if (!path.startsWith("/") || path.startsWith("//")) {
        return path;
    }

    const basePath = getBasePath();
    if (path === "/") {
        return basePath === "" ? "/" : `${basePath}/`;
    }

    return basePath === "" ? path : `${basePath}${path}`;
}