interface ImportMetaEnv {
  readonly VITE_E2E_TEST: 'true' | 'false' | undefined

}

interface ImportMeta {
  readonly env: ImportMetaEnv
}

interface DoryDocsAPI {
  init(config: {
    baseUrl: string;
    position?: 'left' | 'right';
    width?: number;
    theme?: 'light' | 'dark' | 'inherit';
    closeOnEscape?: boolean;
    closeOnOverlayClick?: boolean;
    onOpen?: () => void;
    onClose?: () => void;
  }): void;
  open(path?: string): void;
  close(): void;
  toggle(): void;
}

declare const __APP_VERSION__: string;

interface Window {
  DoryDocs?: DoryDocsAPI;
}