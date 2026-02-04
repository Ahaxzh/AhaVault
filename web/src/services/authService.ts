import { api } from './api';

export interface User {
    user_id: string;
    email: string;
    role: string;
}

export interface RegisterRequest {
    email: string;
    password?: string; // Optional if using magic link, but we are using password
    invite_code?: string;
}

export interface LoginRequest {
    email: string;
    password?: string;
    captcha_token?: string;
}

export interface AuthResponse {
    code: number;
    message: string;
    data: {
        user_id: string;
        email: string;
        role?: string;
        token: string;
        expires_in: number;
    };
}

export const authService = {
    register: (data: RegisterRequest) => {
        return api.post<any, AuthResponse>('/auth/register', data);
    },

    login: (data: LoginRequest) => {
        return api.post<any, AuthResponse>('/auth/login', data);
    },

    logout: () => {
        return api.post('/auth/logout');
    },

    getCurrentUser: () => {
        return api.get<any, { data: User }>('/user/me');
    }
};
