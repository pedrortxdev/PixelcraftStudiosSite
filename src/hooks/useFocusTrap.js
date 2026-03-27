import { useEffect, useRef } from 'react';

export function useFocusTrap(isActive = true) {
  const ref = useRef(null);

  useEffect(() => {
    if (!isActive || !ref.current) return;

    const focusableElementsString =
      'a[href], area[href], input:not([disabled]), select:not([disabled]), textarea:not([disabled]), button:not([disabled]), iframe, object, embed, [tabindex="0"], [contenteditable]';
    
    let focusableElements = Array.from(ref.current.querySelectorAll(focusableElementsString));
    
    focusableElements = focusableElements.filter((el) => {
        return el.offsetWidth > 0 || el.offsetHeight > 0 || el.getClientRects().length > 0;
    });

    if (focusableElements.length === 0) return;

    const firstElement = focusableElements[0];
    const lastElement = focusableElements[focusableElements.length - 1];

    const handleKeyDown = (e) => {
      if (e.key !== 'Tab') return;

      if (e.shiftKey) {
        if (document.activeElement === firstElement) {
          lastElement.focus();
          e.preventDefault();
        }
      } else {
        if (document.activeElement === lastElement) {
          firstElement.focus();
          e.preventDefault();
        }
      }
    };

    ref.current.addEventListener('keydown', handleKeyDown);

    // initial focus
    setTimeout(() => {
        if (ref.current && !ref.current.contains(document.activeElement)) {
            firstElement.focus();
        }
    }, 100);

    return () => {
      if (ref.current) {
        ref.current.removeEventListener('keydown', handleKeyDown);
      }
    };
  }, [isActive]);

  return ref;
}
