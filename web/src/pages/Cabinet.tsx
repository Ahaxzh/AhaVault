import { useEffect, useState } from 'react'
import { DashboardLayout } from '@/components/layout/DashboardLayout'
import { Card, CardContent } from '@/components/ui/Card'
import { Button } from '@/components/ui/Button'
import { Input } from '@/components/ui/Input'
import { fileService, type FileItem } from '@/services/fileService'
import { Loader2, Search, Trash2, File as FileIcon, Download, Share2 } from 'lucide-react'
import { UploadButton } from '@/components/upload/UploadButton'
import { CreateShareModal } from '@/components/share/CreateShareModal'

export default function Cabinet() {
    const [files, setFiles] = useState<FileItem[]>([])
    const [loading, setLoading] = useState(true)
    const [search, setSearch] = useState('')
    const [shareFileId, setShareFileId] = useState<string | null>(null)

    const fetchFiles = async () => {
        setLoading(true)
        try {
            const res = await fileService.listFiles(1, 100, search)
            setFiles(res.data.items || [])
        } catch (error) {
            console.error("Failed to fetch files", error)
        } finally {
            setLoading(false)
        }
    }

    useEffect(() => {
        fetchFiles()
    }, [search])

    const handleDelete = async (id: string) => {
        if (!confirm("Are you sure you want to delete this file?")) return;
        try {
            await fileService.deleteFile(id)
            fetchFiles() // Refresh
        } catch (error) {
            console.error("Delete failed", error)
        }
    }

    const formatSize = (bytes: number) => {
        if (bytes === 0) return '0 B'
        const k = 1024
        const sizes = ['B', 'KB', 'MB', 'GB', 'TB']
        const i = Math.floor(Math.log(bytes) / Math.log(k))
        return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i]
    }

    return (
        <DashboardLayout>
            <div className="space-y-6">

                {/* Header Actions */}
                <div className="flex flex-col md:flex-row gap-4 items-center justify-between">
                    <div>
                        <h1 className="text-3xl font-bold tracking-tight">My Cabinet</h1>
                        <p className="text-muted-foreground">Manage your secure files</p>
                    </div>
                    <div className="flex items-center gap-2 w-full md:w-auto">
                        <div className="relative flex-1 md:w-64">
                            <Search className="absolute left-2.5 top-2.5 h-4 w-4 text-muted-foreground" />
                            <Input
                                placeholder="Search files..."
                                className="pl-9"
                                value={search}
                                onChange={(e) => setSearch(e.target.value)}
                            />
                        </div>
                        <UploadButton onUploadComplete={fetchFiles} />
                    </div>
                </div>

                {/* File List */}
                <Card className="bg-surface/50 backdrop-blur-sm border-primary/10">
                    <CardContent className="p-0">
                        {loading ? (
                            <div className="flex justify-center py-12">
                                <Loader2 className="h-8 w-8 animate-spin text-primary" />
                            </div>
                        ) : files.length === 0 ? (
                            <div className="flex flex-col items-center justify-center py-12 text-center text-muted-foreground">
                                <FileIcon className="h-12 w-12 mb-4 opacity-50" />
                                <p>No files found</p>
                                <div className="mt-4">
                                    <UploadButton onUploadComplete={fetchFiles} />
                                </div>
                            </div>
                        ) : (
                            <div className="divide-y divide-border/50">
                                {files.map((file) => (
                                    <div key={file.id} className="flex items-center justify-between p-4 hover:bg-primary/5 transition-colors group">
                                        <div className="flex items-center gap-4 overflow-hidden">
                                            <div className="h-10 w-10 rounded bg-primary/10 flex items-center justify-center text-primary">
                                                <FileIcon className="h-5 w-5" />
                                            </div>
                                            <div className="truncate">
                                                <p className="font-medium truncate">{file.filename}</p>
                                                <div className="flex items-center gap-2 text-xs text-muted-foreground">
                                                    <span>{formatSize(file.size)}</span>
                                                    <span>•</span>
                                                    <span>{new Date(file.created_at).toLocaleDateString()}</span>
                                                </div>
                                            </div>
                                        </div>

                                        <div className="flex items-center gap-1 opacity-100 md:opacity-0 group-hover:opacity-100 transition-opacity">
                                            <Button size="icon" variant="ghost" title="Download" onClick={() => window.open(`${import.meta.env.VITE_API_URL || '/api'}/files/${file.id}/download`)}>
                                                <Download className="h-4 w-4" />
                                            </Button>
                                            <Button size="icon" variant="ghost" title="Share" onClick={() => setShareFileId(file.id)}>
                                                <Share2 className="h-4 w-4" />
                                            </Button>
                                            <Button size="icon" variant="ghost" title="Delete" className="text-destructive hover:text-destructive hover:bg-destructive/10" onClick={() => handleDelete(file.id)}>
                                                <Trash2 className="h-4 w-4" />
                                            </Button>
                                        </div>
                                    </div>
                                ))}
                            </div>
                        )}
                    </CardContent>
                </Card>
            </div>

            {/* Share Modal */}
            {shareFileId && (
                <CreateShareModal
                    fileIds={[shareFileId]}
                    onClose={() => setShareFileId(null)}
                />
            )}
        </DashboardLayout>
    )
}
