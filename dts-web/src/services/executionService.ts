import axios from 'axios';
import { Execution } from '../types';

const API_BASE_URL = process.env.REACT_APP_API_BASE_URL || 'http://localhost:8082';

export const getExecutionHistory = async (jobId: string): Promise<Execution[]> => {
    try {
        const response = await axios.get(`${API_BASE_URL}/v1/executions?job_id=${jobId}`);
        return response.data.executions;
    } catch (error) {
        console.error('Error fetching execution history:', error);
        throw error;
    }
};
