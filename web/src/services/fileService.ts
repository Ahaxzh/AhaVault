import { api } from './api';

export interface FileItem {
    id: string;
    filename: string;
    size: number;
    mime_type: string;
    hash: string;
    created_at: string;
    is_shared: boolean;
    share_count: number;
}

export interface ListFilesResponse {
    code: number;
    message: string;
    data: {
        items: FileItem[];
        total: number;
        page: number;
        page_size: number;
    };
}

export const fileService = {
    listFiles: (page = 1, pageSize = 20, search?: string) => {
        return api.get<any, ListFilesResponse>('/files', {
            params: { page, page_size: pageSize, search }
        });
    },

    deleteFile: (fileId: string) => {
        return api.delete(`/files/${fileId}`);
    },

    // Note: Upload logic usually involves Tus or specialized handling, 
    // keeping it simple or TBD for now.
    // We will need a specific upload service later for Tus.
};
