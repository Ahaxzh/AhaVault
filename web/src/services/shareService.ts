import { api } from './api';

export interface FileInfo {
    file_id: string;
    filename: string;
    size: number;
    mime_type: string;
}

export interface ShareInfo {
    share_id: string;
    files: FileInfo[];
    expires_at: string;
    remaining_downloads: number;
    requires_password?: boolean;
    requires_captcha?: boolean;
}

export interface GetShareResponse {
    code: number;
    message: string;
    data: ShareInfo;
}

export interface CreateShareRequest {
    file_ids: string[];
    expires_in: number; // in seconds
    max_downloads: number;
    password?: string;
}

export interface CreateShareResponse {
    code: number;
    message: string;
    data: {
        share_id: string;
        pickup_code: string;
        expires_at: string;
    };
}

export const shareService = {
    /**
     * Get share info by pickup code
     * @param code 8-digit pickup code
     * @param password Optional password
     * @param captchaToken Optional captcha token
     */
    getShareByCode: async (code: string, password?: string, captchaToken?: string) => {
        // Note: Backend route is POST /api/public/shares/:code
        // Based on API.md and routes.go
        return api.post<any, GetShareResponse>(`/public/shares/${code}`, {
            password,
            captcha_token: captchaToken
        });
    },

    /**
     * Create a new share
     */
    createShare: async (data: CreateShareRequest) => {
        return api.post<any, CreateShareResponse>('/shares', data);
    }
};
