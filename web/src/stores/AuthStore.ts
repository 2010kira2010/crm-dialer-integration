import { makeAutoObservable, runInAction } from 'mobx';
import { RootStore } from './RootStore';
import api from '../services/api';

interface User {
    id: string;
    email: string;
    name?: string;
    role?: string;
}

export class AuthStore {
    rootStore: RootStore;
    user: User | null = null;
    token: string | null = null;
    isAuthenticated = false;
    isLoading = false;
    isInitialized = false;

    constructor(rootStore: RootStore) {
        this.rootStore = rootStore;
        makeAutoObservable(this);
        this.initializeAuth();
    }

    async initializeAuth() {
        const token = localStorage.getItem('auth_token');
        if (token) {
            this.token = token;
            api.defaults.headers.common['Authorization'] = `Bearer ${token}`;

            try {
                // Validate token by fetching user info
                const response = await api.get('/api/v1/auth/me');
                runInAction(() => {
                    this.user = response.data;
                    this.isAuthenticated = true;
                    this.isInitialized = true;
                });
            } catch (error) {
                // Token is invalid
                this.logout();
                runInAction(() => {
                    this.isInitialized = true;
                });
            }
        } else {
            runInAction(() => {
                this.isInitialized = true;
            });
        }
    }

    async login(email: string, password: string): Promise<boolean> {
        this.isLoading = true;
        try {
            const response = await api.post('/api/v1/auth/login', { email, password });
            const { token, user } = response.data;

            runInAction(() => {
                this.token = token;
                this.user = user;
                this.isAuthenticated = true;
                this.isLoading = false;
            });

            localStorage.setItem('auth_token', token);
            api.defaults.headers.common['Authorization'] = `Bearer ${token}`;

            return true;
        } catch (error: any) {
            runInAction(() => {
                this.isLoading = false;
            });

            if (error.response?.status === 401) {
                return false;
            }

            throw error;
        }
    }

    logout() {
        this.token = null;
        this.user = null;
        this.isAuthenticated = false;

        localStorage.removeItem('auth_token');
        delete api.defaults.headers.common['Authorization'];

        // Clear all stores data
        this.rootStore.flowStore.clear();
        this.rootStore.dataStore.clear();
    }

    async refreshToken(): Promise<boolean> {
        try {
            const response = await api.post('/api/v1/auth/refresh');
            const { token } = response.data;

            runInAction(() => {
                this.token = token;
            });

            localStorage.setItem('auth_token', token);
            api.defaults.headers.common['Authorization'] = `Bearer ${token}`;

            return true;
        } catch (error) {
            this.logout();
            return false;
        }
    }
}