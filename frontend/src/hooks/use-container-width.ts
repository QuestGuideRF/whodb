import { RefObject, useEffect, useState } from "react";

export function useContainerWidth(ref: RefObject<HTMLElement | null>): number {
    const [width, setWidth] = useState(0);

    useEffect(() => {
        const update = () => {
            if (ref.current) {
                setWidth(ref.current.offsetWidth);
            }
        };

        update();
        window.addEventListener('resize', update);
        return () => window.removeEventListener('resize', update);
    }, [ref]);

    return width;
}