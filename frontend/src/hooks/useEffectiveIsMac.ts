import { useAppSelector } from '../store/hooks';
import { isMacPlatform } from '../utils/platform';

export function useEffectiveIsMac(): boolean {
    const os = useAppSelector(state => state.settings.os);
    if (os === undefined) return isMacPlatform;
    return os === 'macos';
}