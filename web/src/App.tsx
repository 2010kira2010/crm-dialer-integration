import React from 'react';
import { BrowserRouter as Router, Routes, Route, Navigate } from 'react-router-dom';
import { ThemeProvider, createTheme } from '@mui/material/styles';
import CssBaseline from '@mui/material/CssBaseline';
import { StoreProvider } from './contexts/StoreContext';
import { Layout } from './components/Layout/Layout';
import { PrivateRoute } from './components/PrivateRoute/PrivateRoute';
import { LoginPage } from './pages/LoginPage/LoginPage';
import { Dashboard } from './pages/Dashboard/Dashboard';
import { FlowsPage } from './pages/FlowsPage/FlowsPage';
import { FlowEditorPage } from './pages/FlowEditorPage/FlowEditorPage';
import { SettingsPage } from './pages/SettingsPage/SettingsPage';

const theme = createTheme({
    palette: {
        mode: 'light',
        primary: {
            main: '#1976d2',
        },
        secondary: {
            main: '#dc004e',
        },
    },
});

function App() {
    return (
        <StoreProvider>
            <ThemeProvider theme={theme}>
                <CssBaseline />
                <Router>
                    <Routes>
                        {/* Public routes */}
                        <Route path="/login" element={<LoginPage />} />

                        {/* Private routes */}
                        <Route
                            path="/"
                            element={
                                <PrivateRoute>
                                    <Layout />
                                </PrivateRoute>
                            }
                        >
                            <Route index element={<Dashboard />} />
                            <Route path="flows" element={<FlowsPage />} />
                            <Route path="flows/:id" element={<FlowEditorPage />} />
                            <Route path="settings" element={<SettingsPage />} />
                        </Route>

                        {/* Catch all redirect */}
                        <Route path="*" element={<Navigate to="/" replace />} />
                    </Routes>
                </Router>
            </ThemeProvider>
        </StoreProvider>
    );
}

export default App;