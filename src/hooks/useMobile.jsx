import { useState, useEffect } from 'react';

export function useMobile(breakpoint = 768) {
    const [isMobile, setIsMobile] = useState(false);

    useEffect(() => {
        // Check initially
        const checkMobile = () => {
            setIsMobile(window.innerWidth <= breakpoint);
        };

        checkMobile(); // Initial check

        // Optionally listen to resize (for orientation change or window resize)
        window.addEventListener('resize', checkMobile);
        return () => window.removeEventListener('resize', checkMobile);
    }, [breakpoint]);

    return isMobile;
}
