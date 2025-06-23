import React, { useState } from 'react';
import { useNavigate, Navigate } from 'react-router-dom';
import {
    Box,
    Paper,
    TextField,
    Button,
    Typography,
    Alert,
    CircularProgress,
    Link,
    InputAdornment,
    IconButton,
} from '@mui/material';
import {
    Visibility,
    VisibilityOff,
    Login as LoginIcon,
} from '@mui/icons-material';
import { observer } from 'mobx-react-lite';
import { useStores } from '../../hooks/useStores';

export const LoginPage: React.FC = observer(() => {
    const navigate = useNavigate();
    const { authStore } = useStores();
    const [email, setEmail] = useState('');
    const [password, setPassword] = useState('');
    const [showPassword, setShowPassword] = useState(false);
    const [error, setError] = useState('');

    // If already authenticated, redirect to dashboard
    if (authStore.isAuthenticated) {
        return <Navigate to="/" replace />;
    }

    const handleSubmit = async (e: React.FormEvent) => {
        e.preventDefault();
        setError('');

        if (!email || !password) {
            setError('Пожалуйста, заполните все поля');
            return;
        }

        try {
            const success = await authStore.login(email, password);
            if (success) {
                navigate('/');
            } else {
                setError('Неверный email или пароль');
            }
        } catch (err) {
            setError('Произошла ошибка при входе в систему');
        }
    };

    return (
        <Box
            sx={{
                minHeight: '100vh',
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'center',
                bgcolor: 'background.default',
                backgroundImage: 'linear-gradient(135deg, #667eea 0%, #764ba2 100%)',
            }}
        >
            <Paper
                elevation={3}
                sx={{
                    p: 4,
                    width: '100%',
                    maxWidth: 400,
                    borderRadius: 6,
                }}
            >
                <Box sx={{ textAlign: 'center', mb: 3 }}>
                    <Typography variant="h4" gutterBottom fontWeight="bold">
                        CRM-Dialer Integration
                    </Typography>
                    <Typography variant="body2" color="text.secondary">
                        Войдите в систему для продолжения
                    </Typography>
                </Box>

                <form onSubmit={handleSubmit}>
                    {error && (
                        <Alert severity="error" sx={{ mb: 2 }} onClose={() => setError('')}>
                            {error}
                        </Alert>
                    )}

                    <TextField
                        fullWidth
                        label="Email"
                        type="email"
                        value={email}
                        onChange={(e) => setEmail(e.target.value)}
                        margin="normal"
                        required
                        autoComplete="email"
                        autoFocus
                    />

                    <TextField
                        fullWidth
                        label="Пароль"
                        type={showPassword ? 'text' : 'password'}
                        value={password}
                        onChange={(e) => setPassword(e.target.value)}
                        margin="normal"
                        required
                        autoComplete="current-password"
                        InputProps={{
                            endAdornment: (
                                <InputAdornment position="end">
                                    <IconButton
                                        onClick={() => setShowPassword(!showPassword)}
                                        edge="end"
                                    >
                                        {showPassword ? <VisibilityOff /> : <Visibility />}
                                    </IconButton>
                                </InputAdornment>
                            ),
                        }}
                    />

                    <Button
                        type="submit"
                        fullWidth
                        variant="contained"
                        size="large"
                        sx={{ mt: 3, mb: 2 }}
                        disabled={authStore.isLoading}
                        startIcon={authStore.isLoading ? <CircularProgress size={20} /> : <LoginIcon />}
                    >
                        {authStore.isLoading ? 'Вход...' : 'Войти'}
                    </Button>

                    <Box sx={{ textAlign: 'center', mb: 2 }}>
                        <Link
                            component="button"
                            variant="body2"
                            onClick={(e) => {
                                e.preventDefault();
                                alert('Обратитесь к администратору для восстановления пароля');
                            }}
                        >
                            Забыли пароль?
                        </Link>
                    </Box>


                </form>

                <Box sx={{ mt: 3, textAlign: 'center' }}>
                    <Typography variant="caption" color="text.secondary">
                        © 2025 Все права защищены.
                    </Typography>
                </Box>
            </Paper>
        </Box>
    );
});
