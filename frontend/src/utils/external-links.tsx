import React from 'react';

export const isDesktopApp = (): boolean => {

  if (typeof window === 'undefined') {
    return false;
  }

    const wailsGo = (window as any).go;
    return !!(wailsGo?.main?.App || wailsGo?.common?.App);
};

export const openExternalLink = async (url: string, event?: React.MouseEvent): Promise<void> => {

  if (event) {
    event.preventDefault();
    event.stopPropagation();
  }

  if (isDesktopApp()) {

    const wailsGo = (window as any).go;

      const app = wailsGo?.main?.App || wailsGo?.common?.App;
      if (app && app.OpenURL) {
      try {
          await app.OpenURL(url);
      } catch (error) {
        console.error('Failed to open external link:', error);

        const newWindow = window.open(url, '_blank');
        if (!newWindow) {
          console.warn('Unable to open external link in desktop app:', url);
          alert(`Please open this URL in your browser: ${url}`);
        }
      }
    } else {

      const newWindow = window.open(url, '_blank');
      if (!newWindow) {
        console.warn('Unable to open external link in desktop app:', url);
        alert(`Please open this URL in your browser: ${url}`);
      }
    }
  } else {

    window.open(url, '_blank', 'noopener,noreferrer');
  }
};

interface ExternalLinkProps extends React.AnchorHTMLAttributes<HTMLAnchorElement> {
  href: string;
  children: React.ReactNode;
}

export const ExternalLink: React.FC<ExternalLinkProps> = ({
  href,
  children,
  onClick,
  ...props
}) => {
  const handleClick = (e: React.MouseEvent<HTMLAnchorElement>) => {
    openExternalLink(href, e);
    if (onClick) {
      onClick(e);
    }
  };

  return (
    <a
      href={href}
      target="_blank"
      rel="noopener noreferrer"
      onClick={handleClick}
      {...props}
    >
      {children}
    </a>
  );
};