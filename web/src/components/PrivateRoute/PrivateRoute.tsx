import React from 'react';
import { Navigate, useLocation } from 'react-router-dom';
import { observer } from 'mobx-react-lite';
import { useStores } from '../../hooks/useStores';
import { Box, CircularProgress } from '@mui/material';

interface PrivateRouteProps {
    children: React.ReactElement;
}

export const PrivateRoute: React.FC<PrivateRouteProps> = observer(({ children }) => {
    const { authStore } = useStores();
    const location = useLocation();

    // Check if we're still loading auth state
    if (authStore.isLoading) {
        return (
            <Box
                sx={{
                    display: 'flex',
                    justifyContent: 'center',
                    alignItems: 'center',
                    minHeight: '100vh',
                }}
            >
                <CircularProgress />
            </Box>
        );
    }

    if (!authStore.isAuthenticated) {
        // Redirect to login page but save the attempted location
        return <Navigate to="/login" state={{ from: location }} replace />;
    }

    return children;
});