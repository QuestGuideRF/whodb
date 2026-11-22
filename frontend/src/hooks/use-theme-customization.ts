import { useEffect } from 'react';
import { useAppSelector } from '../store/hooks';
import { applyUICustomization } from '../utils/theme-customization';

export const useThemeCustomization = () => {
  const settings = useAppSelector(state => state.settings);

  useEffect(() => {
    applyUICustomization(settings);
  }, [settings.fontSize, settings.borderRadius, settings.spacing]);

  useEffect(() => {
    if (settings.disableAnimations) {
      document.body.classList.add('disable-animations');
    } else {
      document.body.classList.remove('disable-animations');
    }
  }, [settings.disableAnimations]);

};