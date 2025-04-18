import axios from 'axios';

const API_BASE_URL = 'http://localhost:8080';

export interface UploadedFile {
    filename: string;
    size: number;
}

export interface MergeResponse {
    downloadUrl: string;
}

const api = {
    createSession: async (): Promise<string> => {
        const response = await axios.post(`${API_BASE_URL}/api/sessions/`);
        return response.data.sessionId;
    },

    uploadFile: async (sessionId: string, file: File): Promise<UploadedFile> => {
        const formData = new FormData();
        formData.append('pdf', file);
        const response = await axios.post(`${API_BASE_URL}/api/sessions/${sessionId}/files`, formData);
        return response.data;
    },

    updateOrder: async (sessionId: string, files: string[]): Promise<boolean> => {
        const response = await axios.put(`${API_BASE_URL}/api/sessions/${sessionId}/order`, { files });
        return response.data.success;
    },

    mergeFiles: async (sessionId: string): Promise<MergeResponse> => {
        const response = await axios.post(`${API_BASE_URL}/api/sessions/${sessionId}/actions/merge`);
        return response.data;
    },

    downloadFile: (url: string) => {
        window.location.href = `${API_BASE_URL}${url}`;
        setTimeout(() => {
            window.location.reload();
        }, 500); 
    }
};

export default api;