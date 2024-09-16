import React from 'react';
import * as Progress from '@radix-ui/react-progress';

export function Spinner() {
    return (
        <Progress.Root
            className="relative overflow-hidden bg-gray-200 rounded-full w-8 h-8"
            style={{
                transform: 'translateZ(0)',
            }}
        >
            <Progress.Indicator
                className="w-full h-full bg-blue-500 rounded-full transition-transform duration-660 ease-[cubic-bezier(0.65,0,0.35,1)]"
                style={{
                    transform: 'translateX(-100%)',
                    animation: 'spin 1s linear infinite',
                }}
            />
        </Progress.Root>
    );
}
