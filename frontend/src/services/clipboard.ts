import {isDesktopApp} from '../utils/external-links';

const isDesktop = isDesktopApp();

export async function copyToClipboard(text: string): Promise<boolean> {
  if (isDesktop && window.go?.common?.App?.CopyToClipboard) {
    try {
      await window.go.common.App.CopyToClipboard(text);
      return true;
    } catch (error) {
      console.error('Desktop clipboard failed:', error);
    }
  }

  try {
    if (navigator.clipboard && window.isSecureContext) {
      await navigator.clipboard.writeText(text);
      return true;
    }
  } catch (error) {
    console.error('Failed to copy to clipboard:', error);
  }

  try {
    const textarea = document.createElement('textarea');
    textarea.value = text;
    textarea.style.position = 'fixed';
    textarea.style.left = '-9999px';
    textarea.style.opacity = '0';
    document.body.appendChild(textarea);
    textarea.select();
    const success = document.execCommand('copy');
    document.body.removeChild(textarea);
    return success;
  } catch (error) {
    console.error('execCommand copy fallback failed:', error);
  }

  return false;
}