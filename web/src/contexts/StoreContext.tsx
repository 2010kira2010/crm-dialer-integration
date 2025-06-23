import React, { createContext, ReactNode } from 'react';
import { RootStore } from '../stores/RootStore';

export const StoreContext = createContext<RootStore | null>(null);

interface StoreProviderProps {
    children: ReactNode;
}

export const StoreProvider: React.FC<StoreProviderProps> = ({ children }) => {
    const rootStore = new RootStore();

    return (
        <StoreContext.Provider value={rootStore}>
            {children}
        </StoreContext.Provider>
    );
};