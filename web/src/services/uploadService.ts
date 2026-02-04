import * as tus from 'tus-js-client';
import { useAuthStore } from '@/store/useAuthStore';

export interface UploadOptions {
    onProgress?: (bytesUploaded: number, bytesTotal: number) => void;
    onSuccess?: () => void;
    onError?: (error: Error) => void;
}

export const uploadService = {
    uploadFile: (file: File, options: UploadOptions) => {
        // Get token from store (non-reactive way using getState)
        const token = useAuthStore.getState().token;

        const upload = new tus.Upload(file, {
            endpoint: `${import.meta.env.VITE_API_URL || '/api'}/tus/upload`,
            retryDelays: [0, 3000, 5000, 10000, 20000],
            headers: {
                Authorization: `Bearer ${token}`,
            },
            metadata: {
                filename: file.name,
                filetype: file.type,
            },
            onError: (error) => {
                console.error("Upload failed:", error);
                if (options.onError) options.onError(error);
            },
            onProgress: (bytesUploaded, bytesTotal) => {
                if (options.onProgress) options.onProgress(bytesUploaded, bytesTotal);
            },
            onSuccess: () => {
                if (options.onSuccess) options.onSuccess();
            },
        });

        // Check if there are any previous uploads to continue.
        upload.findPreviousUploads().then((previousUploads) => {
            // Ask the user using UI if they want to resume - but for now, let's just resume if found (Basic PoC)
            // Or we can start fresh for simplicity in this MVP
            if (previousUploads.length) {
                upload.resumeFromPreviousUpload(previousUploads[0]);
            }

            upload.start();
        });

        return upload;
    }
};
