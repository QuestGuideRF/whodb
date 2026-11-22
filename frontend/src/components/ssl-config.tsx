import { Alert, AlertDescription, Button, Input, Label, TextArea, cn } from '@clidey/ux';
import { FC, useCallback, useEffect, useMemo, useRef, useState } from 'react';
import { SearchSelect } from './ux';
import { DocumentTextIcon, ExclamationCircleIcon, FolderIcon } from './heroicons';
import { useTranslation } from '@/hooks/use-translation';
import { SSLModeOption } from '@/config/source-types';

export const SSL_KEYS = {
  MODE: 'SSL Mode',
  CA_CONTENT: 'SSL CA Content',
  CLIENT_CERT_CONTENT: 'SSL Client Cert Content',
  CLIENT_KEY_CONTENT: 'SSL Client Key Content',
  SERVER_NAME: 'SSL Server Name',
} as const;

const MODES_REQUIRING_CA = ['verify-ca', 'verify-identity', 'enabled'];

const MODES_SUPPORTING_CLIENT_CERT = ['verify-ca', 'verify-identity', 'enabled'];

export interface SSLConfigProps {

  supportsCustomCAContent: boolean;

  sslModes?: SSLModeOption[];

  advancedForm: Record<string, string>;

  onAdvancedFormChange: (key: string, value: string) => void;

  className?: string;
}

function isInsecureConnection(): boolean {
  return typeof window !== 'undefined' && window.location.protocol === 'http:' && window.location.hostname !== 'localhost' && window.location.hostname !== '127.0.0.1';
}

export const SSLConfig: FC<SSLConfigProps> = ({
  supportsCustomCAContent,
  sslModes,
  advancedForm,
  onAdvancedFormChange,
  className,
}) => {
  const { t } = useTranslation('components/ssl-config');
  const [inputModes, setInputModes] = useState<Record<string, 'file' | 'paste'>>({
    ca: 'file',
    clientCert: 'file',
    clientKey: 'file',
  });
  const [showHttpWarning, setShowHttpWarning] = useState(false);

  useEffect(() => {
    setShowHttpWarning(isInsecureConnection());
  }, []);

  const rawMode = advancedForm[SSL_KEYS.MODE] || 'disabled';

  const currentMode = useMemo(() => {
    if (!sslModes) return rawMode;

    const exactMatch = sslModes.find((m) => m.value === rawMode);
    if (exactMatch) return rawMode;

    const aliasMatch = sslModes.find((m) => m.aliases?.includes(rawMode));
    if (aliasMatch) return aliasMatch.value;

    return rawMode;
  }, [rawMode, sslModes]);

  const requiresCA = useMemo(
    () => MODES_REQUIRING_CA.includes(currentMode),
    [currentMode]
  );
  const supportsClientCert = useMemo(
    () => MODES_SUPPORTING_CLIENT_CERT.includes(currentMode),
    [currentMode]
  );

  const handleModeChange = useCallback(
    (mode: string) => {
      onAdvancedFormChange(SSL_KEYS.MODE, mode);

      if (mode === 'disabled') {
        onAdvancedFormChange(SSL_KEYS.CA_CONTENT, '');
        onAdvancedFormChange(SSL_KEYS.CLIENT_CERT_CONTENT, '');
        onAdvancedFormChange(SSL_KEYS.CLIENT_KEY_CONTENT, '');
        onAdvancedFormChange(SSL_KEYS.SERVER_NAME, '');
      }
    },
    [onAdvancedFormChange]
  );

  const toggleInputMode = useCallback((field: 'ca' | 'clientCert' | 'clientKey') => {
    setInputModes(prev => ({
      ...prev,
      [field]: prev[field] === 'file' ? 'paste' : 'file',
    }));
  }, []);

  const modeOptions = useMemo(() => {
    if (!sslModes) return [];
    return sslModes.map((mode) => ({
      value: mode.value,
      label: t(`modes.${mode.value}.label`),
    }));
  }, [sslModes, t]);

  const currentModeInfo = useMemo(
    () => sslModes?.find((m) => m.value === currentMode),
    [sslModes, currentMode]
  );

  if (!sslModes || sslModes.length === 0) {
    return null;
  }

  return (
    <div className={cn('flex flex-col gap-md', className)}>
      { }
      {showHttpWarning && currentMode !== 'disabled' && (
        <Alert variant="destructive">
          <ExclamationCircleIcon className="h-4 w-4" />
          <AlertDescription>
            {t('httpWarning')}
          </AlertDescription>
        </Alert>
      )}

      { }
      <div className="flex flex-col gap-sm">
        <Label htmlFor="ssl-mode-select">{t('sslMode')}</Label>
        <SearchSelect
          options={modeOptions}
          value={currentMode}
          onChange={handleModeChange}
          placeholder={t('selectMode')}
          buttonProps={{
            "data-testid": "ssl-mode-select",
          }}
        />
        {currentModeInfo && (
          <span className="text-xs text-muted-foreground">
            {t(`modes.${currentModeInfo.value}.description`)}
          </span>
        )}
      </div>

      { }
      {requiresCA && !supportsCustomCAContent && (
        <div className="text-sm text-muted-foreground bg-muted/50 rounded-md p-3">
          {t('systemCaOnly')}
        </div>
      )}

      {/* CA Certificate Input */}
      {requiresCA && supportsCustomCAContent && (
        <CertificateInput
          label={t('caCertificate')}
          contentValue={advancedForm[SSL_KEYS.CA_CONTENT] || ''}
          inputMode={inputModes.ca}
          onToggleMode={() => toggleInputMode('ca')}
          onContentChange={(value) => onAdvancedFormChange(SSL_KEYS.CA_CONTENT, value)}
          testIdPrefix="ssl-ca-certificate"
        />
      )}

      { }
      {supportsClientCert && supportsCustomCAContent && (
        <>
          <CertificateInput
            label={t('clientCertificate')}
            contentValue={advancedForm[SSL_KEYS.CLIENT_CERT_CONTENT] || ''}
            inputMode={inputModes.clientCert}
            onToggleMode={() => toggleInputMode('clientCert')}
            onContentChange={(value) => onAdvancedFormChange(SSL_KEYS.CLIENT_CERT_CONTENT, value)}
            testIdPrefix="ssl-client-certificate"
            optional
          />
          <CertificateInput
            label={t('clientKey')}
            contentValue={advancedForm[SSL_KEYS.CLIENT_KEY_CONTENT] || ''}
            inputMode={inputModes.clientKey}
            onToggleMode={() => toggleInputMode('clientKey')}
            onContentChange={(value) => onAdvancedFormChange(SSL_KEYS.CLIENT_KEY_CONTENT, value)}
            testIdPrefix="ssl-client-private-key"
            optional
            isPrivateKey
            showHttpWarning={showHttpWarning}
          />
        </>
      )}

      {/* Server Name Override (for verify-identity) */}
      {currentMode === 'verify-identity' && (
        <div className="flex flex-col gap-sm">
          <Label htmlFor="ssl-server-name">
            {t('serverName')}
            <span className="text-xs text-muted-foreground ml-1">({t('optional')})</span>
          </Label>
          <Input
            id="ssl-server-name"
            value={advancedForm[SSL_KEYS.SERVER_NAME] || ''}
            onChange={(e) => onAdvancedFormChange(SSL_KEYS.SERVER_NAME, e.target.value)}
            placeholder={t('serverNamePlaceholder')}
            data-testid="ssl-server-name-input"
          />
        </div>
      )}
    </div>
  );
};

interface CertificateInputProps {
  label: string;
  contentValue: string;
  inputMode: 'file' | 'paste';
  onToggleMode: () => void;
  onContentChange: (value: string) => void;
  testIdPrefix: string;
  optional?: boolean;
  isPrivateKey?: boolean;
  showHttpWarning?: boolean;
}

const CertificateInput: FC<CertificateInputProps> = ({
  label,
  contentValue,
  inputMode,
  onToggleMode,
  onContentChange,
  testIdPrefix,
  optional,
  isPrivateKey,
  showHttpWarning,
}) => {
  const { t } = useTranslation('components/ssl-config');
  const fileInputRef = useRef<HTMLInputElement>(null);
  const [fileName, setFileName] = useState<string>('');
  const [fileError, setFileError] = useState<string>('');

  const handleFileSelect = useCallback((event: React.ChangeEvent<HTMLInputElement>) => {
    const file = event.target.files?.[0];
    if (!file) return;

    setFileError('');
    setFileName(file.name);

    const reader = new FileReader();
    reader.onload = (e) => {
      const content = e.target?.result as string;
      if (content) {
        onContentChange(content);
      }
    };
    reader.onerror = () => {
      setFileError(t('fileReadError'));
      setFileName('');
    };
    reader.readAsText(file);
  }, [onContentChange, t]);

  const handleChooseFile = useCallback(() => {
    fileInputRef.current?.click();
  }, []);

  const handleClearFile = useCallback(() => {
    setFileName('');
    onContentChange('');
    if (fileInputRef.current) {
      fileInputRef.current.value = '';
    }
  }, [onContentChange]);

  return (
    <div className="flex flex-col gap-sm">
      <div className="flex items-center justify-between">
        <Label>
          {label}
          {optional && (
            <span className="text-xs text-muted-foreground ml-1">({t('optional')})</span>
          )}
        </Label>
        <Button
          variant="ghost"
          size="sm"
          onClick={onToggleMode}
          className="h-6 px-2 text-xs"
          title={inputMode === 'file' ? t('switchToPaste') : t('switchToFile')}
        >
          {inputMode === 'file' ? (
            <>
              <DocumentTextIcon className="w-3 h-3 mr-1" />
              {t('pasteContent')}
            </>
          ) : (
            <>
              <FolderIcon className="w-3 h-3 mr-1" />
              {t('chooseFile')}
            </>
          )}
        </Button>
      </div>

      {/* Private key HTTP warning */}
      {isPrivateKey && showHttpWarning && inputMode === 'file' && (
        <span className="text-xs text-destructive">
          {t('privateKeyHttpWarning')}
        </span>
      )}

      {inputMode === 'file' ? (
        <div className="flex flex-col gap-sm">
          { }
          <input
            ref={fileInputRef}
            type="file"
            accept=".pem,.crt,.cer,.key"
            onChange={handleFileSelect}
            className="hidden"
            data-testid={`${testIdPrefix}-file-input`}
          />

          { }
          <div className="flex items-center gap-sm">
            <Button
              type="button"
              variant="outline"
              size="sm"
              onClick={handleChooseFile}
              className="flex-shrink-0"
              data-testid={`${testIdPrefix}-choose-file`}
            >
              <FolderIcon className="w-4 h-4 mr-2" />
              {t('chooseFile')}
            </Button>
            {fileName && (
              <div className="flex items-center gap-sm flex-1 min-w-0">
                <span className="text-sm text-muted-foreground truncate" title={fileName}>
                  {fileName}
                </span>
                <Button
                  type="button"
                  variant="ghost"
                  size="sm"
                  onClick={handleClearFile}
                  className="h-6 px-2 text-xs flex-shrink-0"
                >
                  {t('clear')}
                </Button>
              </div>
            )}
            {!fileName && contentValue && (
              <span className="text-sm text-muted-foreground">
                {t('contentLoaded')}
              </span>
            )}
          </div>

          {fileError && (
            <span className="text-xs text-destructive">{fileError}</span>
          )}
        </div>
      ) : (
        <TextArea
          value={contentValue}
          onChange={(e: React.ChangeEvent<HTMLTextAreaElement>) => onContentChange(e.target.value)}
          placeholder={"-----BEGIN CERTIFICATE-----\n...\n-----END CERTIFICATE-----"}
          className="font-mono text-xs h-24"
          data-testid={`${testIdPrefix}-content`}
        />
      )}
    </div>
  );
};