import { useState, useRef } from 'react';
import { Button } from '@/components/ui/Button';
import { uploadService } from '@/services/uploadService';
import { Loader2, Plus, UploadCloud, X } from 'lucide-react';
import * as tus from 'tus-js-client';

interface UploadButtonProps {
    onUploadComplete?: () => void;
}

export function UploadButton({ onUploadComplete }: UploadButtonProps) {
    const [uploading, setUploading] = useState(false);
    const [progress, setProgress] = useState(0);
    const fileInputRef = useRef<HTMLInputElement>(null);
    const uploadRef = useRef<tus.Upload | null>(null);

    const handleFileSelect = (e: React.ChangeEvent<HTMLInputElement>) => {
        const file = e.target.files?.[0];
        if (!file) return;

        startUpload(file);
        // Reset input so same file can be selected again
        e.target.value = '';
    };

    const startUpload = (file: File) => {
        setUploading(true);
        setProgress(0);

        const upload = uploadService.uploadFile(file, {
            onProgress: (bytesUploaded, bytesTotal) => {
                const percentage = (bytesUploaded / bytesTotal) * 100;
                setProgress(percentage);
            },
            onSuccess: () => {
                setUploading(false);
                setProgress(100);
                if (onUploadComplete) onUploadComplete();
            },
            onError: (error) => {
                setUploading(false);
                alert(`Upload failed: ${error.message}`);
            }
        });

        uploadRef.current = upload;
    };

    const cancelUpload = () => {
        if (uploadRef.current) {
            uploadRef.current.abort();
            setUploading(false);
            uploadRef.current = null;
        }
    }

    return (
        <div className="flex items-center gap-2">
            <input
                type="file"
                ref={fileInputRef}
                className="hidden"
                onChange={handleFileSelect}
            />

            {uploading ? (
                <div className="flex items-center gap-2 bg-secondary/50 px-3 py-2 rounded-md animate-in fade-in slide-in-from-right-4">
                    <div className="flex flex-col w-32 gap-1">
                        <div className="flex justify-between text-xs text-muted-foreground">
                            <span>Uploading...</span>
                            <span>{Math.round(progress)}%</span>
                        </div>
                        <div className="h-1.5 w-full bg-primary/20 rounded-full overflow-hidden">
                            <div
                                className="h-full bg-primary transition-all duration-300 ease-out"
                                style={{ width: `${progress}%` }}
                            />
                        </div>
                    </div>
                    <Button size="icon" variant="ghost" className="h-6 w-6" onClick={cancelUpload}>
                        <X className="w-3 h-3" />
                    </Button>
                </div>
            ) : (
                <Button onClick={() => fileInputRef.current?.click()}>
                    <Plus className="mr-2 h-4 w-4" /> Upload
                </Button>
            )}
        </div>
    );
}
