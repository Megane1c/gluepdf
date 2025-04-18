import { useState, useEffect, useCallback } from 'react';
import { useDropzone } from 'react-dropzone';
import { DndContext, DragEndEvent, closestCenter, MouseSensor, TouchSensor, useSensor, useSensors } from '@dnd-kit/core';
import { arrayMove } from '@dnd-kit/sortable';
import { FiUpload, FiTrash2, FiDownload } from 'react-icons/fi';
import api, { UploadedFile } from '../services/api';
import PDFList from './PDFList';

const PDFMerger = () => {
    const [sessionId, setSessionId] = useState<string>('');
    const [files, setFiles] = useState<UploadedFile[]>([]);
    const [loading, setLoading] = useState(false);
    const [downloadUrl, setDownloadUrl] = useState('');
    const [error, setError] = useState<string | null>(null);
    const [isDragActive, setIsDragActive] = useState(false);
    const [isDropzoneDisabled, setIsDropzoneDisabled] = useState(false);

    const sensors = useSensors(
        useSensor(MouseSensor),
        useSensor(TouchSensor)
    );

    useEffect(() => {
        const createSession = async () => {
            try {
                const id = await api.createSession();
                setSessionId(id);
                window.history.pushState({}, '', `/${id}`);
            } catch (err) {
                setError('Failed to initialize session');
            }
        };
        createSession();
    }, []);

    useEffect(() => {
        if (error) {
            const timer = setTimeout(() => {
                setError(null);
            }, 5000); // auto-clear after 5 seconds
    
            return () => clearTimeout(timer); // cleanup if component unmounts early
        }
    }, [error]);

    const onDrop = useCallback(async (acceptedFiles: File[]) => {
        setLoading(true);
        setError('');
        try {
            const newFiles = await Promise.all(
                acceptedFiles.map(file => api.uploadFile(sessionId, file))
            );
            setFiles(prev => [...prev, ...newFiles]);
        } catch (err) {
            setError('Failed to upload one or more files');
        }
        setLoading(false);
    }, [sessionId]);

    const { getRootProps, getInputProps } = useDropzone({
        onDrop,
        accept: { 'application/pdf': ['.pdf'] },
        noClick: files.length > 0,
        disabled: isDropzoneDisabled,
        onDragEnter: () => setIsDragActive(true),
        onDragLeave: () => setIsDragActive(false),
        onDropAccepted: () => setIsDragActive(false),
        onDropRejected: () => setIsDragActive(false),
    });

    const handleRemoveFile = (filename: string) => {
        setFiles(prev => prev.filter(file => file.filename !== filename));
    };

    const handleReorder = async (newFiles: UploadedFile[]) => {
        setFiles(newFiles);
        try {
            await api.updateOrder(sessionId, newFiles.map(f => f.filename));
        } catch (err) {
            setError('Failed to update file order');
        }
    };

    const handleDragEnd = (event: DragEndEvent) => {
        const { active, over } = event;
        
        if (over && active.id !== over.id) {
            const oldIndex = files.findIndex(file => file.filename === active.id);
            const newIndex = files.findIndex(file => file.filename === over.id);
            
            if (oldIndex !== -1 && newIndex !== -1) {
                const newFiles = arrayMove(files, oldIndex, newIndex);
                handleReorder(newFiles);
            }
        }
    };

    const handleMerge = async () => {
        if (files.length < 2) {
            setError('Please upload at least 2 files to merge');
            return;
        }
        setLoading(true);
        setError('');
        try {
            const response = await api.mergeFiles(sessionId);
            setDownloadUrl(response.downloadUrl);
            setIsDropzoneDisabled(true);
        } catch (err) {
            setError('Failed to merge files');
        }
        setLoading(false);
    };

    const handleDownload = () => {
        if (downloadUrl) {
            api.downloadFile(downloadUrl);
        }
    };

    return (
        <div className="layout-container">
            {error && (
                <div className="error-message">
                    <div className="error-message__content">
                        <span>{error}</span>
                    </div>
                </div>
            )}
    
            <div className="pdf-merger-container">
                <div className="pdf-box">
                    <div className="header-section">
                        <div className="header-content">
                            <h1 className="title">Glue PDF</h1>
                            {files.length === 0 || loading || !isDropzoneDisabled && (
                                <button
                                    onClick={() => setFiles([])}
                                    id="clear-button"
                                >
                                    <FiTrash2 className="remove-button" />
                                    Clear All
                                </button>
                            )}
                        </div>
                    </div>
    
                    <div {...getRootProps()} 
                        className={`dropzone-section ${isDragActive ? 'drag-active' : ''} ${isDropzoneDisabled ? 'dropzone-disabled' : ''}`}>
                        <DndContext
                            sensors={sensors}
                            collisionDetection={closestCenter}
                            onDragEnd={handleDragEnd}
                        >
                            <div
                                className={`dropzone ${files.length > 0 ? 'compact' : 'spacious'} ${isDropzoneDisabled ? 'dropzone-disabled' : ''}`}
                            >
                                <input {...getInputProps()} />
                                {files.length > 0 ? (
                                    <div className="dropzone-content">
                                        <div className="file-list-container">
                                            <PDFList 
                                                files={files} 
                                                onReorder={handleReorder}
                                                onRemove={handleRemoveFile}
                                                disabled={isDropzoneDisabled}
                                            />
                                        </div>
                                        <p className="dropzone-hint">
                                            You can drag more files here or rearrange the order
                                        </p>
                                    </div>
                                ) : (
                                    <div className="dropzone-empty">
                                        <FiUpload className="upload-icon" />
                                        <p className="upload-text">Drag and drop PDFs here, or click to browse</p>
                                    </div>
                                )}
                            </div>
                        </DndContext>
                    </div>
    
                    {files.length > 0 && (
                        <div className="action-section">
                            <div className="action-buttons">
                                {downloadUrl ? (
                                    <button
                                        onClick={handleDownload}
                                        className="download-button"
                                    >
                                        <FiDownload className="button-icon" />
                                        Download Merged PDF
                                    </button>
                                ) : (
                                    <button
                                        onClick={handleMerge}
                                        disabled={files.length < 2 || loading}
                                        className="merge-button"
                                    >
                                        {loading ? 'Processing...' : 'Merge PDFs'}
                                    </button>
                                )}
                            </div>
                        </div>
                    )}
                </div>
            </div>
        </div>
    );
    
};

export default PDFMerger;