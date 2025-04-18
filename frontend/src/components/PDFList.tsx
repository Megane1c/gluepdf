import { useMemo } from 'react';
import { useSortable, SortableContext, horizontalListSortingStrategy } from '@dnd-kit/sortable';
import { CSS } from '@dnd-kit/utilities';
import { FiFile, FiTrash } from 'react-icons/fi';
import { UploadedFile } from '../services/api';

interface PDFItemProps {
    file: UploadedFile;
    onRemove: (filename: string) => void;
    disabled?: boolean;
}

function truncateFilename(filename: string, maxLength = 20): { name: string; ext: string } {
    const lastDot = filename.lastIndexOf('.');
    if (lastDot === -1) return { name: filename, ext: '' };

    const name = filename.substring(0, lastDot);
    const ext = filename.substring(lastDot);

    if (name.length <= maxLength) return { name, ext };

    return {
        name: name.slice(0, maxLength - 1) + '…',
        ext,
    };
}

const PDFItem = ({ file, onRemove , disabled }: PDFItemProps) => {
    const {
        attributes,
        listeners,
        setNodeRef,
        transform,
        transition,
        isDragging
    } = useSortable({
        id: file.filename,
        resizeObserverConfig: undefined,
        disabled: disabled,
    });

    const style = {
        transform: CSS.Transform.toString(transform),
        transition,
        opacity: isDragging ? 0.5 : 1,
        width: '180px',
        minWidth: '180px',
        display: 'inline-block',
        margin: '8px',
        verticalAlign: 'top',
        backgroundColor: '#1a1a1a',
        borderRadius: '10px',
        padding: '12px',
        gap: '10px'
    };

    const formatFileSize = (bytes: number) => {
        if (bytes < 1024) return `${bytes} B`;
        const kb = bytes / 1024;
        if (kb < 1024) return `${kb.toFixed(1)} KB`;
        const mb = kb / 1024;
        return `${mb.toFixed(1)} MB`;
    };

    // Extract the original filename by removing the UUID prefix
    const displayName = (() => {
        const dashIndex = file.filename.split('-').length > 4 
            ? file.filename.split('-', 5).join('-').length
            : -1;
        
        return dashIndex > 0 
            ? file.filename.slice(dashIndex + 1) 
            : file.filename;
    })();

    const { name, ext } = truncateFilename(displayName);
    
    return (
        <div
            ref={setNodeRef}
            style={style}
            {...attributes}
            {...(disabled ? {} : listeners)}
            className={`pdf-item ${disabled ? 'pdf-item-disabled' : ''}`}
        >
            <button
                onClick={() => onRemove(file.filename)}
                title="Remove file"
                disabled={disabled}
                className={`${disabled ? 'button-disabled' : ''}`}
            >
                <FiTrash size={16} className="remove-button" />
            </button>
            <div className="file-content">
                <FiFile size={20} />
                <div className="file-name" title={displayName}>
                    <span>{name}</span>
                    <span className="file-ext">{ext}</span>
                </div>
                <div className="file-size">
                    {formatFileSize(file.size)}
                </div>
            </div>
        </div>
    );
};

interface PDFListProps {
    files: UploadedFile[];
    onReorder: (files: UploadedFile[]) => void;
    onRemove: (filename: string) => void;
    disabled?: boolean;
}

const PDFList = ({ files, onRemove , disabled}: PDFListProps) => {
    const items = useMemo(() => files.map(file => file.filename), [files]);

    return (
        <SortableContext items={items} strategy={horizontalListSortingStrategy}>
            {files.map((file) => (
                <PDFItem
                    key={file.filename}
                    file={file}
                    onRemove={onRemove}
                    disabled={disabled}
                />
            ))}
        </SortableContext>
    );
};

export default PDFList;