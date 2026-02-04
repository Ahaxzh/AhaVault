import axios from 'axios';

// Create basic axios instance
export const api = axios.create({
    baseURL: import.meta.env.VITE_API_URL || '/api',
    timeout: 10000,
    headers: {
        'Content-Type': 'application/json',
    },
});

// Response interceptor for error handling
api.interceptors.response.use(
    (response) => response.data,
    (error) => {
        // Handle standard API error format
        const apiError = error.response?.data;

        // Create a normalized error object
        const normalizedError = new Error(apiError?.message || error.message || 'Unknown error');
        (normalizedError as any).code = apiError?.code;
        (normalizedError as any).status = error.response?.status;
        (normalizedError as any).data = apiError?.data;

        return Promise.reject(normalizedError);
    }
);
